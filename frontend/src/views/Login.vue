<template>
    <div style="max-width: 300px; margin: 100px auto;">
      <h2>用户登录</h2>
      <el-input v-model="form.username" placeholder="用户名" style="margin-bottom: 10px;" />
      <el-input v-model="form.password" type="password" placeholder="密码" style="margin-bottom: 10px;" />
      <el-button type="primary" @click="doLogin" style="width: 100%">登录</el-button>
      <el-link type="primary" @click="$router.push('/register')">还没有账号？立即注册</el-link>
    </div>
  </template>
  <script setup>
  import { ref } from 'vue'
  import axios from 'axios'
  import { useRouter } from 'vue-router'
  
  const router = useRouter()
  const form = ref({ username: '', password: '' })
  
  const doLogin = async () => {
    const res = await axios.post('http://localhost:8080/login', form.value)
    sessionStorage.setItem('user', JSON.stringify(res.data.user)) // 保存登录状态
    router.push('/')
  }
  </script>