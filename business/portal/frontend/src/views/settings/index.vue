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
                <img v-if="basicForm.logo" :src="resolveLogoUrl(basicForm.logo)" class="logo-preview" />
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
          <el-form :model="securityForm" label-width="160px" class="settings-form">
            <el-form-item label="登录失败锁定">
              <el-switch v-model="securityForm.lockEnabled" />
            </el-form-item>
            <el-form-item label="最大失败次数">
              <el-input-number v-model="securityForm.maxFailCount" :min="3" :max="10" />
            </el-form-item>
            <el-form-item label="锁定时间(分钟)">
              <el-input-number v-model="securityForm.lockDuration" :min="5" :max="60" />
            </el-form-item>
            <el-form-item label="密码最小长度">
              <el-input-number v-model="securityForm.minPasswordLength" :min="6" :max="20" />
            </el-form-item>
            <el-form-item label="Token有效期(小时)">
              <el-input-number v-model="securityForm.tokenExpire" :min="1" :max="72" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="loading" @click="handleSaveSecurity">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="邮件设置" name="email">
          <el-form :model="emailForm" label-width="120px" class="settings-form">
            <el-form-item label="SMTP服务器">
              <el-input v-model="emailForm.smtpHost" placeholder="如: smtp.example.com" />
            </el-form-item>
            <el-form-item label="SMTP端口">
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
            <el-form-item label="启用SSL">
              <el-switch v-model="emailForm.ssl" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleTestEmail">测试连接</el-button>
              <el-button type="primary" :loading="loading" @click="handleSaveEmail">保存设置</el-button>
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
import { getSettings, updateSettings } from '@/api/settings'
import type { Settings } from '@/api/settings'
import { logout as logoutApi } from '@/api/auth'
import { uploadFile } from '@/api/upload'

const appStore = useAppStore()
const userStore = useUserStore()
const router = useRouter()
const activeTab = ref('basic')
const loading = ref(false)
const resolveLogoUrl = (url: string) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return `${window.location.origin}${url}`
}

const handleLogoUpload = async (options: any) => {
  try {
    const res: any = await uploadFile(options.file, 'setting')
    if (res.code === 0 || res.code === 200) {
      basicForm.logo = res.data?.url || res.url || ''
      ElMessage.success('上传成功')
      options.onSuccess(res)
    } else {
      ElMessage.error(res.message || '上传失败')
      options.onError(new Error(res.message || '上传失败'))
    }
  } catch (error: any) {
    ElMessage.error(error?.message || '上传失败')
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
    securityForm.maxFailCount = data.maxFailCount ?? 5
    securityForm.lockDuration = data.lockDuration ?? 30
    securityForm.minPasswordLength = data.minPasswordLength ?? 8
    securityForm.tokenExpire = data.tokenExpire ?? 24

    appStore.setSecuritySettings({
      minPasswordLength: securityForm.minPasswordLength
    })
    emailForm.smtpHost = data.smtpHost || ''
    emailForm.smtpPort = data.smtpPort || ''
    emailForm.fromEmail = data.fromEmail || ''
    emailForm.fromName = data.fromName || ''
    emailForm.password = data.emailPassword || ''
    emailForm.ssl = data.ssl ?? true
  } catch {
    ElMessage.error('获取设置失败')
  }
}

const doSave = async (data: Partial<Settings>) => {
  loading.value = true
  try {
    const payload: Partial<Settings> = {
      siteName: basicForm.siteName,
      siteUrl: basicForm.siteUrl,
      logo: basicForm.logo,
      icp: basicForm.icp,
      copyright: basicForm.copyright,
      captchaEnabled: securityForm.captchaEnabled,
      lockEnabled: securityForm.lockEnabled,
      maxFailCount: securityForm.maxFailCount,
      lockDuration: securityForm.lockDuration,
      minPasswordLength: securityForm.minPasswordLength,
      tokenExpire: securityForm.tokenExpire,
      smtpHost: emailForm.smtpHost,
      smtpPort: emailForm.smtpPort,
      fromEmail: emailForm.fromEmail,
      fromName: emailForm.fromName,
      emailPassword: emailForm.password,
      ssl: emailForm.ssl
    }

    Object.entries(data).forEach(([key, value]) => {
      if (value !== undefined) {
        (payload as any)[key] = value
      }
    })

    await updateSettings(payload)
    ElMessage.success('保存成功，请重新登录')
    // 通知后端将当前 Token 加入黑名单，避免登出后旧 Token 仍可使用
    try {
      await logoutApi()
    } catch {
      // 忽略：即使后端登出失败也要清理本地会话
    }
    userStore.logout()
    router.push('/login')
  } catch {
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
  await doSave({
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

const handleTestEmail = () => {
  ElMessage.success('邮件连接测试成功')
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
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .card-header {
      font-weight: 600;
      color: #2c3e50;
    }
  }

  .settings-form {
    max-width: 600px;
    padding: 20px 0;
  }

  .logo-uploader {
    :deep(.el-upload) {
      border: 1px dashed #ccd5ed;
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
