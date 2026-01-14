<template>
  <div style="padding: 20px;">
    <!-- 顶部状态栏 -->
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
      <el-page-header @back="$router.push('/')" title="返回列表" />
      <div style="font-size: 14px; color: #409EFF; font-style: italic;">{{ typingStatus }}</div>
    </div>
    
    <div style="display: flex; gap: 20px;">
      <!-- 左侧：编辑器区域 -->
      <div style="flex: 3;">
        <div style="margin-bottom: 10px; display: flex; justify-content: space-between;">
          <el-button-group>
            <input type="file" id="importFile" style="display:none" @change="handleImport" accept=".txt,.md" />
            <el-button @click="triggerImport">导入文档</el-button>
            <el-button v-if="user.role !== 'viewer'" type="primary" @click="saveVersion">保存当前版本</el-button>
            <el-button @click="showVersions = true">历史版本记录</el-button>
            <el-button type="success" @click="exportTxt">导出为TXT</el-button>
          </el-button-group>
          <el-tag v-if="user.role === 'viewer'" type="info">只读模式</el-tag>
        </div>
        
        <div id="editor-container" style="height: 500px; border: 1px solid #ccc; background: #fff;"></div>
      </div>
      
      <!-- 右侧：实时聊天室 -->
      <div style="flex: 1; border: 1px solid #ddd; padding: 15px; height: 550px; display: flex; flex-direction: column; background: #fdfdfd; border-radius: 8px;">
        <h4 style="margin: 0 0 15px 0; color: #666; border-bottom: 1px solid #eee; padding-bottom: 10px;">在线交流</h4>
        <div style="flex: 1; overflow-y: auto; margin-bottom: 10px; padding-right: 5px;" id="chat-box">
          <div v-for="(m, index) in chatHistory" :key="index" style="margin-bottom: 12px;">
            <div :style="{ textAlign: m.user_id === user.id.toString() ? 'right' : 'left' }">
              <el-tag size="small" :type="m.user_id === user.id.toString() ? '' : 'success'">{{ m.username }}</el-tag>
              <div :style="chatBubbleStyle(m.user_id === user.id.toString())">{{ m.content }}</div>
            </div>
          </div>
        </div>
        <el-input v-model="chatInput" @keyup.enter="sendChat" placeholder="按回车发送..." />
      </div>
    </div>

    <!-- 悬浮批注按钮 -->
    <div style="position: fixed; right: 30px; bottom: 100px;">
      <el-badge :value="commentList.length" class="item">
        <el-button type="warning" icon="ChatDotRound" circle size="large" @click="openComments" />
      </el-badge>
    </div>

    <!-- 批注抽屉 -->
    <el-drawer v-model="showComments" title="文档批注 / 审阅" direction="rtl" size="380px" @open="loadComments">
      <div style="height: 100%; display: flex; flex-direction: column;">
        <div style="flex: 1; overflow-y: auto; padding: 10px;">
          <div v-for="c in commentList" :key="c.id" style="margin-bottom: 15px; padding: 12px; background: #f4f4f5; border-radius: 8px; position: relative;">
            <div style="font-size: 12px; color: #909399; margin-bottom: 5px;">
              第 {{ c.line_num }} 行 · {{ c.username }}
            </div>
            <div style="font-size: 14px; line-height: 1.5;">{{ c.content }}</div>
            <el-button 
              v-if="c.user_id == user.id" 
              type="danger" 
              link 
              :icon="Delete"
              style="position: absolute; right: 10px; top: 10px;" 
              @click="deleteComment(c.id)"
            >
              <!-- 建议临时加个文字，测试按钮到底在不在 -->
              删除
            </el-button>
          </div>
        </div>
        <div style="padding: 20px; border-top: 1px solid #eee;">
          <el-input v-model="newComment" placeholder="先在编辑器选位置，再输入..." type="textarea" :rows="3" />
          <el-button type="primary" style="margin-top: 10px; width: 100%;" @click="submitComment">提交批注</el-button>
        </div>
      </div>
    </el-drawer>

    <!-- 版本记录抽屉 -->
    <el-drawer v-model="showVersions" title="版本历史" direction="rtl" size="400px" @open="loadVersions">
      <el-table :data="versionList" size="small">
        <el-table-column prop="created_at" label="备份时间" width="150" />
        <el-table-column prop="version_name" label="名称" />
        <el-table-column label="操作" width="80">
          <template #default="scope">
            <el-button size="small" type="primary" link @click="restoreVersion(scope.row.content)">恢复</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-drawer>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'
import Quill from 'quill'
import 'quill/dist/quill.snow.css'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import { ChatDotRound, Delete } from '@element-plus/icons-vue'

const route = useRoute()
const docId = route.params.id
const user = JSON.parse(sessionStorage.getItem('user') || '{}')

// 状态变量
let quill = null
let ws = null
const typingStatus = ref('')
let typingTimer = null
const lastRange = ref(null) // 核心：记录光标最后位置

