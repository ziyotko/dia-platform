<template>
  <div class="reset-page">
    <div class="reset-card">
      <h2>重置密码</h2>
      <el-form :model="form" size="large">
        <el-form-item>
          <el-input v-model="form.email" placeholder="注册邮箱" prefix-icon="Message" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="full-btn" @click="sendEmail">发送重置链接</el-button>
        </el-form-item>
        <el-divider />
        <el-form-item>
          <el-input v-model="form.token" placeholder="重置令牌" prefix-icon="Key" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.newPassword" type="password" placeholder="新密码" prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="success" class="full-btn" @click="resetPwd">重置密码</el-button>
        </el-form-item>
      </el-form>
      <div class="back-link">
        <router-link to="/login">返回登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '@/api/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const form = reactive({ email: '', token: '', newPassword: '' })

async function sendEmail() {
  if (!form.email) { ElMessage.warning('请输入邮箱'); return }
  try {
    await authApi.sendResetEmail(form.email)
    ElMessage.success('重置链接已发送')
  } catch {}
}

async function resetPwd() {
  if (!form.token || !form.newPassword) { ElMessage.warning('请填写完整'); return }
  try {
    await authApi.resetPassword({ token: form.token, new_password: form.newPassword })
    ElMessage.success('密码重置成功')
    router.push('/login')
  } catch {}
}
</script>

<style scoped lang="scss">
.reset-page {
  min-height: calc(100vh - 200px);
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f5f7fa, #e5e7eb);
  padding: 40px;
}
.reset-card {
  width: 420px;
  background: #fff;
  border-radius: 16px;
  padding: 48px 40px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.08);
  h2 { text-align: center; margin-bottom: 32px; }
  .full-btn { width: 100%; }
  .back-link { text-align: center; a { color: #1a6fb5; text-decoration: none; font-size: 14px; } }
}
</style>
