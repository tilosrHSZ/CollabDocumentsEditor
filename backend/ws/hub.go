package ws

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// Message 传输的消息格式
type Message struct {
	Type     string `json:"type"`
	DocID    string `json:"doc_id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Content  string `json:"content"`
}

// Hub 管理所有的在线连接
type Hub struct {
	// 按照 DocID 分组管理连接：map[文档ID]map[连接对象]布尔值
	Clients map[string]map[*websocket.Conn]bool
	Mutex   sync.Mutex // 防止多个协程同时操作 map 导致崩溃
	DB      *gorm.DB   // 数据库连接
}

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 允许跨域
}

func NewHub() *Hub {
	return &Hub{
		Clients: make(map[string]map[*websocket.Conn]bool),
	}
}

// 处理 WebSocket 连接请求
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request, docID string) {
	userID := r.URL.Query().Get("userId")

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 1. 注册连接
	h.Mutex.Lock()
	if h.Clients[docID] == nil {
		h.Clients[docID] = make(map[*websocket.Conn]bool)
	}
	h.Clients[docID][conn] = true
	h.Mutex.Unlock()

	if userID != "" {
		h.DB.Table("operation_logs").Create(map[string]interface{}{
			"user_id": userID,
			"action":  "进入文档",
			"doc_id":  docID,
		})
	}

	// 2. 监听消息
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			// 连接断开
			h.Mutex.Lock()
			delete(h.Clients[docID], conn)
			h.Mutex.Unlock()
			break
		}

		if msg.Type == "edit" {
			// 如果是编辑消息，更新数据库文档内容
			h.DB.Table("documents").Where("id = ?", msg.DocID).Update("content", msg.Content)
		} else if msg.Type == "chat" {
			// 如果是聊天消息，记录到 operation_logs
			h.DB.Table("operation_logs").Create(map[string]interface{}{
				"user_id": msg.UserID,
				"action":  "发送聊天: " + msg.Content,
				"doc_id":  msg.DocID,
			})
		}

		// 3. 广播给同一个文档的其他所有人
		h.Broadcast(msg, conn)
	}
}

// 将消息发送给同一文档房间内的其他用户
func (h *Hub) Broadcast(msg Message, sender *websocket.Conn) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	for client := range h.Clients[msg.DocID] {
		if client != sender { // 不发给发送者本人
			client.WriteJSON(msg)
		}
	}
}
