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
            <el-form-item label="系统Logo">
              <el-upload
                class="logo-uploader"
                action="#"
                :auto-upload="false"
                :show-file-list="false"
              >
                <img v-if="basicForm.logo" :src="basicForm.logo" class="logo-preview" />
                <el-icon v-else class="logo-icon"><Plus /></el-icon>
              </el-upload>
            </el-form-item>
            <el-form-item label="备案信息">
              <el-input v-model="basicForm.icp" placeholder="请输入备案信息" />
            </el-form-item>
            <el-form-item label="版权信息">
              <el-input v-model="basicForm.copyright" placeholder="请输入版权信息" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleSaveBasic">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="安全设置" name="security">
          <el-form :model="securityForm" label-width="160px" class="settings-form">
            <el-form-item label="登录验证码">
              <el-switch v-model="securityForm.captchaEnabled" />
            </el-form-item>
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
              <el-button type="primary" @click="handleSaveSecurity">保存设置</el-button>
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
              <el-button type="primary" @click="handleSaveEmail">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="主题设置" name="theme">
          <el-form label-width="120px" class="settings-form">
            <el-form-item label="主题色">
              <el-color-picker v-model="themeColor" show-alpha />
            </el-form-item>
            <el-form-item label="侧边栏风格">
              <el-radio-group v-model="sidebarStyle">
                <el-radio-button label="light">浅色</el-radio-button>
                <el-radio-button label="dark">深色</el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="开启标签页">
              <el-switch v-model="tagsView" />
            </el-form-item>
            <el-form-item label="开启面包屑">
              <el-switch v-model="breadcrumb" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleSaveTheme">保存设置</el-button>
              <el-button @click="handleResetTheme">恢复默认</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

const activeTab = ref('basic')

const basicForm = reactive({
  siteName: '门户网站管理后台',
  logo: '',
  icp: '京ICP备12345678号',
  copyright: ' 门户网站管理系统 版权所有'
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
  smtpHost: 'smtp.example.com',
  smtpPort: '587',
  fromEmail: 'noreply@example.com',
  fromName: '系统通知',
  password: '',
  ssl: true
})

const themeColor = ref('#409eff')
const sidebarStyle = ref('light')
const tagsView = ref(true)
const breadcrumb = ref(true)

const handleSaveBasic = () => {
  ElMessage.success('基础设置已保存')
}

const handleSaveSecurity = () => {
  ElMessage.success('安全设置已保存')
}

const handleSaveEmail = () => {
  ElMessage.success('邮件设置已保存')
}

const handleTestEmail = () => {
  ElMessage.success('邮件连接测试成功')
}

const handleSaveTheme = () => {
  ElMessage.success('主题设置已保存')
}

const handleResetTheme = () => {
  themeColor.value = '#409eff'
  sidebarStyle.value = 'light'
  tagsView.value = true
  breadcrumb.value = true
  ElMessage.success('已恢复默认主题')
}
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
      border: 1px dashed #d9ecff;
      border-radius: 6px;
      cursor: pointer;
      position: relative;
      overflow: hidden;
      transition: var(--el-transition-duration-fast);

      &:hover {
        border-color: #409eff;
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