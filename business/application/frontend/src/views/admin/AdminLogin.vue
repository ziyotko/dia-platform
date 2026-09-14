<template>
  <div class="login-page">
    <el-card class="login-card">
      <h2>管理端登录</h2>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="0" size="large">
        <el-form-item prop="username"><el-input v-model="form.username" placeholder="用户名" prefix-icon="User" /></el-form-item>
        <el-form-item prop="password"><el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" show-password @keyup.enter="handleLogin" /></el-form-item>
        <el-form-item prop="captchaCode">
          <div class="captcha-row">
            <el-input
              v-model="form.captchaCode"
              placeholder="验证码"
              maxlength="5"
              prefix-icon="Grid"
              clearable
              @keyup.enter="handleLogin"
            />
            <div class="captcha-image" @click="loadCaptcha">
              <img v-if="captchaImage" :src="captchaImage" alt="验证码" />
              <div v-else class="captcha-placeholder">点击刷新</div>
            </div>
          </div>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="handleLogin" :loading="loading" style="width:100%">登 录</el-button></el-form-item>
      </el-form>
      <div style="text-align:center;color:#999;font-size:13px">默认账号：admin / manager / reviewer</div>
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
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captchaCode: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
}

async function loadCaptcha() {
  try {
    const res = await authApi.getCaptcha()
    captchaId.value = res.data.captcha_id
    captchaImage.value = res.data.captcha_img
  } catch {
    captchaImage.value = ''
  }
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
.login-card h2 { text-align: center; margin-bottom: 24px; color: #002fa7; }

.captcha-row { display: flex; gap: 12px; width: 100%; }
.captcha-row .el-input { flex: 1; }
.captcha-image {
  width: 150px;
  height: 50px;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  background: #f5f7fb;
  border: 1px solid #e3e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  align-self: center;
}
.captcha-image img { width: 100%; height: 100%; object-fit: cover; }
.captcha-placeholder { color: #98a2b3; font-size: 12px; }
</style>
