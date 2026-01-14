<template>
  <div style="padding: 20px;">
    <h2>平台文档系统</h2>
    
  <!-- 顶部操作栏：整合所有搜索和排序逻辑 -->
  <div style="margin-bottom: 20px; display: flex; flex-wrap: wrap; align-items: center; gap: 15px;">
    <!-- 主要操作 -->
    <el-button type="success" @click="createDoc">新建文档</el-button>

    <!-- 高级搜索 -->
    <div style="display: flex; gap: 8px; align-items: center; border: 1px solid #dcdfe6; padding: 5px 15px; border-radius: 4px;">
      <el-icon><Search /></el-icon>
      <!-- 标题搜索 -->
      <el-input 
        v-model="searchKey" 
        placeholder="搜索标题..." 
        style="width: 150px;" 
        @input="handleSearch" 
        variant="borderless"
        clearable
      />
      <el-divider direction="vertical" />
      <!-- 创建者搜索 -->
      <el-input 
        v-model="searchCreator" 
        placeholder="搜索作者..." 
        style="width: 120px;" 
        @input="handleSearch" 
        clearable
      />
      <el-divider direction="vertical" />
      <!-- 日期筛选 -->
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        start-placeholder="开始"
        end-placeholder="结束"
        value-format="YYYY-MM-DD"
        style="width: 240px;"
        @change="onDateChange"
        size="small"
      />
    </div>

    <!-- 排序控制 -->
    <div style="display: flex; align-items: center; gap: 5px;">
      <el-select v-model="sortProp" @change="handleSearch" style="width: 120px;">
        <el-option label="按时间排序" value="documents.created_at" />
        <el-option label="按标题排序" value="title" />
      </el-select>
      <el-button @click="toggleOrder">
        {{ sortOrder === 'asc' ? '正序 ↑' : '倒序 ↓' }}
      </el-button>
    </div>
  </div>

    <el-container style="border: 1px solid #eee; height: calc(100vh - 160px);">
      <!-- 左侧：固定文件夹 + 自定义文件夹 -->
      <el-aside width="200px" style="border-right: 1px solid #eee">
        <el-menu :default-active="activeCategory" @select="handleCategorySelect">
          <el-menu-item index="全部">全部文档</el-menu-item>
          <el-menu-item index="重要">重要记录</el-menu-item>
          
          <el-divider content-position="left" style="margin: 10px 0; font-size: 12px;">我的文件夹</el-divider>
          
          <!-- 动态渲染自定义文件夹 -->
          <el-menu-item v-for="f in folders" :key="f.id" :index="'folder-' + f.id">
            <span>{{ f.name }}</span>
            <el-button type="text" icon="Delete" @click.stop="deleteFolder(f.id)" style="color:red; margin-left:10px" />
          </el-menu-item>
          
          <el-menu-item @click="addFolder">+ 新建文件夹</el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 右侧：文档列表 -->
      <el-main>
        <el-table :data="filteredDocs" style="width: 100%">
          <!-- 收藏按钮 -->
          <el-table-column width="50">
            <template #default="scope">
              <el-icon @click="toggleStar(scope.row)" style="cursor: pointer; font-size: 18px">
                <StarFilled v-if="scope.row.is_starred" style="color: #ff9900" />
                <Star v-else />
              </el-icon>
            </template>
          </el-table-column>

          <el-table-column prop="title" label="标题" />
          <el-table-column prop="creator_name" label="创建者" width="120" />
          <el-table-column prop="folder_name" label="所属文件夹" width="120" />
          <el-table-column prop="created_at" label="创建时间" width="180" />
          
          <el-table-column label="操作" width="450">

            <template #default="scope">
              <el-button size="small" @click="$router.push('/edit/' + scope.row.id)">进入编辑</el-button>
              
              <!-- 移动文件夹 -->
              <el-dropdown @command="(fId) => moveDoc(scope.row.id, fId)" style="margin: 0 10px">
                <el-button size="small">移动至</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item :command="0">全部文档</el-dropdown-item>
                    <el-dropdown-item v-for="f in folders" :key="f.id" :command="f.id">{{ f.name }}</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>

              <el-button size="small" type="warning" @click="handleRename(scope.row)">改名</el-button>
              <el-button size="small" type="danger" @click="handleDelete(scope.row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-main>
    </el-container>
  </div>
