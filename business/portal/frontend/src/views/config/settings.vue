<template>
  <div class="page-container">
    <el-card shadow="hover" class="settings-card">
      <template #header>
        <div class="card-header">系统设置</div>
      </template>

      <el-tabs v-model="activeTab">
        <el-tab-pane label="基础设置" name="basic">
          <el-form :model="basicForm" label-width="120px" class="settings-form">
            <el-form-item label="系统名称">
              <el-input v-model="basicForm.siteName" placeholder="请输入系统名称" />
            </el-form-item>
            <el-form-item label="访问域名">
              <el-input v-model="basicForm.siteUrl" placeholder="如: https://www.miic.com.cn" />
            </el-form-item>
            <el-form-item label="系统Logo">
              <el-upload
                class="logo-uploader"
                action="#"
                accept="image/*"
                :show-file-list="false"
                :http-request="handleLogoUpload"
                :before-upload="beforeLogoUpload"
              >
                <img
                  v-if="basicForm.logo && !logoPreviewError"
                  :src="resolveLogoUrl(basicForm.logo)"
                  alt="站点 Logo"
                  class="logo-preview"
                  @error="logoPreviewError = true"
                />
                <el-icon v-else class="logo-icon"><Plus /></el-icon>
              </el-upload>
            </el-form-item>
            <el-form-item label="备案信息">
              <el-input v-model="basicForm.icp" placeholder="请输入备案信息" />
            </el-form-item>
            <el-form-item label="版权信息">
              <el-input v-model="basicForm.copyright" placeholder="请输入版权信息" />
            </el-form-item>
            <el-form-item label="机构名称">
              <el-input v-model="basicForm.orgName" placeholder="请输入机构名称" />
            </el-form-item>
            <el-form-item label="机构编码">
              <el-input v-model="basicForm.orgCode" placeholder="请输入机构编码" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="loading" @click="handleSaveBasic">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="安全设置" name="security">
          <!-- 「Token 有效期（小时）」较长，宽度需 ≥160px 才不会折行 -->
          <el-form :model="securityForm" label-width="160px" class="settings-form">
            <el-form-item label="登录验证码">
              <el-switch v-model="securityForm.captchaEnabled" />
              <span class="form-tip">关闭后登录页不再显示验证码输入框</span>
            </el-form-item>
            <el-form-item label="登录失败锁定">
              <el-switch v-model="securityForm.lockEnabled" />
            </el-form-item>
            <el-form-item label="最大失败次数">
              <el-input-number v-model="securityForm.maxFailCount" :min="3" :max="10" />
            </el-form-item>
            <el-form-item label="锁定时间（分钟）">
              <el-input-number v-model="securityForm.lockDuration" :min="5" :max="60" />
            </el-form-item>
            <el-form-item label="密码最小长度">
              <el-input-number v-model="securityForm.minPasswordLength" :min="6" :max="64" />
            </el-form-item>
            <el-form-item label="Token 有效期（小时）">
              <el-input-number v-model="securityForm.tokenExpire" :min="1" :max="72" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="loading" @click="handleSaveSecurity">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="邮件设置" name="email">
          <el-form :model="emailForm" label-width="120px" class="settings-form">
            <el-form-item label="SMTP 服务器">
              <el-input v-model="emailForm.smtpHost" placeholder="如: smtp.example.com" />
            </el-form-item>
            <el-form-item label="SMTP 端口">
              <el-input v-model="emailForm.smtpPort" placeholder="如: 587" />
            </el-form-item>
            <el-form-item label="发件人邮箱">
              <el-input v-model="emailForm.fromEmail" placeholder="请输入发件人邮箱" />
            </el-form-item>
            <el-form-item label="发件人名称">
              <el-input v-model="emailForm.fromName" placeholder="请输入发件人名称" />
            </el-form-item>
            <el-form-item label="邮箱密码">
              <el-input v-model="emailForm.password" type="password" show-password placeholder="请输入邮箱密码或授权码" />
            </el-form-item>
            <el-form-item label="启用 SSL">
              <el-switch v-model="emailForm.ssl" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="testingEmail" @click="handleTestEmail">测试连接</el-button>
              <el-button type="primary" :loading="loading" @click="handleSaveEmail">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="静态化设置" name="static">
          <!-- 本页签含「静态化程序访问令牌名」等 10 字标签，宽度需 ≥180px 才不会折行 -->
          <el-form :model="staticForm" label-width="180px" class="settings-form">
            <el-form-item label="静态化输出路径">
              <el-input
                v-model="staticForm.staticPath"
                placeholder="请输入静态化输出路径，如 D:/static"
                clearable
              />
            </el-form-item>

            <el-form-item label="静态化程序访问地址">
              <el-input
                v-model="staticForm.staticProgramAddr"
                placeholder="请输入静态化程序访问地址，如 127.0.0.1:8889"
                clearable
              />
            </el-form-item>

            <el-form-item label="静态化程序访问令牌名">
              <el-input
                v-model="staticForm.staticProgramTokenName"
                placeholder="请输入静态化程序访问令牌名，如 CAAM_TOKEN"
                clearable
              />
            </el-form-item>

            <el-form-item label="首页整体变灰">
              <el-switch v-model="staticForm.homeGray" active-text="开启" inactive-text="关闭" />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" :loading="loading" @click="handleSaveStatic">保存设置</el-button>
              <span class="form-tip">此项保存后不会退出登录</span>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useAppStore } from '@/stores/app'
