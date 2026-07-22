<template>
  <div class="setting-page">
    <el-card>
      <template #header>
        <span>系统设置</span>
      </template>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="基础配置" name="basic">
          <el-form label-width="120px" style="max-width: 600px">
            <el-form-item label="平台名称">
              <el-input v-model="settings.basic.platformName" />
            </el-form-item>
            <el-form-item label="Logo">
              <el-input v-model="settings.basic.logo" />
            </el-form-item>
            <el-form-item label="版权信息">
              <el-input v-model="settings.basic.copyright" type="textarea" :rows="3" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleSave">保存</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="安全策略" name="security">
          <el-form label-width="120px" style="max-width: 600px">
            <el-form-item label="登录失败锁定">
              <el-switch v-model="settings.security.loginLock" />
            </el-form-item>
            <el-form-item label="密码最小长度">
              <el-input-number v-model="settings.security.pwdMinLength" :min="6" :max="32" />
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="邮件配置" name="email">
          <el-form label-width="120px" style="max-width: 600px">
            <el-form-item label="SMTP 服务器">
              <el-input v-model="settings.email.host" placeholder="如 smtp.example.com" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input-number v-model="settings.email.port" :min="1" :max="65535" />
            </el-form-item>
            <el-form-item label="用户名">
              <el-input v-model="settings.email.username" />
            </el-form-item>
            <el-form-item label="密码/授权码">
              <el-input v-model="settings.email.password" type="password" show-password />
            </el-form-item>
            <el-form-item label="发件人">
              <el-input v-model="settings.email.from" placeholder="如 noreply@example.com" />
            </el-form-item>
            <el-form-item label="启用 SSL">
              <el-switch v-model="settings.email.ssl" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleSave">保存</el-button>
            </el-form-item>
            <el-divider />
            <el-form-item label="测试收件人">
              <el-input v-model="testEmail.to" placeholder="请输入测试邮箱地址" />
            </el-form-item>
            <el-form-item>
              <el-button type="success" :loading="sending" @click="handleTestEmail">发送测试邮件</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSettings, saveSettings } from '@/api/setting'
import { testEmail as testEmailApi } from '@/api/setting'
import type { SettingItem } from '@/api/setting'

const activeTab = ref('basic')
const sending = ref(false)

const settings = reactive({
  basic: {
    platformName: 'Base 底座平台',
    logo: '',
    copyright: '© 2026 Base Platform'
  },
  security: {
    loginLock: true,
    pwdMinLength: 8
  },
  email: {
    host: '',
    port: 465,
    username: '',
    password: '',
    from: '',
    ssl: true
  }
})

const testEmail = reactive({
  to: ''
})

const parseEmailSettings = (data: any) => {
  return {
    host: data?.host || '',
    port: Number(data?.port || 465),
    username: data?.username || '',
    password: data?.password || '',
    from: data?.from || '',
    ssl: data?.ssl === 'true'
  }
}

const loadSettings = async () => {
  try {
    const res: any = await getSettings()
    const data = res.data || {}
    if (data.basic) {
      Object.assign(settings.basic, data.basic)
    }
    if (data.security) {
      Object.assign(settings.security, {
        loginLock: data.security.loginLock === 'true',
        pwdMinLength: Number(data.security.pwdMinLength || 8)
      })
    }
    if (data.email) {
      Object.assign(settings.email, parseEmailSettings(data.email))
    }
  } catch (error) {
    ElMessage.error('加载设置失败')
  }
}

const handleSave = async () => {
  const payload: SettingItem[] = [
    { category: 'basic', key: 'platformName', value: settings.basic.platformName, type: 'string' },
    { category: 'basic', key: 'logo', value: settings.basic.logo, type: 'string' },
    { category: 'basic', key: 'copyright', value: settings.basic.copyright, type: 'string' },
    { category: 'security', key: 'loginLock', value: String(settings.security.loginLock), type: 'boolean' },
    { category: 'security', key: 'pwdMinLength', value: String(settings.security.pwdMinLength), type: 'number' },
    { category: 'email', key: 'host', value: settings.email.host, type: 'string' },
    { category: 'email', key: 'port', value: String(settings.email.port), type: 'number' },
    { category: 'email', key: 'username', value: settings.email.username, type: 'string' },
    { category: 'email', key: 'password', value: settings.email.password, type: 'string' },
    { category: 'email', key: 'from', value: settings.email.from, type: 'string' },
    { category: 'email', key: 'ssl', value: String(settings.email.ssl), type: 'boolean' }
  ]
  try {
    await saveSettings(payload)
    ElMessage.success('设置已保存')
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const handleTestEmail = async () => {
  if (!testEmail.to) {
    ElMessage.warning('请输入测试收件人邮箱')
    return
  }
  sending.value = true
  try {
    await testEmailApi({
      host: settings.email.host,
      port: settings.email.port,
      username: settings.email.username,
      password: settings.email.password,
      from: settings.email.from,
      ssl: settings.email.ssl,
      to: testEmail.to
    })
    ElMessage.success('测试邮件已发送')
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.message || '发送失败')
  } finally {
    sending.value = false
  }
}

onMounted(loadSettings)
</script>

<style scoped lang="scss">
.setting-page {
  .el-tabs {
    min-height: 400px;
  }
}
</style>
