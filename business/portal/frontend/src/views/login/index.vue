<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-left">
        <div class="login-brand">
          <img v-if="siteInfo.logo" :src="resolveLogoUrl(siteInfo.logo)" alt="logo" class="login-logo" />
          <el-icon v-else size="72" color="#409eff"><Platform /></el-icon>
          <h1>{{ siteInfo.siteName }}</h1>
          <p>统一内容管理平台</p>
        </div>
        <div class="login-features">
           <div class="feature-item">
            <el-icon><Check /></el-icon>
            <span>复杂组织机构管理</span>
          </div>
          <div class="feature-item">
            <el-icon><Check /></el-icon>
            <span>统一内容管理</span>
          </div>
          <div class="feature-item">
            <el-icon><Check /></el-icon>
            <span>站点静态化</span>
          </div>
          <div class="feature-item">
            <el-icon><Check /></el-icon>
            <span>多角色权限控制</span>
          </div>
          <div class="feature-item">
            <el-icon><Check /></el-icon>
            <span>操作日志审计</span>
          </div>
        </div>
      </div>
      <div class="login-right">
        <div class="login-form-wrapper">
          <h2>欢迎回来</h2>
          <p class="login-subtitle">登录门户一体化统一管理后台</p>
          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            size="large"
            class="login-form"
          >
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
                <div class="captcha-image" @click="refreshCaptcha">
                  <img v-if="captchaImage" :src="captchaImage" alt="验证码" />
                  <div v-else class="captcha-placeholder">点击刷新</div>
                </div>
              </div>
            </el-form-item>
            <el-form-item>
              <el-button
                type="primary"
                class="login-btn"
                :loading="loading"
                @click="handleLogin"
              >
                登 录
              </el-button>
            </el-form-item>
          </el-form>
        </div>
      </div>
    </div>
    <div class="login-footer">
      <div class="footer-content">
        <span v-if="siteInfo.copyright">{{ siteInfo.copyright }}</span>
        <i v-if="siteInfo.copyright && siteInfo.icp" class="footer-dot" />
        <span v-if="siteInfo.icp">{{ siteInfo.icp }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Grid, Platform, Check } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { getCaptcha, login } from '@/api/auth'
import { getPublicSiteInfo } from '@/api/settings'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref()
const loading = ref(false)
const captchaImage = ref('')
const captchaId = ref('')

const form = reactive({
  username: '',
  password: '',
  captchaId: '',
  captchaCode: ''
})

const siteInfo = reactive({
  siteName: '门户一体化统一管理后台',
  logo: 'computer.svg',
  icp: '',
  copyright: '机械工业信息中心数智应用处 版权所有'
})

const resolveLogoUrl = (url: string) => {
  return url
}

const loadSiteInfo = async () => {
  try {
    const res: any = await getPublicSiteInfo()
    if (res.data) {
      siteInfo.siteName = res.data.siteName || siteInfo.siteName
      siteInfo.logo = res.data.logo || ''
      siteInfo.icp = res.data.icp || ''
      siteInfo.copyright = res.data.copyright || siteInfo.copyright

      document.title = siteInfo.siteName
      const favicon = document.querySelector('link[rel="icon"]') as HTMLLinkElement | null
      if (favicon && siteInfo.logo) {
        favicon.href = resolveLogoUrl(siteInfo.logo)
      }
    }
  } catch {
    // 使用默认值
  }
}



const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captchaCode: [{ required: true, message: '请输入验证码', trigger: 'blur' }]
}

const refreshCaptcha = async () => {
  try {
    const res: any = await getCaptcha()
    captchaId.value = res.data.captcha_id
    captchaImage.value = res.data.captcha_img
    form.captchaId = res.data.captcha_id
  } catch {
    ElMessage.warning('验证码加载失败，请稍后重试')
  }
}

const handleLogin = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const res: any = await login({
      account: form.username,
      password: form.password,
      captcha_id: captchaId.value,
      captcha_code: form.captchaCode
    })
    userStore.setToken(res.data.token)
    userStore.setSignKey(res.data.signKey)
    userStore.setUserInfo(res.data.user)
    await userStore.fetchUserMenusAndGenerateRoutes()
    ElMessage.success('登录成功')
    router.push('/')
  } catch {
    refreshCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refreshCaptcha()
  loadSiteInfo()
})
</script>

