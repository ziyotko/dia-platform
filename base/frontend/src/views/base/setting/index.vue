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
              <el-input v-model="settings.platformName" />
            </el-form-item>
            <el-form-item label="Logo">
              <el-input v-model="settings.logo" />
            </el-form-item>
            <el-form-item label="版权信息">
              <el-input v-model="settings.copyright" type="textarea" :rows="3" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleSave">保存</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="安全策略" name="security">
          <el-form label-width="120px" style="max-width: 600px">
            <el-form-item label="登录失败锁定">
              <el-switch v-model="settings.loginLock" />
            </el-form-item>
            <el-form-item label="密码最小长度">
              <el-input-number v-model="settings.pwdMinLength" :min="6" :max="32" />
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'

const activeTab = ref('basic')
const settings = reactive({
  platformName: 'Base 底座平台',
  logo: '',
  copyright: '© 2026 Base Platform',
  loginLock: true,
  pwdMinLength: 8
})

const handleSave = () => {
  ElMessage.success('设置已保存')
}
</script>

<style scoped lang="scss">
.setting-page {
  .el-tabs {
    min-height: 400px;
  }
}
</style>