</template>
<script setup>
  import { ref, onMounted, computed } from 'vue'
  import axios from 'axios'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { Star, StarFilled, Delete } from '@element-plus/icons-vue'
  
  // 数据定义
  const docs = ref([])
  const folders = ref([])
  const activeCategory = ref('全部')
  const searchKey = ref('')
  const searchCreator = ref('')  // 创建者
  const dateRange = ref([])      // 日期原始数据
  const startDate = ref('')      // 格式化后的开始日期
  const endDate = ref('')        // 格式化后的结束日期
  const sortProp = ref('documents.created_at')
  const sortOrder = ref('asc')
  const user = JSON.parse(sessionStorage.getItem('user') || '{}')
  
  // 计算属性过滤 
  // 自动刷新
  const filteredDocs = computed(() => {
    let list = [...docs.value]
    
    // 侧边栏过滤
    if (activeCategory.value === '重要') {
      list = list.filter(d => d.is_starred)
    } else if (activeCategory.value.startsWith('folder-')) {
      const fId = parseInt(activeCategory.value.split('-')[1])
      list = list.filter(d => d.folder_id === fId)
    }
    
    return list
  })
  
  // 加载数据
  const loadDocs = async () => {
    const res = await axios.get(`http://localhost:8080/search`, {
      params: { keyword: searchKey.value, sort: sortProp.value, order: sortOrder.value }
    })
    docs.value = res.data
  }

  // 加载文件夹
  const loadFolders = async () => {
    const res = await axios.get(`http://localhost:8080/folders?user_id=${user.id}`)
    folders.value = res.data
  }
  
  // 事件处理
  const handleCategorySelect = (index) => {
    activeCategory.value = index
  }
  
  const toggleStar = async (doc) => {
    await axios.put(`http://localhost:8080/documents/${doc.id}/star`, { is_starred: !doc.is_starred })
    loadDocs() // 刷新数据
  }
  
  const moveDoc = async (docId, folderId) => {
    await axios.put(`http://localhost:8080/documents/${docId}/move`, { folder_id: folderId })
    ElMessage.success('已移至文件夹')
    loadDocs()
  }
  
  const addFolder = () => {
    ElMessageBox.prompt('文件夹名称', '新建文件夹').then(async ({ value }) => {
      await axios.post('http://localhost:8080/folders', { name: value, user_id: user.id })
      loadFolders()
    })
  }
  
  const deleteFolder = async (id) => {
    await axios.delete(`http://localhost:8080/folders/${id}`)
    loadFolders()
    loadDocs()
  }
  
    // 搜索与排序
  const onDateChange = (val) => {
    if (val) {
      startDate.value = val[0]
      endDate.value = val[1]
    } else {
      startDate.value = ''
      endDate.value = ''
    }
    handleSearch()
  }

  const handleSearch = async () => {
    try {
      const res = await axios.get(`http://localhost:8080/search`, {
        params: {
          keyword: searchKey.value,
          creator: searchCreator.value,
          start: startDate.value,
          end: endDate.value,
          sort: sortProp.value,
          order: sortOrder.value
        }
      })
      docs.value = res.data
    } catch (err) {
      console.error("搜索失败", err)
    }
  }


  const toggleOrder = () => {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
    loadDocs()
  }
  
  // 文档管理
  // 创建
  const createDoc = async () => {
    const tag = Math.floor(Math.random()*9000+1000)
    await axios.post('http://localhost:8080/documents', { title: `新文档_${tag}`, content: "", owner_id: user.id })
    loadDocs()
  }
  // 重命名
  const handleRename = (row) => {
    ElMessageBox.prompt('新名称', '改名', { inputValue: row.title }).then(async ({ value }) => {
      await axios.put(`http://localhost:8080/documents/${row.id}`, { title: value })
      loadDocs()
    })
  }
  // 删除
  const handleDelete = (id) => {
    ElMessageBox.confirm('删除此文档？').then(async () => {
      await axios.delete(`http://localhost:8080/documents/${id}`)
      loadDocs()
    })
  }
  
  onMounted(() => {
    loadDocs()
    loadFolders()
  })
  </script>
  <style scoped>

    h2 {
      color: #409EFF; 
      border-left: 4px solid #409EFF;
      padding-left: 15px;
      margin-bottom: 25px;
    }
    
    :deep(.el-table th) {
      background-color: #fafafa !important;
      color: #606266;
    }
    
    :deep(.el-table__row:hover > td) {
      background-color: #e6f7ff !important;
    }
    
    .search-bar {
      background: white;
      padding: 15px;
      border-radius: 8px;
      box-shadow: 0 2px 12px 0 rgba(0,0,0,0.05); /* 加个阴影 */
      margin-bottom: 20px;
    }
    </style>