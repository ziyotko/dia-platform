<template>
  <div class="login-page">
    <el-card class="login-card">
      <h2>申报人登录</h2>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="0">
        <el-form-item prop="username"><el-input v-model="form.username" placeholder="用户名" prefix-icon="User" /></el-form-item>
        <el-form-item prop="password"><el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" show-password /></el-form-item>
        <el-form-item>
          <div style="display:flex;gap:12px;align-items:center">
            <el-input v-model="form.captchaCode" placeholder="验证码" style="flex:1" />
            <img :src="captchaImage" @click="loadCaptcha" style="height:40px;cursor:pointer" alt="验证码" />
          </div>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="handleLogin" :loading="loading" style="width:100%">登 录</el-button></el-form-item>
      </el-form>
      <div style="text-align:center"><router-link to="/register">还没有账号？立即注册</router-link></div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { authApi } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref()
const loading = ref(false)
const captchaId = ref('')
const captchaImage = ref('')

const form = reactive({ username: '', password: '', captchaCode: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
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
    await userStore.login(form.username, form.password, captchaId.value, form.captchaCode)
    ElMessage.success('登录成功')
    router.push('/member/dashboard')
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
.login-card h2 { text-align: center; margin-bottom: 24px; color: #002fa7; }
</style>
