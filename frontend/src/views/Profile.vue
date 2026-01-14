<template>
    <div style="max-width: 600px; margin: 20px auto; padding: 20px;">
      <!-- 增加返回按钮 -->
      <el-page-header @back="$router.push('/')" content="个人资料设置" style="margin-bottom: 30px;" />
  
      <!-- 头像上传部分 -->
      <el-form-item label="个人头像">
        <el-upload
          class="avatar-uploader"
          action="http://localhost:8080/user/avatar"
          :data="{ user_id: user.id }"
          name="avatar"
          :show-file-list="false"
          :on-success="handleAvatarSuccess"
        >
          <img v-if="user.avatar" :src="user.avatar" class="avatar-img" />
          <el-icon v-else class="avatar-uploader-icon"><Plus /></el-icon>
        </el-upload>
        <div style="font-size: 12px; color: #999; margin-left: 10px;">点击图片更换头像</div>
      </el-form-item>

      <el-form label-width="100px" :model="user" style="background: #fff; padding: 30px; border-radius: 8px; border: 1px solid #eee;">
        <el-form-item label="用户名">
          <el-input v-model="user.username" placeholder="既是登录账号也是用户名" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="user.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="user.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="个人简介">
          <el-input v-model="user.bio" type="textarea" rows="4" placeholder="向大家介绍一下自己吧" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveProfile" style="width: 120px;">保存修改</el-button>
        </el-form-item>
      </el-form>
    </div>
</template>
  
<script setup>
    import { ref, onMounted } from 'vue'
    import axios from 'axios'
    import { ElMessage } from 'element-plus'
    import { Plus } from '@element-plus/icons-vue'
    import { useRouter } from 'vue-router'
    
    const router = useRouter()
    const user = ref({
      id: 0,
      username: '',
      email: '',
      phone: '',
      bio: ''
    })

    const handleAvatarSuccess = (res) => {
      user.value.avatar = res.url
      // 更新本地存储，让导航栏头像也能变
      const localUser = JSON.parse(sessionStorage.getItem('user'))
      localUser.avatar = res.url
      sessionStorage.setItem('user', JSON.stringify(localUser))
      ElMessage.success('头像上传成功')
    }
    
    onMounted(() => {
      const savedUser = JSON.parse(sessionStorage.getItem('user') || '{}')
      if (savedUser.id) {
        user.value = savedUser
      } else {
        router.push('/login')
      }
    })
    
    const saveProfile = async () => {
      try {
        const res = await axios.put('http://localhost:8080/user/profile', user.value)
        // 更新本地存储
        sessionStorage.setItem('user', JSON.stringify(user.value))
        ElMessage.success('资料更新成功')
      } catch (err) {
        ElMessage.error('保存失败')
      }
    }
</script>

<style scoped>
.avatar-img {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  object-fit: cover;
}
.avatar-uploader-icon {
  font-size: 28px;
  color: #8c939d;
  width: 100px;
  height: 100px;
  line-height: 100px;
  text-align: center;
  border: 1px dashed #d9d9d9;
  border-radius: 50%;
}
</style>