<style scoped lang="scss">
.login-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #eef4ff 0%, #f7f9fc 50%, #eef4ff 100%);
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    width: 640px;
    height: 640px;
    background: radial-gradient(circle, rgba(64, 158, 255, 0.12) 0%, transparent 70%);
    top: -180px;
    right: -120px;
  }

  &::after {
    content: '';
    position: absolute;
    width: 520px;
    height: 520px;
    background: radial-gradient(circle, rgba(99, 102, 241, 0.1) 0%, transparent 70%);
    bottom: -140px;
    left: -120px;
  }
}

.login-box {
  display: flex;
  width: 960px;
  min-height: 560px;
  background: #fff;
  border-radius: 20px;
  box-shadow: 0 20px 60px rgba(16, 24, 40, 0.12);
  overflow: hidden;
  z-index: 1;
}

.login-left {
  position: relative;
  width: 440px;
  background: linear-gradient(160deg, #2f7fe0 0%, #1e5fb4 55%, #16395f 100%);
  padding: 56px 44px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  color: #fff;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    inset: 0;
    background-image: linear-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px),
      linear-gradient(90deg, rgba(255, 255, 255, 0.05) 1px, transparent 1px);
    background-size: 34px 34px;
  }

  &::after {
    content: '';
    position: absolute;
    width: 280px;
    height: 280px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.06);
    bottom: -80px;
    right: -80px;
  }
}

.login-brand {
  position: relative;
  z-index: 1;
  text-align: left;
  margin-bottom: 48px;

  .login-logo {
    width: 72px;
    height: 72px;
    object-fit: contain;
    border-radius: 16px;
    background: rgba(255, 255, 255, 0.15);
    padding: 10px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
  }

  .el-icon {
    width: 72px;
    height: 72px;
    border-radius: 16px;
    background: rgba(255, 255, 255, 0.15);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
  }

  h1 {
    font-size: 26px;
    margin: 20px 0 8px;
    font-weight: 700;
    letter-spacing: 0.5px;
  }

  p {
    font-size: 14px;
    opacity: 0.8;
    letter-spacing: 2px;
  }
}

.login-features {
  position: relative;
  z-index: 1;

  .feature-item {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 18px;
    font-size: 15px;
    opacity: 0.95;

    .el-icon {
      width: 26px;
      height: 26px;
      background: rgba(255, 255, 255, 0.18);
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
    }
  }
}

.login-right {
  flex: 1;
  padding: 56px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.login-form-wrapper {
  width: 100%;
  max-width: 360px;
  margin: 0 auto;

  h2 {
    font-size: 24px;
    color: #1d2739;
    margin-bottom: 8px;
    font-weight: 700;
  }

  .login-subtitle {
    font-size: 14px;
    color: #98a2b3;
    margin-bottom: 32px;
  }
}

.login-form {
  .el-form-item {
    margin-bottom: 22px;
  }
}

.login-form {
  .el-form-item {
    margin-bottom: 22px;
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
  background: #f5f7fb;
  border: 1px solid #e3e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.captcha-placeholder {
  color: #98a2b3;
  font-size: 12px;
}

.login-btn {
  width: 100%;
  height: 46px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 10px;
  background: linear-gradient(135deg, #409eff 0%, #2f7fe0 100%);
  border: none;
  box-shadow: 0 6px 16px rgba(64, 158, 255, 0.35);
  letter-spacing: 4px;

  &:hover {
    background: linear-gradient(135deg, #66b1ff 0%, #409eff 100%);
    box-shadow: 0 8px 20px rgba(64, 158, 255, 0.4);
  }
}

.login-footer {
  margin-top: 32px;
  padding: 12px 24px;
  color: #98a2b3;
  font-size: 12px;
  z-index: 1;
  text-align: center;

  .footer-content {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
    justify-content: center;
  }

  .footer-dot {
    width: 3px;
    height: 3px;
    background: #b9c2d0;
    border-radius: 50%;
    display: inline-block;
  }
}

@media (max-width: 768px) {
  .login-box {
    width: 90%;
    flex-direction: column;
  }

  .login-left {
    width: 100%;
    padding: 30px;
    min-height: 200px;
  }

  .login-right {
    padding: 30px;
  }
}
</style>