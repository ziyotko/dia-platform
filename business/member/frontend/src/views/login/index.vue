<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-left">
        <div class="login-brand">
          <el-icon size="72" color="#fff"><OfficeBuilding /></el-icon>
          <h1>XXXXXXXXXXXXXXXXX</h1>
          <p>会员服务系统</p>
        </div>
        <div class="login-features">
          <div class="feature-item">
            <el-icon><Check /></el-icon>
            <span>在线入会申请，全流程跟踪</span>
          </div>
          <div class="feature-item">
            <el-icon><Check /></el-icon>
            <span>会费在线缴纳，年度管理</span>
          </div>
          <div class="feature-item">
            <el-icon><Check /></el-icon>
            <span>会员证书在线生成与下载</span>
          </div>
          <div class="feature-item">
            <el-icon><Check /></el-icon>
            <span>行业动态与技术交流</span>
          </div>
          <div class="feature-item">
            <el-icon><Check /></el-icon>
            <span>意见反馈与诉求直达</span>
          </div>
        </div>
      </div>
      <div class="login-right">
        <div class="login-form-wrapper">
          <h2>会员登录</h2>
          <el-form ref="formRef" :model="form" :rules="rules" size="large" class="login-form">
            <el-form-item prop="username">
              <el-input
                v-model="form.username"
                placeholder="请输入用户名"
                :prefix-icon="User"
                clearable
              />
            </el-form-item>
            <el-form-item prop="password">
              <el-input
                v-model="form.password"
                type="password"
                placeholder="请输入密码"
                :prefix-icon="Lock"
                show-password
                clearable
                @keyup.enter="handleLogin"
              />
            </el-form-item>
            <el-form-item prop="captchaCode">
              <div class="captcha-row">
                <el-input
                  v-model="form.captchaCode"
                  placeholder="请输入验证码"
                  :prefix-icon="Grid"
                  clearable
                  @keyup.enter="handleLogin"
                />
                <div class="captcha-image" @click="loadCaptcha">
                  <img v-if="captchaImage" :src="captchaImage" alt="验证码" />
                  <div v-else class="captcha-placeholder">点击刷新</div>
                </div>
              </div>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" class="login-btn" :loading="loading" @click="handleLogin">
                登 录
              </el-button>
            </el-form-item>
          </el-form>
          <div class="login-links">
            <router-link to="/register">还没有账号？立即注册</router-link>
            <router-link to="/reset-password">忘记密码？</router-link>
          </div>
        </div>
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
import { User, Lock, Grid, Check } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const captchaImage = ref('')
const captchaId = ref('')
const formRef = ref()
const siteInfo = ref<any>({})

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

onMounted(async () => {
  loadCaptcha()
  try {
    const res = await authApi.getSiteInfo()
    siteInfo.value = res.data || {}
  } catch {}
})

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
      router.push('/member/dashboard')
    }
  } catch {
    loadCaptcha()
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
$primary: #1a6fb5;
$primary-dark: #0d4f85;
$primary-gradient: linear-gradient(160deg, #1a6fb5 0%, #0d4f85 100%);

.login-container {
  min-height: calc(100vh - 200px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #e8f4fd 0%, #f0f7ff 50%, #e8f4fd 100%);
  position: relative;
  overflow: hidden;
  padding: 40px 24px;

  &::before {
    content: '';
    position: absolute;
    width: 800px;
    height: 800px;
    background: radial-gradient(circle, rgba(26, 111, 181, 0.08) 0%, transparent 70%);
    top: -200px;
    left: -200px;
  }

  &::after {
    content: '';
    position: absolute;
    width: 600px;
    height: 600px;
    background: radial-gradient(circle, rgba(26, 111, 181, 0.06) 0%, transparent 70%);
    bottom: -100px;
    right: -100px;
  }
}

.login-box {
  display: flex;
  width: 900px;
  min-height: 520px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(26, 111, 181, 0.15);
  overflow: hidden;
  z-index: 1;
}

.login-left {
  width: 420px;
  background: $primary-gradient;
  padding: 60px 40px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  color: #fff;
}

.login-brand {
  text-align: center;
  margin-bottom: 50px;

  h1 {
    font-size: 24px;
    margin: 16px 0 8px;
    font-weight: 600;
  }

  p {
    font-size: 14px;
    opacity: 0.85;
    letter-spacing: 1px;
  }
}

.login-features {
  .feature-item {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 20px;
    font-size: 15px;

    .el-icon {
      width: 24px;
      height: 24px;
      background: rgba(255, 255, 255, 0.2);
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }
}

.login-right {
  flex: 1;
  padding: 60px 50px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.login-form-wrapper {
  h2 {
    font-size: 22px;
    color: #303133;
    margin-bottom: 30px;
    text-align: center;
    font-weight: 600;
  }
}

.login-form {
  :deep(.el-input__wrapper) {
    box-shadow: 0 0 0 1px #d9ecff inset;
    background: #f5f9ff;
    border-radius: 8px;
    transition: box-shadow 0.2s;

    &.is-focus {
      box-shadow: 0 0 0 1px $primary inset;
    }
  }
}

.captcha-row {
  display: flex;
  gap: 12px;
  width: 100%;

  .el-input {
    flex: 1;
  }
}

.captcha-image {
  width: 120px;
  height: 40px;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  background: #f5f9ff;
  border: 1px solid #d9ecff;
  display: flex;
  align-items: center;
  justify-content: center;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.captcha-placeholder {
  color: #909399;
  font-size: 12px;
}

.login-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 8px;
  background: linear-gradient(135deg, #1a6fb5 0%, #0d4f85 100%);
  border: none;

  &:hover {
    background: linear-gradient(135deg, #4a9fd5 0%, #1a6fb5 100%);
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(26, 111, 181, 0.3);
  }
}

.login-links {
  display: flex;
  justify-content: space-between;
  margin-top: 16px;

  a {
    font-size: 13px;
    color: $primary;
    text-decoration: none;

    &:hover {
      color: $primary-dark;
      text-decoration: underline;
    }
  }
}

.login-footer {
  margin-top: 32px;
  padding: 12px 24px;
  color: #7a8b9a;
  font-size: 12px;
  z-index: 1;
  text-align: center;

  .footer-dot {
    display: inline-block;
    width: 3px;
    height: 3px;
    background: #b0c4de;
    border-radius: 50%;
    margin: 0 10px;
    vertical-align: middle;
  }
}

@media (max-width: 768px) {
  .login-box {
    flex-direction: column;
    width: 100%;
    max-width: 420px;
  }
  .login-left {
    width: 100%;
    padding: 40px 30px;
  }
  .login-right {
    padding: 40px 30px;
  }
}
</style>
