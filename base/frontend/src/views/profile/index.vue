<template>
  <div class="profile-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="page-title">
        <el-avatar :size="48" class="page-avatar">
          <el-icon :size="24"><UserFilled /></el-icon>
        </el-avatar>
        <div class="page-title-text">
          <h1>个人中心</h1>
          <p>查看并管理您的账户信息</p>
        </div>
      </div>
    </div>

    <el-row :gutter="20">
      <el-col :xs="24" :lg="14">
        <el-card class="info-card" shadow="never">
          <template #header>
            <div class="card-header">
              <el-icon :size="18" color="#2563eb"><User /></el-icon>
              <span>基本信息</span>
            </div>
          </template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="用户名">{{ userStore.userInfo?.username }}</el-descriptions-item>
            <el-descriptions-item label="真实姓名">{{ userStore.userInfo?.realName }}</el-descriptions-item>
            <el-descriptions-item label="手机号">{{ userStore.userInfo?.phone || '-' }}</el-descriptions-item>
            <el-descriptions-item label="邮箱">{{ userStore.userInfo?.email || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="10">
        <el-card class="pwd-card" shadow="never">
          <template #header>
            <div class="card-header">
              <el-icon :size="18" color="#f56c6c"><Lock /></el-icon>
              <span>修改密码</span>
            </div>
          </template>
          <el-form :model="pwdForm" :rules="pwdRules" ref="pwdRef" label-width="90px">
            <el-form-item label="旧密码" prop="oldPwd">
              <el-input v-model="pwdForm.oldPwd" type="password" show-password placeholder="请输入旧密码" />
            </el-form-item>
            <el-form-item label="新密码" prop="newPwd">
              <el-input v-model="pwdForm.newPwd" type="password" show-password placeholder="请输入新密码" />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirmPwd">
              <el-input v-model="pwdForm.confirmPwd" type="password" show-password placeholder="请再次输入新密码" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleChangePwd" class="save-btn">保存修改</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { UserFilled, User, Lock } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { changePassword } from '@/api/auth'

const userStore = useUserStore()
const pwdRef = ref<any>(null)
const pwdForm = reactive({
  oldPwd: '',
  newPwd: '',
  confirmPwd: ''
})

const pwdRules = {
  oldPwd: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  newPwd: [{ required: true, message: '请输入新密码', trigger: 'blur' }],
  confirmPwd: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: Function) => {
        if (value !== pwdForm.newPwd) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

const handleChangePwd = async () => {
  const valid = await pwdRef.value?.validate().catch(() => false)
  if (!valid) return
  await changePassword({ oldPwd: pwdForm.oldPwd, newPwd: pwdForm.newPwd })
  ElMessage.success('密码修改成功，请重新登录')
  userStore.logout()
}
</script>

<style scoped lang="scss">
.profile-page {
  .page-header {
    margin-bottom: 20px;
  }

  .page-title {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .page-avatar {
    background: linear-gradient(135deg, #2563eb 0%, #4f46e5 100%);
    color: #fff;
    box-shadow: 0 8px 20px rgba(37, 99, 235, 0.3);
  }

  .page-title-text {
    h1 {
      margin: 0;
      font-size: 22px;
      font-weight: 700;
      color: #1e293b;
    }

    p {
      margin: 4px 0 0;
      font-size: 14px;
      color: #64748b;
    }
  }

  .info-card,
  .pwd-card {
    border-radius: 14px;
    border: none;
    margin-bottom: 20px;

    :deep(.el-card__header) {
      padding: 18px 24px;
      border-bottom: 1px solid #f1f5f9;
    }

    :deep(.el-card__body) {
      padding: 24px;
    }
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 16px;
    font-weight: 600;
    color: #1e293b;
  }

  .save-btn {
    width: 100%;
    border-radius: 8px;
    background: linear-gradient(135deg, #2563eb 0%, #4f46e5 100%);
    border: none;
  }
}
</style>
