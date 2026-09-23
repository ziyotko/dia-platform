<template>
  <div class="setting-page">
    <el-card>
      <template #header>
        <span>系统设置</span>
      </template>
      <el-alert
        v-if="!isSuperAdmin"
        title="系统设置为平台级配置，仅平台超级管理员可查看与修改"
        type="warning"
        :closable="false"
        show-icon
        style="margin-bottom: 16px"
      />
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
              <el-button v-if="canWrite('base:setting:save')" type="primary" @click="handleSave">保存</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="安全策略" name="security">
          <el-form label-width="140px" style="max-width: 640px">
            <el-form-item label="登录验证码">
              <el-switch v-model="settings.security.captchaEnabled" />
              <div class="form-tip">关闭后登录页不再显示验证码输入框</div>
            </el-form-item>
            <el-form-item label="登录失败锁定">
              <el-switch v-model="settings.security.loginLock" />
              <div class="form-tip">连续登录失败达到上限后锁定账号一段时间</div>
            </el-form-item>
            <el-form-item label="最大失败次数">
              <el-input-number v-model="settings.security.maxFailCount" :min="3" :max="20" />
            </el-form-item>
            <el-form-item label="锁定时长(分钟)">
              <el-input-number v-model="settings.security.lockDuration" :min="1" :max="1440" />
            </el-form-item>
            <el-form-item label="密码最小长度">
              <el-input-number v-model="settings.security.pwdMinLength" :min="6" :max="32" />
              <div class="form-tip">新增用户与修改密码时服务端校验</div>
            </el-form-item>
            <el-form-item>
              <el-button v-if="canWrite('base:setting:save')" type="primary" @click="handleSaveSecurity">保存</el-button>
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
              <el-button v-if="can('base:setting:save')" type="primary" @click="handleSave">保存</el-button>
            </el-form-item>
            <el-divider />
            <el-form-item label="测试收件人">
              <el-input v-model="testEmail.to" placeholder="请输入测试邮箱地址" />
            </el-form-item>
            <el-form-item>
              <el-button v-if="canWrite('base:setting:test-email')" type="success" :loading="sending" @click="handleTestEmail">发送测试邮件</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="通知渠道" name="notify">
          <el-form label-width="140px" style="max-width: 640px">
            <el-alert
              title="未填写的渠道不会注册，也不会出现在消息发送的渠道下拉中"
              type="info"
              :closable="false"
              show-icon
              style="margin-bottom: 12px"
            />
            <el-form-item label="企业微信机器人">
              <el-input
                v-model="settings.notify.wechatWebhookUrl"
                placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"
              />
              <div class="form-tip">群机器人 webhook：发消息时整条消息推送到对应群</div>
            </el-form-item>
            <el-form-item label="短信网关地址">
              <el-input v-model="settings.notify.smsGatewayUrl" placeholder="https://your-sms-gateway/send" />
              <div class="form-tip">底座按统一约定 POST {to, sign, subject, content}，由网关对接具体短信服务商</div>
            </el-form-item>
            <el-form-item label="短信网关 Token">
              <el-input v-model="settings.notify.smsGatewayToken" type="password" show-password />
            </el-form-item>
            <el-form-item label="短信签名">
              <el-input v-model="settings.notify.smsSign" placeholder="如【Base平台】" />
            </el-form-item>
            <el-form-item>
              <el-button v-if="canWrite('base:setting:save')" type="primary" @click="handleSaveNotify">保存</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSettings, saveSettings } from '@/api/setting'
import { testEmail as testEmailApi } from '@/api/setting'
import type { SettingItem } from '@/api/setting'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
// 按钮级权限：与后端 base:setting:* 权限点对齐
const can = (code: string) => userStore.can(code)
// 系统设置为平台级配置，后端仅允许平台超管访问（/settings 整组挂了 SuperAdminOnly），
// 因此租户管理员虽然 can() 为 true 也不能看到写按钮，否则点击必然 403
const isSuperAdmin = computed(() => userStore.userInfo?.tenantId === 0)
const canWrite = (code: string) => isSuperAdmin.value && can(code)

const activeTab = ref('basic')
const sending = ref(false)