import { useUserStore } from '@/stores/user'
import { useRouter } from 'vue-router'
import { getSettings, updateSettings, testEmailConnection } from '@/api/settings'
import type { Settings } from '@/api/settings'
import { logout as logoutApi } from '@/api/auth'
import { uploadFile } from '@/api/upload'
import { resolveAssetUrl as resolveLogoUrl } from '@/utils/asset'

const appStore = useAppStore()
const userStore = useUserStore()
const testingEmail = ref(false)
const router = useRouter()
const activeTab = ref('basic')
const loading = ref(false)
// Logo 预览加载失败（设置里指向的文件已丢失）时回退为上传占位图，避免裂图
const logoPreviewError = ref(false)

const handleLogoUpload = async (options: any) => {
  try {
    const res: any = await uploadFile(options.file, 'setting')
    if (res.code === 0) {
      basicForm.logo = res.data?.url || res.url || ''
      logoPreviewError.value = false
      ElMessage.success('上传成功')
      options.onSuccess(res)
    }
  } catch (error: any) {
    // 失败提示由 request 拦截器统一给出（code !== 0 会在此处抛出）
    options.onError(error)
  }
}

const beforeLogoUpload = (file: File) => {
  const isImage = file.type.startsWith('image/')
  const isLt2M = file.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error('请上传图片文件')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('图片大小不能超过 2MB')
    return false
  }
  return true
}

const basicForm = reactive({
  siteName: '',
  siteUrl: '',
  logo: '',
  icp: '',
  copyright: '',
  orgName: '',
  orgCode: ''
})

const securityForm = reactive({
  captchaEnabled: true,
  lockEnabled: true,
  maxFailCount: 5,
  lockDuration: 30,
  minPasswordLength: 8,
  tokenExpire: 24
})

const emailForm = reactive({
  smtpHost: '',
  smtpPort: '',
  fromEmail: '',
  fromName: '',
  password: '',
  ssl: true
})

const staticForm = reactive({
  staticPath: '',
  staticProgramAddr: '',
  staticProgramTokenName: '',
  homeGray: false
})

// positiveOr 数值项回退：非正数或非法值统一回落到界面默认值
const positiveOr = (value: unknown, fallback: number) => {
  const num = Number(value)
  return Number.isFinite(num) && num > 0 ? num : fallback
}

const loadSettings = async () => {
  try {
    const res: any = await getSettings()
    const data = res.data as Settings
    basicForm.siteName = data.siteName || ''
    basicForm.siteUrl = data.siteUrl || ''
    basicForm.logo = data.logo || ''
    basicForm.icp = data.icp || ''
    basicForm.copyright = data.copyright || ''
    basicForm.orgName = data.orgName || ''
    basicForm.orgCode = data.orgCode || ''
    securityForm.captchaEnabled = data.captchaEnabled ?? true
    securityForm.lockEnabled = data.lockEnabled ?? true
    // 数值项在非法（<=0）时回退到界面默认值：后端已对保存做区间校验，
    // 若库里是旧脏值（如 0），直接回显会导致本次保存被拦下且用户不知如何修正。
    securityForm.maxFailCount = positiveOr(data.maxFailCount, 5)
    securityForm.lockDuration = positiveOr(data.lockDuration, 30)
    securityForm.minPasswordLength = positiveOr(data.minPasswordLength, 8)
    securityForm.tokenExpire = positiveOr(data.tokenExpire, 24)

    appStore.setSecuritySettings({
      minPasswordLength: securityForm.minPasswordLength
    })
    emailForm.smtpHost = data.smtpHost || ''
    emailForm.smtpPort = data.smtpPort || ''
    emailForm.fromEmail = data.fromEmail || ''
    emailForm.fromName = data.fromName || ''
    emailForm.password = data.emailPassword || ''
    emailForm.ssl = data.ssl ?? true

    staticForm.staticPath = data.staticPath || ''
    staticForm.staticProgramAddr = data.staticProgramAddr || ''
    staticForm.staticProgramTokenName = data.staticProgramTokenName || ''
    staticForm.homeGray = data.homeGray ?? false
  } catch {
    // 失败提示由 request 拦截器统一给出
  }
}

