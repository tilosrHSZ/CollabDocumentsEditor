<template>
  <el-config-provider>
    <!-- 导航栏 -->
    <div v-if="!['/login', '/register'].includes($route.path)" class="navbar" style="display: flex; justify-content: space-between; align-items: center; padding: 10px 20px; background: #fff; border-bottom: 1px solid #eee;">
      <div style="font-weight: bold; font-size: 18px;">多人协作编辑系统</div>
      <div style="display: flex; align-items: center;">
        <el-avatar :size="32" :src="userAvatar" style="margin-right: 10px;" />
        <span style="margin-right: 20px;">欢迎，{{ username }}</span>
      </div>
      <div>
        <el-link type="primary" :underline="false" @click="$router.push('/profile')" style="margin-right: 15px;">个人资料</el-link>
        <el-button type="danger" size="small" @click="logout">退出</el-button>
      </div>
    </div>
    
    <router-view />
  </el-config-provider>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const username = ref('未登录')
const userAvatar = ref('')

const updateUserInfo = () => {
  const user = JSON.parse(sessionStorage.getItem('user') || '{}')
  username.value = user.username || '未登录'
  userAvatar.value = user.avatar || ''
}

updateUserInfo()
watch(() => route.path, () => {
  updateUserInfo()
})

const logout = () => {
  localStorage.clear()
  sessionStorage.clear()
  window.location.href = '/login'
}
</script>