const settings = reactive({
  basic: {
    platformName: 'Base 底座平台',
    logo: '',
    copyright: '© 2026 Base Platform'
  },
  security: {
    captchaEnabled: true,
    loginLock: true,
    maxFailCount: 5,
    lockDuration: 30,
    pwdMinLength: 8
  },
  email: {
    host: '',
    port: 465,
    username: '',
    password: '',
    from: '',
    ssl: true
  },
  notify: {
    wechatWebhookUrl: '',
    smsGatewayUrl: '',
    smsGatewayToken: '',
    smsSign: ''
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
        captchaEnabled: data.security.captchaEnabled !== 'false',
        loginLock: data.security.loginLock !== 'false',
        maxFailCount: Number(data.security.maxFailCount || 5),
        lockDuration: Number(data.security.lockDuration || 30),
        pwdMinLength: Number(data.security.pwdMinLength || 8)
      })
    }
    if (data.email) {
      Object.assign(settings.email, parseEmailSettings(data.email))
    }
    if (data.notify) {
      settings.notify.wechatWebhookUrl = data.notify.wechat_webhook_url || ''
      settings.notify.smsGatewayUrl = data.notify.sms_gateway_url || ''
      settings.notify.smsGatewayToken = data.notify.sms_gateway_token || ''
      settings.notify.smsSign = data.notify.sms_sign || ''
    }
  } catch (error) {
    ElMessage.error('加载设置失败')
    console.error('[base] 加载设置失败', error)
  }
}

const handleSave = async () => {
  // 只提交当前两个页签（基础配置 / 邮件配置）的字段：
  // 安全策略与通知渠道各有独立的保存动作，混在一起会让「保存」的含义变得不确定
  const payload: SettingItem[] = [
    { category: 'basic', key: 'platformName', value: settings.basic.platformName, type: 'string' },
    { category: 'basic', key: 'logo', value: settings.basic.logo, type: 'string' },
    { category: 'basic', key: 'copyright', value: settings.basic.copyright, type: 'string' },
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
    // 失败原因已由 request 拦截器弹出，这里不覆盖
    console.error('[base] 保存设置失败', error)
  }
}

// 安全策略保存（登录验证码开关 + 登录失败锁定 + 密码最小长度）
const handleSaveSecurity = async () => {
  const payload: SettingItem[] = [
    { category: 'security', key: 'captchaEnabled', value: String(settings.security.captchaEnabled), type: 'boolean' },
    { category: 'security', key: 'loginLock', value: String(settings.security.loginLock), type: 'boolean' },
    { category: 'security', key: 'maxFailCount', value: String(settings.security.maxFailCount), type: 'number' },
    { category: 'security', key: 'lockDuration', value: String(settings.security.lockDuration), type: 'number' },
    { category: 'security', key: 'pwdMinLength', value: String(settings.security.pwdMinLength), type: 'number' }
  ]
  try {
    await saveSettings(payload)
    ElMessage.success('安全策略已保存')
  } catch (error) {
    console.error('[base] 保存安全策略失败', error)
  }
}

// 通知渠道保存（企业微信机器人 / 短信网关）
const handleSaveNotify = async () => {
  const payload: SettingItem[] = [
    { category: 'notify', key: 'wechat_webhook_url', value: settings.notify.wechatWebhookUrl, type: 'string' },
    { category: 'notify', key: 'sms_gateway_url', value: settings.notify.smsGatewayUrl, type: 'string' },
    { category: 'notify', key: 'sms_gateway_token', value: settings.notify.smsGatewayToken, type: 'string' },
    { category: 'notify', key: 'sms_sign', value: settings.notify.smsSign, type: 'string' }
  ]
  try {
    await saveSettings(payload)
    ElMessage.success('通知渠道已保存')
  } catch (error) {
    console.error('[base] 保存通知渠道失败', error)
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
    // 统一响应下真实原因在 error.message（错误详情已由拦截器弹出）
    ElMessage.error(error?.message || '发送失败')
  } finally {
    sending.value = false
  }
}

onMounted(() => {
  // 系统设置仅平台超管可读，非超管不请求接口（后端会返回 403）
  if (isSuperAdmin.value) {
    loadSettings()
  }
})
</script>

<style scoped lang="scss">
.setting-page {
  .el-tabs {
    min-height: 400px;
  }

  .form-tip {
    width: 100%;
    font-size: 12px;
    color: #94a3b8;
    line-height: 1.6;
  }
}
</style>
