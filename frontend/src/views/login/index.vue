<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-left">
        <div class="login-brand">
          <el-icon size="48" color="#409eff"><Platform /></el-icon>
          <h1>门户网站管理后台</h1>
          <p>Portal Management System</p>
        </div>
        <div class="login-features">
          <div class="feature-item">
            <el-icon><Check /></el-icon>
            <span>统一内容管理</span>
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
          <h2>账号登录</h2>
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
      <p> 门户网站管理系统 版权所有</p>
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

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captchaCode: [{ required: true, message: '请输入验证码', trigger: 'blur' }]
}

const refreshCaptcha = async () => {
  try {
    const res: any = await getCaptcha()
    captchaId.value = res.data.captchaId
    captchaImage.value = res.data.imageBase64
    form.captchaId = res.data.captchaId
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
      username: form.username,
      password: form.password,
      captchaId: captchaId.value,
      captchaCode: form.captchaCode
    })
    userStore.setToken(res.data.token)
    userStore.setUserInfo(res.data.user)
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
})
</script>

<style scoped lang="scss">
.login-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #e6f2ff 0%, #f0f7ff 50%, #e6f2ff 100%);
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    width: 800px;
    height: 800px;
    background: radial-gradient(circle, rgba(64, 158, 255, 0.08) 0%, transparent 70%);
    top: -200px;
    left: -200px;
  }

  &::after {
    content: '';
    position: absolute;
    width: 600px;
    height: 600px;
    background: radial-gradient(circle, rgba(64, 158, 255, 0.06) 0%, transparent 70%);
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
  box-shadow: 0 8px 32px rgba(64, 158, 255, 0.15);
  overflow: hidden;
  z-index: 1;
}

.login-left {
  width: 380px;
  background: linear-gradient(160deg, #409eff 0%, #337ecc 100%);
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
  .el-input {
    :deep(.el-input__wrapper) {
      box-shadow: 0 0 0 1px #d9ecff inset;
      background: #f5f9ff;

      &.is-focus {
        box-shadow: 0 0 0 1px #409eff inset;
      }
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
  border-radius: 4px;
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
  background: linear-gradient(135deg, #409eff 0%, #337ecc 100%);
  border: none;

  &:hover {
    background: linear-gradient(135deg, #66b1ff 0%, #409eff 100%);
  }
}

.login-footer {
  margin-top: 24px;
  color: #909399;
  font-size: 13px;
  z-index: 1;
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