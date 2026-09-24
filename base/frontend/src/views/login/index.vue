<template>
  <div class="login-page">
    <!-- 背景装饰 -->
    <div class="bg-blobs">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
    </div>

    <div class="login-container">
      <el-card class="login-card" shadow="never">
        <div class="login-header">
          <div class="brand">
            <div class="brand-icon">
              <img v-if="siteStore.logo" :src="siteStore.logo" alt="logo" class="brand-logo" />
              <el-icon v-else :size="32" color="#fff"><Management /></el-icon>
            </div>
            <div class="brand-text">
              <h1 class="title">{{ siteStore.platformName }}</h1>
              <p class="subtitle">统一的企业级管理底座</p>
            </div>
          </div>
        </div>

        <el-form
          :model="form"
          :rules="rules"
          ref="formRef"
          size="large"
          class="login-form"
          @keyup.enter="handleLogin"
        >
          <el-form-item prop="tenantCode">
            <el-input
              v-model="form.tenantCode"
              placeholder="租户编码（平台管理员可留空）"
              :prefix-icon="OfficeBuilding"
              class="login-input"
            />
          </el-form-item>

          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="用户名"
              :prefix-icon="User"
              class="login-input"
            />
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="密码"
              :prefix-icon="Lock"
              show-password
              class="login-input"
            />
          </el-form-item>

          <el-form-item v-if="captchaEnabled" prop="captchaCode">
            <div class="captcha-row">
              <el-input
                v-model="form.captchaCode"
                placeholder="请输入验证码"
                maxlength="5"
                :prefix-icon="Grid"
                class="login-input captcha-input"
                @keyup.enter="handleLogin"
              />
              <div class="captcha-image" @click="loadCaptcha" title="点击刷新验证码">
                <img v-if="captchaImage" :src="captchaImage" alt="验证码" />
                <div v-else class="captcha-placeholder">点击刷新</div>
              </div>
            </div>
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              size="large"
              class="login-btn"
              :loading="loading"
              @click="handleLogin"
            >
              {{ loading ? '登录中...' : '立即登录' }}
            </el-button>
          </el-form-item>
        </el-form>

        <div class="login-footer">
          <p>{{ siteStore.copyright || `© ${currentYear} Base Platform. All rights reserved.` }}</p>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Management, OfficeBuilding, User, Lock, Grid } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { getCaptcha } from '@/api/auth'

const router = useRouter()
const userStore = useUserStore()
const siteStore = useSiteStore()
const formRef = ref<any>(null)
const loading = ref(false)
const captchaImage = ref('')
const captchaId = ref('')
// 登录验证码开关（来自系统设置 → 安全策略 captchaEnabled，公开站点信息接口下发）
const captchaEnabled = ref(true)

const currentYear = computed(() => new Date().getFullYear())

const form = reactive({
  tenantCode: '',
  username: '',
  password: '',
  captchaCode: ''
})

const rules = computed(() => ({
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  ...(captchaEnabled.value
    ? { captchaCode: [{ required: true, message: '请输入验证码', trigger: 'blur' }] }
    : {})
}))

const loadCaptcha = async () => {
  if (!captchaEnabled.value) return
  try {
    const res: any = await getCaptcha()
    captchaId.value = res.data.captcha_id
    captchaImage.value = res.data.captcha_img
    form.captchaCode = ''
  } catch (error) {
    ElMessage.warning('验证码加载失败，请稍后重试')
  }
}

// 验证码开关确定后再加载验证码图片（关闭时不请求，避免无谓的接口调用）
// 站点信息（平台名称/Logo/版权）与开关同一次请求拿到，避免重复调接口
const loadSiteInfo = async () => {
  await siteStore.fetchSiteInfo()
  captchaEnabled.value = siteStore.captchaEnabled
  if (captchaEnabled.value) {
    loadCaptcha()
  }
}

const handleLogin = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await userStore.login({
      username: form.username,
      password: form.password,
      tenantCode: form.tenantCode || undefined,
      captcha_id: captchaEnabled.value ? captchaId.value : '',
      captcha_code: captchaEnabled.value ? form.captchaCode : ''
    })
    ElMessage.success('登录成功')
    try {
      await router.push('/')
    } catch (navError) {
      console.error('[base] 登录后跳转失败', navError)
      ElMessage.error('登录后跳转失败，请刷新页面重试')
    }
  } catch (error: any) {
    // 失败详情已由 request 拦截器弹出（统一响应 HTTP 200 + code!=0，错误信息在 error.message 上），
    // 这里的 response.data.message 恒为 undefined，再弹一次只会把真实原因覆盖成「登录失败」
    console.error('[base] 登录失败', error)
    loadCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(loadSiteInfo)
</script>

<style scoped lang="scss">
.login-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 50%, #bae6fd 100%);
}

