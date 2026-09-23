<template>
  <div class="login-page">
    <el-card class="login-card">
      <h2>申报人注册</h2>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="0">
        <el-form-item prop="username"><el-input v-model="form.username" placeholder="用户名" prefix-icon="User" /></el-form-item>
        <el-form-item prop="realName"><el-input v-model="form.realName" placeholder="真实姓名" prefix-icon="Postcard" /></el-form-item>
        <el-form-item prop="password"><el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" show-password /></el-form-item>
        <el-form-item prop="confirm"><el-input v-model="form.confirm" type="password" placeholder="确认密码" prefix-icon="Lock" show-password /></el-form-item>
        <el-form-item><el-input v-model="form.organization" placeholder="所在单位" prefix-icon="OfficeBuilding" /></el-form-item>
        <el-form-item><el-input v-model="form.phone" placeholder="联系电话" prefix-icon="Phone" /></el-form-item>
        <el-form-item prop="email"><el-input v-model="form.email" placeholder="电子邮箱" prefix-icon="Message" /></el-form-item>
        <el-form-item prop="captchaCode">
          <div class="captcha-row">
            <el-input
              v-model="form.captchaCode"
              placeholder="验证码"
              maxlength="5"
              prefix-icon="Grid"
              clearable
              @keyup.enter="handleRegister"
            />
            <div class="captcha-image" @click="loadCaptcha">
              <img v-if="captchaImage" :src="captchaImage" alt="验证码" />
              <div v-else class="captcha-placeholder">点击刷新</div>
            </div>
          </div>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="handleRegister" :loading="loading" style="width:100%">注 册</el-button></el-form-item>
      </el-form>
      <div style="text-align:center"><router-link to="/login">已有账号？去登录</router-link></div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { authApi } from '@/api/auth'

const router = useRouter()
const formRef = ref()
const loading = ref(false)
const captchaId = ref('')
const captchaImage = ref('')
const form = reactive({ username: '', realName: '', password: '', confirm: '', organization: '', phone: '', email: '', captchaCode: '' })

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  realName: [{ required: true, message: '请输入真实姓名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  confirm: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: (_r: any, v: string, cb: any) => (v === form.password ? cb() : cb(new Error('两次密码不一致'))), trigger: 'blur' },
  ],
  captchaCode: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
}

// 验证码由后端强制校验（注册是公开接口，不加验证码可被脚本无限建号），失败后需刷新换一张
async function loadCaptcha() {
  try {
    const res = await authApi.getCaptcha()
    captchaId.value = res.data.captcha_id
    captchaImage.value = res.data.captcha_img
  } catch {
    captchaImage.value = ''
  }
}

async function handleRegister() {
  await formRef.value?.validate()
  loading.value = true
  try {
    await authApi.userRegister({
      username: form.username, realName: form.realName, password: form.password,
      organization: form.organization, phone: form.phone, email: form.email,
      captcha_id: captchaId.value, captcha_code: form.captchaCode,
    })
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } catch {
    form.captchaCode = ''
    loadCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(loadCaptcha)
</script>

<style scoped>
.login-page { display: flex; justify-content: center; align-items: center; min-height: 80vh; }
.login-card { width: 440px; }
.login-card h2 { text-align: center; margin-bottom: 24px; color: #002fa7; }

.captcha-row { display: flex; gap: 12px; width: 100%; }
.captcha-row .el-input { flex: 1; }
.captcha-image {
  width: 150px;
  height: 40px;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  background: #f5f7fa;
  display: flex;
  align-items: center;
  justify-content: center;
}
.captcha-image img { width: 100%; height: 100%; display: block; }
.captcha-placeholder { color: #909399; font-size: 12px; }
</style>