const chatInput = ref('')
const chatHistory = ref([])
const showComments = ref(false)
const commentList = ref([])
const newComment = ref('')
const showVersions = ref(false)
const versionList = ref([])

// 样式：聊天气泡
const chatBubbleStyle = (isMine) => ({
  background: isMine ? '#e1f3d8' : 'white',
  padding: '8px 12px',
  borderRadius: '4px',
  marginTop: '4px',
  fontSize: '14px',
  border: '1px solid #eee',
  display: 'inline-block',
  textAlign: 'left'
})

onMounted(async () => {
  // 1. 初始化编辑器
  quill = new Quill('#editor-container', { 
    theme: 'snow', 
    readOnly: user.role === 'viewer' 
  })

  // 记录最后光标位置，防止失焦
  quill.on('selection-change', (range) => {
    if (range) lastRange.value = range
  })

  // 2. 加载文档内容
  try {
    const res = await axios.get(`http://localhost:8080/documents/${docId}`)
    if (res.data.content) quill.setText(res.data.content, 'silent')
  } catch (e) { ElMessage.error("文档读取失败") }

  // 3. WebSocket 连通
  ws = new WebSocket(`ws://localhost:8080/ws/${docId}?userId=${user.id}`)

  ws.onmessage = (e) => {
    const data = JSON.parse(e.data)
    if (data.type === 'edit') {
      // 协同编辑逻辑
      typingStatus.value = `${data.username} 正在编辑...`
      clearTimeout(typingTimer)
      typingTimer = setTimeout(() => typingStatus.value = '', 2000)

      if (data.content !== quill.getText()) {
        const range = quill.getSelection()
        quill.setText(data.content, 'silent')
        if (range) quill.setSelection(range)
      }
    } else if (data.type === 'chat') {
      chatHistory.value.push(data)
      nextTick(() => {
        const box = document.getElementById('chat-box')
        box.scrollTop = box.scrollHeight
      })
    }
  }

  quill.on('text-change', (delta, oldDelta, source) => {
    if (source === 'user' && user.role !== 'viewer') {
      ws.send(JSON.stringify({ type: 'edit', doc_id: docId, content: quill.getText(), username: user.username }))
    }
  })
})

// 功能函数 
const sendChat = () => {
  if (!chatInput.value) return
  const msg = { type: 'chat', doc_id: docId, user_id: user.id.toString(), username: user.username, content: chatInput.value }
  ws.send(JSON.stringify(msg))
  chatHistory.value.push(msg)
  chatInput.value = ''
}

const openComments = () => { showComments.value = true }

const loadComments = async () => {
  const res = await axios.get(`http://localhost:8080/documents/${docId}/comments`)
  commentList.value = res.data
}

const submitComment = async () => {
  const range = lastRange.value // 使用最后记录的位置
  if (!range || !newComment.value) {
    ElMessage.warning('请先在编辑器点选位置并输入内容')
    return
  }
  const lineNum = quill.getLines(0, range.index).length
  await axios.post('http://localhost:8080/comments', {
    doc_id: parseInt(docId), user_id: user.id, content: newComment.value, line_num: lineNum
  })
  newComment.value = ''
  loadComments()
  ElMessage.success('批注已添加')
}

const deleteComment = async (id) => {
  await axios.delete(`http://localhost:8080/comments/${id}`)
  loadComments()
}

const loadVersions = async () => {
  const res = await axios.get(`http://localhost:8080/documents/${docId}/versions`)
  versionList.value = res.data
}

const saveVersion = () => {
  ElMessageBox.prompt('版本备注', '存为快照').then(async ({ value }) => {
    await axios.post(`http://localhost:8080/documents/${docId}/versions`, { name: value })
    ElMessage.success('已保存')
  })
}

const restoreVersion = (content) => {
  quill.setText(content, 'user')
  showVersions.value = false
  ElMessage.success('已恢复版本')
}

const triggerImport = () => document.getElementById('importFile').click()
const handleImport = (e) => {
  const file = e.target.files[0]
  const reader = new FileReader()
  reader.onload = (res) => {
    quill.setText(res.target.result, 'user')
    ws.send(JSON.stringify({ type: 'edit', doc_id: docId, content: res.target.result, username: user.username }))
  }
  reader.readAsText(file)
}

const exportTxt = () => {
  const blob = new Blob([quill.getText()], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `文档_${docId}.txt`
  a.click()
}

onUnmounted(() => { if (ws) ws.close() })
</script>
<style scoped>
  /* 修改编辑器的背景和边框 */
  :deep(.ql-container) {
    background-color: #ffffff;
    font-size: 16px;
    line-height: 1.8;
    color: #262626;
  }
  
  /* 修改编辑器上方工具栏的背景 */
  :deep(.ql-toolbar) {
    background-color: #f8f9fa;
    border-color: #dcdfe6 !important;
    border-radius: 8px 8px 0 0;
  }
  
  /* 修改聊天气泡的字体 */
  .chat-content {
    font-size: 13px;
    font-family: "Consolas", monospace; /* 聊天用等宽字体 */
  }
  </style>