<template>
    <div class="auth-container">
      <el-card class="auth-card">
        <template #header>
          <div class="card-header">
            <span>新用户注册</span>
          </div>
        </template>
        
        <el-form :model="form" label-width="80px">
          <el-form-item label="用户名">
            <el-input v-model="form.username" placeholder="你的登录账号及昵称" />
          </el-form-item>
          
          <el-form-item label="密码">
            <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password />
          </el-form-item>
          
          <el-form-item label="邮箱">
            <el-input v-model="form.email" placeholder="example@mail.com" />
          </el-form-item>
          
          <el-form-item label="手机号">
            <el-input v-model="form.phone" placeholder="11位手机号" />
          </el-form-item>
          <el-form-item label="选择身份">
            <el-select v-model="form.role" placeholder="请选择您的角色" style="width: 100%">
                <el-option label="管理员 (Admin)" value="admin" />
                <el-option label="编辑者 (Editor)" value="editor" />
                <el-option label="查看者 (Viewer)" value="viewer" />
            </el-select>
            </el-form-item>
          <div style="margin-top: 20px;">
            <el-button type="primary" @click="handleRegister" style="width: 100%;">立即注册</el-button>
          </div>
          
          <div style="text-align: center; margin-top: 15px;">
            <el-link type="info" @click="$router.push('/login')">已有账号？返回登录</el-link>
          </div>
        </el-form>
      </el-card>
    </div>
  </template>
  
  <script setup>
  import { ref } from 'vue'
  import axios from 'axios'
  import { useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  
  const router = useRouter()
  const form = ref({
    username: '',
    password: '',
    email: '',
    phone: '',
    role: 'editor'
  })
  
  const handleRegister = async () => {
    if (!form.value.username || !form.value.password) {
      ElMessage.error('用户名和密码不能为空')
      return
    }
    
    try {
      const res = await axios.post('http://localhost:8080/register', form.value)
      ElMessage.success('注册成功！请登录')
      router.push('/login')
    } catch (err) {
      ElMessage.error(err.response?.data?.error || '注册失败，用户名可能已存在')
    }
  }
  </script>
  
  <style scoped>
  .auth-container {
    height: 100vh;
    display: flex;
    justify-content: center;
    align-items: center;
    background-color: #f5f7fa;
  }
  .auth-card {
    width: 400px;
  }
  .card-header {
    text-align: center;
    font-weight: bold;
    font-size: 18px;
  }
  </style>