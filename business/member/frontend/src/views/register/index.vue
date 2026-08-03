<template>
  <div class="register-page">
    <!-- 装饰背景元素 -->
    <div class="bg-decoration">
      <div class="bg-circle bg-circle-1"></div>
      <div class="bg-circle bg-circle-2"></div>
      <div class="bg-circle bg-circle-3"></div>
    </div>

    <div class="register-card">
      <div class="card-header">
        <div class="header-icon">
          <el-icon :size="32"><UserFilled /></el-icon>
        </div>
        <h2>注册会员</h2>
        <p class="header-desc">加入我们，享受更多会员服务</p>
        <el-steps :active="step" simple align-center class="custom-steps">
          <el-step title="账户注册">
            <template #icon><el-icon><EditPen /></el-icon></template>
          </el-step>
          <el-step v-if="form1.memberType === 'unit'" title="填写资料">
            <template #icon><el-icon><Document /></el-icon></template>
          </el-step>
          <el-step title="确认提交">
            <template #icon><el-icon><CircleCheck /></el-icon></template>
          </el-step>
        </el-steps>
      </div>

      <!-- Step 1: Account -->
      <!-- Step 2: Profile (unit only) -->
      <!-- Step 3: Confirm + Captcha + Submit -->
      <transition name="fade-slide" mode="out-in">
        <el-form v-if="step === 0" key="step0" ref="form1Ref" :model="form1" :rules="rules1" size="large" label-width="0" class="register-form">
          <div class="form-section-title">账户信息</div>
          <el-form-item prop="username">
            <el-input v-model="form1.username" maxlength="20" placeholder="请输入用户名" :prefix-icon="User" />
          </el-form-item>
          <el-form-item prop="password">
            <el-input v-model="form1.password" type="password" maxlength="20" placeholder="密码（至少8位）" show-password :prefix-icon="Lock" />
          </el-form-item>
          <el-form-item prop="confirmPwd">
            <el-input v-model="form1.confirmPwd" type="password" maxlength="20" placeholder="再次输入密码" show-password :prefix-icon="Lock" />
          </el-form-item>

          <div class="form-section-title">联系方式</div>
          <el-form-item prop="mobile">
            <el-input v-model="form1.mobile" maxlength="11" placeholder="请输入手机号" :prefix-icon="Phone" />
          </el-form-item>
          <el-form-item prop="email">
            <el-input v-model="form1.email" maxlength="40" placeholder="请输入邮箱" :prefix-icon="Message" />
          </el-form-item>

          <div class="form-section-title">会员类型</div>
          <el-form-item>
            <el-radio-group v-model="form1.memberType" class="member-type-group">
              <el-radio value="unit" class="type-card">
                <div class="type-card-content">
                  <el-icon :size="24"><OfficeBuilding /></el-icon>
                  <span>单位会员</span>
                </div>
              </el-radio>
              <el-radio value="personal" class="type-card">
                <div class="type-card-content">
                  <el-icon :size="24"><User /></el-icon>
                  <span>个人会员</span>
                </div>
              </el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item class="form-actions">
            <el-button type="primary" size="large" class="btn-primary" @click="nextStep">
              下一步
              <el-icon class="btn-icon"><ArrowRight /></el-icon>
            </el-button>
          </el-form-item>
        </el-form>

        <el-form v-else-if="step === 1 && form1.memberType === 'unit'" key="step1" ref="form2Ref" :model="form2" :rules="rules2" size="large" label-width="0" class="register-form">
          <div class="form-section-title">单位信息</div>
          <el-form-item prop="companyName">
            <el-input v-model="form2.companyName" placeholder="请输入单位全称" :prefix-icon="OfficeBuilding" />
          </el-form-item>
          <el-form-item prop="creditCode">
            <el-input v-model="form2.creditCode" maxlength="18" placeholder="统一社会信用代码" :prefix-icon="Document" />
          </el-form-item>
          <el-form-item prop="legalPerson">
            <el-input v-model="form2.legalPerson" placeholder="法定代表人姓名" :prefix-icon="User" />
          </el-form-item>
          <el-form-item>
            <el-input v-model="form2.contactPerson" placeholder="联系人姓名（选填）" :prefix-icon="UserFilled" />
          </el-form-item>
          <el-form-item>
            <el-input v-model="form2.address" placeholder="单位详细地址（选填）" :prefix-icon="Location" />
          </el-form-item>

          <div class="form-section-title">资质文件</div>
          <el-form-item prop="certFile">
            <div class="upload-wrapper">
              <el-upload
                ref="uploadRef"
                :auto-upload="false"
                :show-file-list="true"
                :limit="1"
                :on-change="handleCertChange"
                :on-exceed="() => ElMessage.warning('只能上传一个文件')"
                accept=".jpg,.jpeg,.png,.pdf"
                class="cert-upload"
              >
                <div v-if="!certFileName" class="upload-placeholder">
                  <el-icon :size="32"><UploadFilled /></el-icon>
                  <span>点击上传组织机构证</span>
                  <span class="upload-hint">支持 jpg/png/pdf 格式</span>
                </div>
                <div v-else class="upload-done">
                  <el-icon :size="24" color="#67c23a"><CircleCheckFilled /></el-icon>
                  <span>{{ certFileName }}</span>
                </div>
              </el-upload>
              <div v-if="uploading" class="upload-progress">
                <el-progress :percentage="uploadProgress" :stroke-width="6" />
              </div>
            </div>
          </el-form-item>

          <el-form-item class="form-actions">
            <el-button size="large" class="btn-secondary" @click="step = 0">
              <el-icon class="btn-icon"><ArrowLeft /></el-icon>
              上一步
            </el-button>
            <el-button type="primary" size="large" class="btn-primary" @click="nextStep">
              下一步
              <el-icon class="btn-icon"><ArrowRight /></el-icon>
            </el-button>
          </el-form-item>
        </el-form>

        <div v-else key="stepConfirm" class="confirm-step">
          <div class="confirm-header">
            <el-icon :size="28" color="#409eff"><CircleCheckFilled /></el-icon>
            <span>{{ form1.memberType === 'unit' ? '请确认入会信息' : '确认注册信息' }}</span>
          </div>

          <div class="info-cards">
            <div class="info-card">
              <div class="info-card-title">账户信息</div>
              <div class="info-grid">
                <div class="info-item">
                  <span class="info-label">用户名</span>
                  <span class="info-value">{{ form1.username }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">手机号</span>
                  <span class="info-value">{{ form1.mobile }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">邮箱</span>
                  <span class="info-value">{{ form1.email || '未填写' }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">会员类型</span>
                  <span class="info-value">
                    <el-tag :type="form1.memberType === 'unit' ? 'warning' : 'success'" effect="plain" round>
                      {{ form1.memberType === 'unit' ? '单位会员' : '个人会员' }}
                    </el-tag>
                  </span>
                </div>
                <div class="info-item">
                  <span class="info-label">会员等级</span>
                  <span class="info-value">
                    <el-tag type="info" effect="plain" round>暂无</el-tag>
                  </span>
                </div>
              </div>
            </div>

            <div v-if="form1.memberType === 'unit'" class="info-card">
              <div class="info-card-title">单位信息</div>
              <div class="info-grid">
                <div class="info-item">
                  <span class="info-label">单位名称</span>
                  <span class="info-value">{{ form2.companyName }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">信用代码</span>
                  <span class="info-value">{{ form2.creditCode || '未填写' }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">法定代表人</span>
                  <span class="info-value">{{ form2.legalPerson || '未填写' }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">联系人</span>
                  <span class="info-value">{{ form2.contactPerson || '未填写' }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">单位地址</span>
                  <span class="info-value">{{ form2.address || '未填写' }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">组织机构证</span>
                  <span class="info-value">{{ certFileName || '未上传' }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="captcha-section">
            <div class="captcha-title">
              <el-icon :size="18" color="#909399"><Key /></el-icon>
              <span>安全验证</span>
            </div>
            <el-form ref="captchaFormRef" :model="captchaForm" :rules="captchaRules" size="large" class="captcha-form">
              <el-form-item prop="code">
                <div class="captcha-row">
                  <el-input v-model="captchaForm.code" placeholder="请输入验证码" :prefix-icon="Key" />
                  <img :src="captchaImage" class="captcha-img" @click="loadCaptcha" title="点击刷新验证码" />
                </div>
              </el-form-item>
            </el-form>
          </div>

          <div class="confirm-actions">
            <el-button size="large" class="btn-secondary" @click="step = form1.memberType === 'unit' ? 1 : 0">
              <el-icon class="btn-icon"><ArrowLeft /></el-icon>
              上一步
            </el-button>
            <el-button type="primary" size="large" class="btn-primary btn-submit" :loading="submitting" @click="handleSubmit">
              <el-icon class="btn-icon" v-if="!submitting"><CircleCheck /></el-icon>
              {{ submitting ? '提交中...' : '确认提交' }}
            </el-button>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { authApi } from '@/api/auth'
import { ElMessage } from 'element-plus'
import {
  User, Lock, Phone, Message, ArrowRight, ArrowLeft,
  EditPen, Document, CircleCheck, CircleCheckFilled,
  OfficeBuilding, Location, Key, UserFilled, UploadFilled
} from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()
const step = ref(0)
const submitting = ref(false)
const uploading = ref(false)
const uploadProgress = ref(0)
const captchaImage = ref('')
const captchaId = ref('')
const certFileUrl = ref('')
const certFileName = ref('')
const uploadRef = ref()
const form1Ref = ref()
const form2Ref = ref()
const captchaFormRef = ref()

const form1 = reactive({
  username: '', password: '', confirmPwd: '', mobile: '', email: '',
  memberType: 'unit'
})

const captchaForm = reactive({ code: '' })
const captchaRules = { code: [{ required: true, message: '请输入验证码' }] }

const validatePass = (_rule: any, value: string, callback: any) => {
  if (!value) callback(new Error('请再次输入密码'))
  else if (value !== form1.password) callback(new Error('两次密码不一致'))
  else callback()
}

// 防抖：避免输入过程中频繁请求后端查重
function debounce<A extends any[]>(fn: (...args: A) => void, delay = 400) {
  let timer: ReturnType<typeof setTimeout> | null = null
  return (...args: A) => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }
}

// 密码复杂度：至少8位，且包含大小写字母、数字、特殊字符
const validatePassword = (_rule: any, value: string, callback: any) => {
  if (!value) return callback(new Error('请输入密码'))
  if (value.length > 20) return callback(new Error('密码不能超过20位'))
  if (value.length < 8) return callback(new Error('密码至少8位'))
  const checks: Array<[RegExp, string]> = [
    [/[a-z]/, '密码需包含小写字母'],
    [/[A-Z]/, '密码需包含大写字母'],
    [/[0-9]/, '密码需包含数字'],
    [/[^A-Za-z0-9]/, '密码需包含特殊字符']
  ]
  for (const [re, msg] of checks) {
    if (!re.test(value)) return callback(new Error(msg))
  }
  callback()
}

// 用户名：只能小写字母、数字、-、_，且查重
const checkUsernameAvailable = debounce(async (value: string, callback: any) => {
  if (!value) return callback()
  try {
    const res = await authApi.checkExists({ field: 'username', value })
    if (res.data.exists) callback(new Error('该用户名已被注册'))
    else callback()
  } catch {
    callback()
  }
})

const validateUsername = (_rule: any, value: string, callback: any) => {
  if (!value) return callback(new Error('请输入用户名'))
  if (value.length > 20) return callback(new Error('用户名不能超过20位'))
  if (!/^[a-z0-9_-]+$/.test(value)) {
    return callback(new Error('用户名只能包含小写字母、数字、- 和 _'))
  }
  checkUsernameAvailable(value, callback)
}

// 手机号：校验合法性并查重
const checkMobileAvailable = debounce(async (value: string, callback: any) => {
  if (!value) return callback()
  try {
    const res = await authApi.checkExists({ field: 'mobile', value })
    if (res.data.exists) callback(new Error('该手机号已注册'))
    else callback()
  } catch {
    callback()
  }
})

const validateMobile = (_rule: any, value: string, callback: any) => {
  if (!value) return callback(new Error('请输入手机号'))
  if (value.length > 11) return callback(new Error('手机号不能超过11位'))
  if (!/^1[3-9]\d{9}$/.test(value)) {
    return callback(new Error('请输入合法的手机号'))
  }
  checkMobileAvailable(value, callback)
}

// 邮箱：校验合法性并查重
const checkEmailAvailable = debounce(async (value: string, callback: any) => {
  if (!value) return callback()
  try {
    const res = await authApi.checkExists({ field: 'email', value })
    if (res.data.exists) callback(new Error('该邮箱已注册'))
    else callback()
  } catch {
    callback()
  }
})

const validateEmail = (_rule: any, value: string, callback: any) => {
  if (!value) return callback(new Error('请输入邮箱'))
  if (value.length > 40) return callback(new Error('邮箱不能超过40位'))
  if (!/^[\w.+-]+@[\w-]+(\.[\w-]+)+$/.test(value)) {
    return callback(new Error('邮箱格式不正确'))
  }
  checkEmailAvailable(value, callback)
}

const rules1 = {
  username: [{ validator: validateUsername, trigger: 'blur' }],
  password: [{ validator: validatePassword, trigger: 'blur' }],
  confirmPwd: [{ required: true, validator: validatePass, trigger: 'blur' }],
  mobile: [{ validator: validateMobile, trigger: 'blur' }],
  email: [{ validator: validateEmail, trigger: 'blur' }]
}

// 统一社会信用代码合法性校验（GB 32100-2015：18位，含校验码算法）
const validateCreditCode = (_rule: any, value: string, callback: any) => {
  if (!value) return callback(new Error('请输入统一社会信用代码'))
  const code = value.trim().toUpperCase()
  // 字符集：数字0-9 + 大写字母（去掉 I、O、S、V、Z）
  const charSet = '0123456789ABCDEFGHJKLMNPQRTUWXY'
  // 前17位的加权因子
  const weights = [1, 3, 9, 27, 19, 26, 16, 17, 20, 29, 25, 13, 8, 24, 10, 30, 28]
  if (!/^[0-9A-HJ-NPQRTUWXY]{2}[0-9]{6}[0-9A-HJ-NPQRTUWXY]{10}$/.test(code)) {
    return callback(new Error('统一社会信用代码格式不正确'))
  }
  let sum = 0
  for (let i = 0; i < 17; i++) {
    sum += charSet.indexOf(code[i]) * weights[i]
  }
  const check = (31 - (sum % 31)) % 31
  if (charSet[check] !== code[17]) {
    return callback(new Error('统一社会信用代码校验不通过'))
  }
  callback()
}

const rules2 = {
  companyName: [{ required: true, message: '请输入单位名称' }],
  creditCode: [{ validator: validateCreditCode, trigger: 'blur' }],
  legalPerson: [{ required: true, message: '请输入法定代表人姓名' }]
}

const form2 = reactive({
  companyName: '', creditCode: '', legalPerson: '',
  contactPerson: '', address: ''
})

onMounted(async () => {
  loadCaptcha()
})

async function handleCertChange(uploadFile: any) {
  if (!uploadFile.raw) return
  uploading.value = true
  uploadProgress.value = 0

  // Simulate progress
  const timer = setInterval(() => {
    if (uploadProgress.value < 90) uploadProgress.value += 10
  }, 200)

  try {
    const formData = new FormData()
    formData.append('file', uploadFile.raw)
    formData.append('dir', 'certs')
    const res = await authApi.upload(formData)
    certFileUrl.value = res.data.url
    certFileName.value = uploadFile.raw.name
    uploadProgress.value = 100
  } catch {
    ElMessage.error('文件上传失败，请重试')
    certFileName.value = ''
    certFileUrl.value = ''
  } finally {
    clearInterval(timer)
    uploading.value = false
  }
}

async function loadCaptcha() {
  try {
    const res = await authApi.getCaptcha()
    captchaId.value = res.data.captcha_id
    captchaImage.value = res.data.captcha_image
  } catch {}
}

async function nextStep() {
  if (step.value === 0) {
    const valid = await form1Ref.value?.validate().catch(() => false)
    if (!valid) return
    step.value = form1.memberType === 'personal' ? 2 : 1
  } else if (step.value === 1 && form1.memberType === 'unit') {
    const valid = await form2Ref.value?.validate().catch(() => false)
    if (!valid) return
    step.value = 2
  }
}

async function handleSubmit() {
  // Validate captcha first
  const valid = await captchaFormRef.value?.validate().catch(() => false)
  if (!valid) { loadCaptcha(); return }

  submitting.value = true
  try {
    const registerData: any = {
      username: form1.username,
      password: form1.password,
      mobile: form1.mobile,
      email: form1.email,
      member_type: form1.memberType,
      captcha_id: captchaId.value,
      captcha_code: captchaForm.code
    }
    // For unit members, also send company info with registration
    if (form1.memberType === 'unit') {
      registerData.company_name = form2.companyName
      registerData.credit_code = form2.creditCode
      registerData.legal_person = form2.legalPerson
      registerData.contact_person = form2.contactPerson
      registerData.address = form2.address
      registerData.cert_file = certFileUrl.value
    }

    const res = await authApi.register(registerData)

    // 先清空本地缓存登录信息
    localStorage.clear()
    sessionStorage.clear()

    // 再自动登录系统
    const { token } = res.data
    if (token) {
      userStore.setToken(token)
      userStore.userInfo = res.data.member
    }

    ElMessage.success(form1.memberType === 'unit' ? '入会申请提交成功！' : '注册成功！')
    // 等待 store 状态更新完成再跳转
    await nextTick()
    router.push('/member/dashboard')
  } catch {
    loadCaptcha()
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped lang="scss">
/* ===== 页面背景 ===== */
.register-page {
  min-height: calc(100vh - 180px);
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 40px 20px;
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #e8f0fe 0%, #f0f4ff 40%, #fce4ec 100%);
}

/* ===== 装饰背景元素 ===== */
.bg-decoration {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}
.bg-circle {
  position: absolute;
  border-radius: 50%;
  opacity: 0.3;
}
.bg-circle-1 {
  width: 500px; height: 500px;
  background: radial-gradient(circle, #409eff44 0%, transparent 70%);
  top: -150px; left: -100px;
  animation: float 8s ease-in-out infinite;
}
.bg-circle-2 {
  width: 400px; height: 400px;
  background: radial-gradient(circle, #67c23a33 0%, transparent 70%);
  bottom: -100px; right: -80px;
  animation: float 10s ease-in-out infinite reverse;
}
.bg-circle-3 {
  width: 300px; height: 300px;
  background: radial-gradient(circle, #e6a23c33 0%, transparent 70%);
  top: 30%; right: 10%;
  animation: float 12s ease-in-out infinite 2s;
}

@keyframes float {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(-30px) scale(1.05); }
}

/* ===== 主卡片 ===== */
.register-card {
  width: 620px;
  max-width: 100%;
  background: rgba(255,255,255,0.92);
  backdrop-filter: blur(20px);
  border-radius: 24px;
  padding: 44px 48px 40px;
  box-shadow:
    0 8px 40px rgba(0,0,0,0.06),
    0 1px 4px rgba(0,0,0,0.04);
  border: 1px solid rgba(255,255,255,0.7);
  position: relative;
  z-index: 1;
  transition: transform 0.3s ease;

  &:hover {
    transform: translateY(-2px);
  }
}

/* ===== 卡片头部 ===== */
.card-header {
  text-align: center;
  margin-bottom: 32px;

  .header-icon {
    width: 64px;
    height: 64px;
    margin: 0 auto 16px;
    background: linear-gradient(135deg, #409eff, #337ecc);
    border-radius: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    box-shadow: 0 6px 20px rgba(64,158,255,0.3);
  }

  h2 {
    margin: 0 0 6px;
    font-size: 26px;
    font-weight: 700;
    color: #1d2129;
    letter-spacing: 1px;
  }

  .header-desc {
    margin: 0 0 28px;
    font-size: 14px;
    color: #86909c;
  }
}

.custom-steps {
  :deep(.el-step) {
    .el-step__head.is-process,
    .el-step__head.is-finish {
      color: #409eff;
      border-color: #409eff;
    }
    .el-step__title {
      font-size: 13px;
    }
    .el-step__main {
      margin-top: 4px;
    }
  }
}

/* ===== 表单 ===== */
.register-form {
  padding: 0 4px;

  .form-section-title {
    font-size: 14px;
    font-weight: 600;
    color: #4e5969;
    margin-bottom: 12px;
    padding-left: 12px;
    border-left: 3px solid #409eff;
    line-height: 1;

    &:not(:first-child) {
      margin-top: 20px;
    }
  }

  .el-form-item {
    margin-bottom: 18px;
  }

  :deep(.el-input__wrapper) {
    border-radius: 10px;
    padding: 4px 12px;
    box-shadow: 0 0 0 1px #e5e6eb inset;
    transition: all 0.25s ease;
    background: #f7f8fa;

    &:hover {
      box-shadow: 0 0 0 1px #c9cdd4 inset;
      background: #fff;
    }

    &.is-focus {
      box-shadow: 0 0 0 2px #409eff40 inset;
      background: #fff;
    }
  }

  :deep(.el-input__prefix) {
    color: #a9aeb8;
    font-size: 16px;
  }

  :deep(.el-input__inner) {
    height: 42px;
    font-size: 14px;
  }
}

/* ===== 会员类型选择卡片 ===== */
.member-type-group {
  display: flex;
  gap: 16px;
  width: 100%;

  .el-radio {
    flex: 1;
    height: auto;
    margin-right: 0;

    .el-radio__input { display: none; }

    .type-card-content {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      padding: 16px 20px;
      border: 2px solid #e5e6eb;
      border-radius: 14px;
      cursor: pointer;
      transition: all 0.25s ease;
      background: #f7f8fa;
      font-size: 15px;
      font-weight: 500;
      color: #4e5969;
    }

    &.is-checked .type-card-content {
      border-color: #409eff;
      background: #ecf5ff;
      color: #409eff;
      box-shadow: 0 4px 12px rgba(64,158,255,0.15);
    }

    &:hover .type-card-content {
      border-color: #409eff;
      background: #f0f7ff;
    }
  }
}

/* ===== 表单按钮 ===== */
.form-actions {
  margin-top: 28px !important;
  margin-bottom: 0 !important;
  display: flex;
  justify-content: center;
  gap: 16px;
}

.btn-primary, .btn-secondary {
  border-radius: 12px;
  padding: 12px 32px;
  font-weight: 600;
  font-size: 15px;
  transition: all 0.25s ease;
  display: inline-flex;
  align-items: center;
  gap: 6px;

  .btn-icon {
    font-size: 16px;
  }
}

.btn-primary {
  box-shadow: 0 4px 14px rgba(64,158,255,0.3);
  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 6px 20px rgba(64,158,255,0.4);
  }
  &:active { transform: translateY(0); }
}

.btn-secondary {
  &:hover {
    background: #f0f2f5;
    border-color: #c9cdd4;
  }
}

.btn-submit {
  min-width: 160px;
}

/* ===== 确认步骤 ===== */
.confirm-step {
  .confirm-header {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    font-size: 18px;
    font-weight: 600;
    color: #1d2129;
    margin-bottom: 28px;
  }
}

.info-cards {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-bottom: 24px;
}

.info-card {
  background: #f7f8fa;
  border-radius: 16px;
  padding: 20px 24px;
  border: 1px solid #e5e6eb;
  transition: all 0.25s ease;

  &:hover {
    border-color: #c9cdd4;
    box-shadow: 0 2px 8px rgba(0,0,0,0.04);
  }
}

.info-card-title {
  font-size: 14px;
  font-weight: 600;
  color: #4e5969;
  margin-bottom: 16px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e5e6eb;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 24px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.info-label {
  font-size: 12px;
  color: #86909c;
}

.info-value {
  font-size: 14px;
  color: #1d2129;
  font-weight: 500;
}

/* ===== 验证码区域 ===== */
.captcha-section {
  background: linear-gradient(135deg, #f0f5ff, #f5f7fa);
  border-radius: 16px;
  padding: 18px 20px;
  margin-bottom: 24px;
  border: 1px solid #e5e6eb;

  .captcha-title {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    font-weight: 600;
    color: #4e5969;
    margin-bottom: 12px;
  }

  .captcha-form {
    :deep(.el-form-item) {
      margin-bottom: 0;
    }
    :deep(.el-input__wrapper) {
      border-radius: 10px;
      background: #fff;
    }
  }
}

.captcha-row {
  display: flex;
  gap: 12px;
  align-items: center;

  .el-input { flex: 1; }

  .captcha-img {
    width: 130px;
    height: 42px;
    border-radius: 10px;
    cursor: pointer;
    border: 1px solid #e5e6eb;
    transition: all 0.25s ease;
    object-fit: cover;

    &:hover {
      border-color: #409eff;
      box-shadow: 0 2px 8px rgba(64,158,255,0.15);
      transform: scale(1.02);
    }
  }
}

/* ===== 文件上传 ===== */
.upload-wrapper {
  width: 100%;
}

.cert-upload {
  width: 100%;

  :deep(.el-upload) {
    width: 100%;
  }
}

.upload-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 28px 20px;
  border: 2px dashed #d9dde3;
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.25s ease;
  background: #f7f8fa;

  .el-icon { color: #a9aeb8; }

  span { font-size: 14px; color: #4e5969; font-weight: 500; }

  .upload-hint {
    font-size: 12px;
    color: #a9aeb8;
    font-weight: 400;
  }

  &:hover {
    border-color: #409eff;
    background: #ecf5ff;

    .el-icon { color: #409eff; }
    span { color: #409eff; }
  }
}

.upload-done {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 20px;
  border: 2px solid #67c23a;
  border-radius: 14px;
  background: #f0f9eb;
  font-size: 14px;
  font-weight: 500;
  color: #4e5969;
}

.upload-progress {
  margin-top: 10px;
}

/* ===== 底部操作按钮 ===== */
.confirm-actions {
  display: flex;
  justify-content: center;
  gap: 16px;
}

/* ===== 页面切换动画 ===== */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.35s ease;
}
.fade-slide-enter-from {
  opacity: 0;
  transform: translateY(20px);
}
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

/* ===== 响应式 ===== */
@media (max-width: 640px) {
  .register-page { padding: 20px 12px; }
  .register-card { padding: 28px 20px 32px; border-radius: 18px; }
  .info-grid { grid-template-columns: 1fr; }
  .member-type-group { flex-direction: column; }
  .captcha-row { flex-direction: column; align-items: stretch; }
}
</style>
