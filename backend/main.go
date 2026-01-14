package main

import (
	"collab-doc-backend/models"
	"collab-doc-backend/ws"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DocumentDetail struct {
	models.Document
	CreatorName string `json:"creator_name"`
	FolderName  string `json:"folder_name"`
}

func main() {
	// 配置数据库 DSN
	dsn := "root:123456@tcp(127.0.0.1:3306)/collab_doc?charset=utf8mb4&parseTime=True&loc=Local"
	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("数据库连接失败: %v\n", err)
		return
	}
	fmt.Println("数据库连接成功")
	// 启动 Gin 框架
	r := gin.Default()
	r.Static("/uploads", "./uploads")
	// 允许跨域（解决前后端分离运行时的报错）
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 初始化 Hub
	hub := &ws.Hub{
		Clients: make(map[string]map[*websocket.Conn]bool),
		DB:      db, // gorm.DB
	}
	// 增加 WebSocket 路由
	r.GET("/ws/:id", func(c *gin.Context) {
		docID := c.Param("id")
		hub.HandleWS(c.Writer, c.Request, docID)
	})

	// 注册接口
	r.POST("/register", func(c *gin.Context) {
		var newUser models.User
		if err := c.ShouldBindJSON(&newUser); err != nil {
			c.JSON(400, gin.H{"error": "参数错误"})
			return
		}
		if newUser.Email == "" && newUser.Phone == "" {
			c.JSON(400, gin.H{"error": "邮箱或手机号必须填写一项"})
			return
		}
		//  GORM 自动把 newUser 里的 email, phone 存入数据库
		if newUser.Role == "" {
			newUser.Role = "editor"
		}
		if err := db.Create(&newUser).Error; err != nil {
			c.JSON(500, gin.H{"error": "注册失败，账号可能已存在"})
			return
		}
		c.JSON(200, gin.H{"message": "注册成功"})
	})

	// 头像上传接口
	r.POST("/user/avatar", func(c *gin.Context) {
		userID := c.PostForm("user_id")
		file, err := c.FormFile("avatar") // 获取上传的文件
		if err != nil {
			c.JSON(400, gin.H{"error": "获取文件失败"})
			return
		}

		// 生成唯一文件名 (当前时间戳 + 原文件名)
		fileName := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
		dst := "./uploads/" + fileName

		// 保存文件到本地 uploads 文件夹
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(500, gin.H{"error": "保存文件失败"})
			return
		}

		// 将文件路径更新到数据库
		avatarURL := "http://localhost:8080/uploads/" + fileName
		db.Model(&models.User{}).Where("id = ?", userID).Update("avatar", avatarURL)

		c.JSON(200, gin.H{
			"message": "头像上传成功",
			"url":     avatarURL,
		})
	})

	// 1.1 用户登录
	r.POST("/login", func(c *gin.Context) {
		var loginInfo struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		// 解析前端发来的 JSON
		if err := c.ShouldBindJSON(&loginInfo); err != nil {
			c.JSON(400, gin.H{"error": "请求参数错误"})
			return
		}

		var user models.User
		// 在数据库中查找用户名和密码是否匹配
		// 对应 SQL: SELECT * FROM users WHERE username = 'xxx' AND password = 'xxx' LIMIT 1
		err := db.Where("username = ? AND password = ?", loginInfo.Username, loginInfo.Password).First(&user).Error

		if err != nil {
			c.JSON(401, gin.H{"error": "用户名或密码错误"})
			return
		}

		c.JSON(200, gin.H{
			"message": "登录成功",
			"user":    user,
		})
	})

	// 1.2 更新用户信息
	r.PUT("/user/profile", func(c *gin.Context) {
		var req models.User
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数错误"})
			return
		}
		// 根据 ID 更新非空字段
		db.Model(&models.User{}).Where("id = ?", req.ID).Updates(req)
		c.JSON(200, gin.H{"message": "资料更新成功"})
	})

	// 2.1 文档管理：创建文档
	r.POST("/documents", func(c *gin.Context) {
		var newDoc models.Document
		if err := c.ShouldBindJSON(&newDoc); err != nil {
			c.JSON(400, gin.H{"error": "参数格式错误"})
			return
		}

		// 执行保存
		if err := db.Create(&newDoc).Error; err != nil {
			c.JSON(500, gin.H{"error": "创建文档失败: " + err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "文档创建成功", "doc_id": newDoc.ID})
	})

	// 2.1 文档管理：获取所有文档列表
	r.GET("/documents", func(c *gin.Context) {
		var docs []models.Document
		db.Find(&docs)
		c.JSON(200, docs)
	})

	// 2.1 更新文档（保存编辑内容）
	r.PUT("/documents/:id", func(c *gin.Context) {
		id := c.Param("id") // 获取 URL 里的文档 ID
		var doc models.Document

		// 验证存在性
		if err := db.First(&doc, id).Error; err != nil {
			c.JSON(404, gin.H{"error": "文档不存在"})
			return
		}

		// 解析前端传来的新标题或新内容
		if err := c.ShouldBindJSON(&doc); err != nil {
			c.JSON(400, gin.H{"error": "无效的更新数据"})
			return
		}

		// 保存修改到数据库
		db.Save(&doc)
		c.JSON(200, gin.H{"message": "文档更新成功", "doc": doc})
	})

	// 2.1 获取单个文档详情
	r.GET("/documents/:id", func(c *gin.Context) {
		id := c.Param("id")
		var doc models.Document
		if err := db.First(&doc, id).Error; err != nil {
			c.JSON(404, gin.H{"error": "未找到该文档"})
			return
		}
		c.JSON(200, doc)
	})

	// 2.1 删除文档
	r.DELETE("/documents/:id", func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Delete(&models.Document{}, id).Error; err != nil {
			c.JSON(500, gin.H{"error": "删除失败"})
			return
		}
		c.JSON(200, gin.H{"message": "文档已成功删除"})
	})

	// 2.2 版本控制：保存当前快照为新版本
	r.POST("/documents/:id/versions", func(c *gin.Context) {
		docID := c.Param("id")

		// 从 documents 表中查出当前最新的内容
		var currentDoc models.Document
		if err := db.First(&currentDoc, docID).Error; err != nil {
			c.JSON(404, gin.H{"error": "未找到原文档"})
			return
		}

		// 创建一个版本记录
		var versionName struct {
			Name string `json:"name"`
		}
		c.ShouldBindJSON(&versionName) // 允许前端传个名字

		newVersion := models.DocVersion{
			DocID:       currentDoc.ID,
			Content:     currentDoc.Content,
			VersionName: versionName.Name,
		}

		// 存入版本表
		if err := db.Create(&newVersion).Error; err != nil {
			c.JSON(500, gin.H{"error": "保存版本失败"})
			return
		}

		c.JSON(200, gin.H{"message": "版本快照已保存", "version_id": newVersion.ID})
	})

	// 2.2 版本控制：获取某篇文档的所有历史版本列表
	r.GET("/documents/:id/versions", func(c *gin.Context) {
		docID := c.Param("id")
		var versions []models.DocVersion

		// 按照创建时间倒序排列，最新的版本在最前面
		db.Where("doc_id = ?", docID).Order("created_at desc").Find(&versions)

		c.JSON(200, versions)
	})

	// 2.3 搜索文档
	r.GET("/search", func(c *gin.Context) {
		keyword := c.Query("keyword")
		creator := c.Query("creator")
		startDate := c.Query("start") // 格式: yyyy-mm-dd
		endDate := c.Query("end")
		sort := c.DefaultQuery("sort", "documents.created_at")
		order := c.DefaultQuery("order", "desc")

		var results []DocumentDetail

		// Joins 连接查询
		query := db.Table("documents").
			Select("documents.*, users.username as creator_name, IFNULL(folders.name, '全部文档') as folder_name").
			Joins("left join users on users.id = documents.owner_id").
			Joins("left join folders on folders.id = documents.folder_id")

		// 动态条件过滤
		if keyword != "" {
			query = query.Where("documents.title LIKE ?", "%"+keyword+"%")
		}
		if creator != "" {
			query = query.Where("users.username LIKE ?", "%"+creator+"%")
		}
		if startDate != "" && endDate != "" {
			query = query.Where("documents.created_at BETWEEN ? AND ?", startDate+" 00:00:00", endDate+" 23:59:59")
		}

		query.Order(sort + " " + order).Scan(&results)
		c.JSON(200, results)
	})

	// 文件夹管理
	r.GET("/folders", func(c *gin.Context) {
		userID := c.Query("user_id")
		var folders []models.Folder
		db.Where("user_id = ?", userID).Find(&folders)
		c.JSON(200, folders)
	})

	r.POST("/folders", func(c *gin.Context) {
		var f models.Folder
		c.ShouldBindJSON(&f)
		db.Create(&f)
		c.JSON(200, f)
	})

	r.DELETE("/folders/:id", func(c *gin.Context) {
		db.Delete(&models.Folder{}, c.Param("id"))
		db.Model(&models.Document{}).Where("folder_id = ?", c.Param("id")).Update("folder_id", 0)
		c.JSON(200, gin.H{"message": "文件夹已删除"})
	})

	// 收藏/取消收藏
	r.PUT("/documents/:id/star", func(c *gin.Context) {
		var body struct {
			IsStarred bool `json:"is_starred"`
		}
		c.ShouldBindJSON(&body)
		db.Model(&models.Document{}).Where("id = ?", c.Param("id")).Update("is_starred", body.IsStarred)
		c.JSON(200, gin.H{"message": "操作成功"})
	})

	// 移动到文件夹
	r.PUT("/documents/:id/move", func(c *gin.Context) {
		var body struct {
			FolderID int `json:"folder_id"`
		}
		c.ShouldBindJSON(&body)
		db.Model(&models.Document{}).Where("id = ?", c.Param("id")).Update("folder_id", body.FolderID)
		c.JSON(200, gin.H{"message": "移动成功"})
	})

	// 获取评论
	r.GET("/documents/:id/comments", func(c *gin.Context) {
		var comments []struct {
			models.Comment
			Username string `json:"username"`
		}
		db.Table("comments").
			Select("comments.*, users.username").
			Joins("left join users on users.id = comments.user_id").
			Where("doc_id = ?", c.Param("id")).
			Order("created_at asc").Scan(&comments)
		c.JSON(200, comments)
	})

	// 发表评论
	r.POST("/comments", func(c *gin.Context) {
		var comment models.Comment
		if err := c.ShouldBindJSON(&comment); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		db.Create(&comment)
		c.JSON(200, comment)
	})

	// 删除评论
	r.DELETE("/comments/:id", func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.Comment{}, id)
		c.JSON(200, gin.H{"message": "批注已删除"})
	})

	// 获取所有用户的测试接口
	r.GET("/users", func(c *gin.Context) {
		var users []models.User
		db.Find(&users)
		c.JSON(200, users)
	})

	// 运行服务器
	r.Run(":8080")
}
