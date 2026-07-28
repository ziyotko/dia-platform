<template>
  <div class="login-page">
    <div class="login-card">
      <div class="card-header">
        <el-icon :size="36" color="#1a6fb5"><OfficeBuilding /></el-icon>
        <h2>会员登录</h2>
        <p>中国电器工业协会会员系统</p>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" size="large">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item prop="captchaCode">
          <div class="captcha-row">
            <el-input v-model="form.captchaCode" placeholder="验证码" prefix-icon="Key" />
            <img :src="captchaImage" class="captcha-img" @click="loadCaptcha" title="点击刷新" />
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="login-btn" :loading="loading" @click="handleLogin">登 录</el-button>
        </el-form-item>
      </el-form>
      <div class="card-footer">
        <router-link to="/register">还没有账号？立即注册</router-link>
        <router-link to="/reset-password">忘记密码？</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { authApi } from '@/api/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const captchaImage = ref('')
const captchaId = ref('')
const formRef = ref()

const form = reactive({
  username: '',
  password: '',
  captchaCode: ''
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captchaCode: [{ required: true, message: '请输入验证码', trigger: 'blur' }]
}

onMounted(() => loadCaptcha())

async function loadCaptcha() {
  try {
    const res = await authApi.getCaptcha()
    captchaId.value = res.data.captcha_id
    captchaImage.value = res.data.captcha_image
  } catch {}
}

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await userStore.login(form.username, form.password, captchaId.value, form.captchaCode)
    ElMessage.success('登录成功')
    if (userStore.isAdmin) {
      router.push('/admin/dashboard')
    } else {
      router.push('/dashboard')
    }
  } catch (e: any) {
    loadCaptcha()
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.login-page {
  min-height: calc(100vh - 200px);
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f5f7fa, #e5e7eb);
  padding: 40px;
}
.login-card {
  width: 420px;
  background: #fff;
  border-radius: 16px;
  padding: 48px 40px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.08);
  .card-header {
    text-align: center;
    margin-bottom: 36px;
    h2 { margin: 12px 0 4px; font-size: 24px; color: #1f2937; }
    p { color: #9ca3af; font-size: 14px; }
  }
  .captcha-row {
    display: flex;
    gap: 12px;
    .captcha-img {
      width: 120px;
      height: 40px;
      border-radius: 8px;
      cursor: pointer;
      border: 1px solid #e5e7eb;
    }
  }
  .login-btn { width: 100%; }
  .card-footer {
    display: flex;
    justify-content: space-between;
    margin-top: 16px;
    a { font-size: 13px; color: #1a6fb5; text-decoration: none; }
  }
}
</style>