// 只提交当前页签的字段：原实现无论保存哪个页签都会提交「基础+安全+邮件」全量字段
// （等于用页面里的旧值覆盖其它页签、甚至覆盖他人刚改的配置），并且一律强制重新登录。
const doSave = async (data: Partial<Settings>) => {
  loading.value = true
  try {
    await updateSettings(data)
    ElMessage.success('保存成功')
  } catch {
    // 失败原因由 request 拦截器统一提示
  } finally {
    loading.value = false
  }
}

// 安全设置（登录验证码/失败锁定/密码长度/Token 有效期）属登录态相关配置：保存后重新登录，
// 让本次会话立即按新策略生效（登出前先把后端 Token 拉黑）。
const doSaveWithRelogin = async (data: Partial<Settings>) => {
  loading.value = true
  try {
    await updateSettings(data)
    ElMessage.success('保存成功，请重新登录')
    try {
      await logoutApi()
    } catch {
      // 忽略：即使后端登出失败也要清理本地会话
    }
    userStore.logout()
    router.push('/login')
  } catch {
    // 失败原因由 request 拦截器统一提示
  } finally {
    loading.value = false
  }
}

const handleSaveBasic = () => {
  doSave({
    siteName: basicForm.siteName,
    siteUrl: basicForm.siteUrl,
    logo: basicForm.logo,
    icp: basicForm.icp,
    copyright: basicForm.copyright,
    orgName: basicForm.orgName,
    orgCode: basicForm.orgCode
  })
}

const handleSaveSecurity = async () => {
  await doSaveWithRelogin({
    captchaEnabled: securityForm.captchaEnabled,
    lockEnabled: securityForm.lockEnabled,
    maxFailCount: securityForm.maxFailCount,
    lockDuration: securityForm.lockDuration,
    minPasswordLength: securityForm.minPasswordLength,
    tokenExpire: securityForm.tokenExpire
  })
  appStore.setSecuritySettings({
    minPasswordLength: securityForm.minPasswordLength
  })
}

const handleSaveEmail = () => {
  doSave({
    smtpHost: emailForm.smtpHost,
    smtpPort: emailForm.smtpPort,
    fromEmail: emailForm.fromEmail,
    fromName: emailForm.fromName,
    emailPassword: emailForm.password,
    ssl: emailForm.ssl
  })
}

const handleTestEmail = async () => {
  testingEmail.value = true
  try {
    await testEmailConnection()
    ElMessage.success('邮件连接测试成功')
  } catch {
    // 失败原因由 request 拦截器统一提示
  } finally {
    testingEmail.value = false
  }
}

// 静态化参数不影响当前登录态，保存后无需强制重新登录
const handleSaveStatic = async () => {
  loading.value = true
  try {
    await updateSettings({
      staticPath: staticForm.staticPath,
      staticProgramAddr: staticForm.staticProgramAddr,
      staticProgramTokenName: staticForm.staticProgramTokenName,
      homeGray: staticForm.homeGray
    })
    ElMessage.success('保存成功')
  } catch {
    // 失败原因由 request 拦截器统一提示
  } finally {
    loading.value = false
  }
}

watch(() => appStore.refreshKey, () => {
  loadSettings()
})

onMounted(() => {
  loadSettings()
})
</script>

<style scoped lang="scss">
.page-container {
  .settings-card {
    border-radius: var(--app-card-radius);
    border: 1px solid var(--app-brand-soft);

    .card-header {
      font-weight: 600;
      color: var(--app-text-heading);
    }
  }

  .settings-form {
    max-width: 600px;
    padding: 20px 0;

    .form-tip {
      margin-left: 12px;
      color: #909399;
      font-size: 12px;
    }
  }

  .logo-uploader {
    :deep(.el-upload) {
      border: 1px dashed var(--el-color-primary-light-8);
      border-radius: 6px;
      cursor: pointer;
      position: relative;
      overflow: hidden;
      transition: var(--el-transition-duration-fast);

      &:hover {
        border-color: var(--el-color-primary);
      }
    }

    .logo-preview {
      width: 120px;
      height: 120px;
      object-fit: contain;
    }

    .logo-icon {
      font-size: 28px;
      color: #8c939d;
      width: 120px;
      height: 120px;
      text-align: center;
      line-height: 120px;
    }
  }
}
</style>
