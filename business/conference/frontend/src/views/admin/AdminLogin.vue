<template>
  <div class="login-page">
    <el-card class="login-card">
      <h2>管理后台登录</h2>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="0">
        <el-form-item prop="username"><el-input v-model="form.username" placeholder="管理员账号" prefix-icon="User" /></el-form-item>
        <el-form-item prop="password"><el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" show-password /></el-form-item>
        <el-form-item>
          <div style="display:flex;gap:12px;align-items:center">
            <el-input v-model="form.captchaCode" placeholder="验证码" style="flex:1" />
            <img :src="captchaImage" @click="loadCaptcha" style="height:40px;cursor:pointer" alt="验证码" />
          </div>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="handleLogin" :loading="loading" style="width:100%">管理员登录</el-button></el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { authApi } from '@/api/auth'
import { useAdminStore } from '@/stores/admin'

const router = useRouter()
const adminStore = useAdminStore()
const formRef = ref()
const loading = ref(false)
const captchaId = ref('')
const captchaImage = ref('')

const form = reactive({ username: '', password: '', captchaCode: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名' }],
  password: [{ required: true, message: '请输入密码' }],
}

async function loadCaptcha() {
  const res = await authApi.getCaptcha()
  captchaId.value = res.data.captchaId
  captchaImage.value = res.data.captchaImage
}

async function handleLogin() {
  await formRef.value?.validate()
  loading.value = true
  try {
    await adminStore.login(form.username, form.password, captchaId.value, form.captchaCode)
    ElMessage.success('登录成功')
    router.push('/admin/dashboard')
  } catch (e: any) {
    loadCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(loadCaptcha)
</script>

<style scoped>
.login-page { display: flex; justify-content: center; align-items: center; min-height: 80vh; }
.login-card { width: 400px; }
.login-card h2 { text-align: center; margin-bottom: 24px; color: #2563eb; }
</style>
