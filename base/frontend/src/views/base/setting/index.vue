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
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSettings, saveSettings } from '@/api/setting'
import type { SettingItem } from '@/api/setting'

const activeTab = ref('basic')

const settings = reactive({
  basic: {
    platformName: 'Base 底座平台',
    logo: '',
    copyright: '© 2026 Base Platform'
  },
  security: {
    loginLock: true,
    pwdMinLength: 8
  }
})

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
    { category: 'security', key: 'pwdMinLength', value: String(settings.security.pwdMinLength), type: 'number' }
  ]
  try {
    await saveSettings(payload)
    ElMessage.success('设置已保存')
  } catch (error) {
    ElMessage.error('保存失败')
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