/* 背景装饰 */
.bg-blobs {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;

  .blob {
    position: absolute;
    border-radius: 50%;
    filter: blur(80px);
    opacity: 0.45;
    animation: float 20s ease-in-out infinite;
  }

  .blob-1 {
    width: 500px;
    height: 500px;
    background: #93c5fd;
    top: -120px;
    left: -120px;
    animation-delay: 0s;
  }

  .blob-2 {
    width: 400px;
    height: 400px;
    background: #a5b4fc;
    bottom: -80px;
    right: -80px;
    animation-delay: -7s;
  }

  .blob-3 {
    width: 300px;
    height: 300px;
    background: #67e8f9;
    top: 50%;
    left: 60%;
    animation-delay: -14s;
  }
}

@keyframes float {
  0%, 100% {
    transform: translate(0, 0) scale(1);
  }
  33% {
    transform: translate(30px, -50px) scale(1.05);
  }
  66% {
    transform: translate(-20px, 30px) scale(0.95);
  }
}

.login-container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 460px;
  padding: 20px;
}

.login-card {
  width: 100%;
  padding: 40px 36px 32px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.5);
  box-shadow: 0 25px 80px rgba(0, 0, 0, 0.25);
  animation: card-enter 0.6s ease-out;
}

@keyframes card-enter {
  from {
    opacity: 0;
    transform: translateY(24px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

:deep(.el-card__body) {
  padding: 0;
}

.login-header {
  margin-bottom: 36px;

  .brand {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 16px;
  }

  .brand-icon {
    width: 58px;
    height: 58px;
    border-radius: 14px;
    background: linear-gradient(135deg, #2563eb 0%, #4f46e5 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 10px 25px rgba(37, 99, 235, 0.35);
    overflow: hidden;
  }

  /* 系统设置 → 基础配置 里配了 Logo 就显示图片，否则回退到默认图标 */
  .brand-logo {
    width: 100%;
    height: 100%;
    object-fit: contain;
    background: #fff;
  }

  .brand-text {
    text-align: left;
  }

  .title {
    margin: 0;
    font-size: 26px;
    font-weight: 700;
    color: #1e293b;
    letter-spacing: 0.5px;
  }

  .subtitle {
    margin: 6px 0 0;
    font-size: 14px;
    color: #64748b;
  }
}

.login-form {
  .el-form-item {
    margin-bottom: 22px;
  }
}

.login-input {
  :deep(.el-input__wrapper) {
    border-radius: 10px;
    box-shadow: 0 0 0 1px #e2e8f0 inset;
    padding: 2px 14px;
    transition: all 0.25s ease;

    &:hover,
    &.is-focus {
      box-shadow: 0 0 0 1px #3b82f6 inset, 0 4px 12px rgba(59, 130, 246, 0.12);
    }
  }

  :deep(.el-input__inner) {
    height: 46px;
    font-size: 15px;

    &::placeholder {
      color: #94a3b8;
    }
  }

  :deep(.el-input__icon) {
    color: #94a3b8;
    font-size: 18px;
  }
}

.captcha-row {
  display: flex;
  width: 100%;
  gap: 12px;

  .captcha-input {
    flex: 1;
    min-width: 0;
  }
}

/* 验证码图片容器尺寸与 portal 保持一致：150x50，object-fit: cover */
.captcha-image {
  width: 150px;
  height: 50px;
  flex-shrink: 0;
  align-self: center;
  border-radius: 10px;
  overflow: hidden;
  cursor: pointer;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.25s ease;

  &:hover {
    border-color: #3b82f6;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  }

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.captcha-placeholder {
  font-size: 12px;
  color: #94a3b8;
}

.login-btn {
  width: 100%;
  height: 50px;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  letter-spacing: 1px;
  margin-top: 8px;
  background: linear-gradient(135deg, #2563eb 0%, #4f46e5 100%);
  border: none;
  box-shadow: 0 10px 25px rgba(37, 99, 235, 0.35);
  transition: all 0.25s ease;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 14px 32px rgba(37, 99, 235, 0.45);
  }

  &:active {
    transform: translateY(0);
  }
}

.login-footer {
  margin-top: 28px;
  text-align: center;

  p {
    margin: 0;
    font-size: 12px;
    color: #94a3b8;
  }
}

@media (max-width: 480px) {
  .login-card {
    padding: 32px 24px 28px;
  }

  .login-header {
    .brand {
      flex-direction: column;
      gap: 12px;
    }

    .brand-text {
      text-align: center;
    }
  }
}
</style>
