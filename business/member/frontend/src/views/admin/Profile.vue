<template>
  <div class="admin-profile" v-loading="loading">
    <el-card class="profile-card">
      <div class="profile-header">
        <el-avatar :size="96" :src="fileUrl(form.avatar)" class="avatar">
          <el-icon :size="40"><UserFilled /></el-icon>
        </el-avatar>
        <div class="avatar-actions">
          <el-button type="primary" :loading="uploading" @click="triggerUpload">
            <el-icon><Upload /></el-icon>&nbsp;更换头像
          </el-button>
          <p class="form-hint">支持 jpg/png/gif/webp，大小不超过 10MB</p>
        </div>
        <input ref="fileInputRef" type="file" accept=".jpg,.jpeg,.png,.gif,.webp,.bmp" style="display:none" @change="handleAvatarUpload" />
      </div>

      <el-divider />

      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px" size="large">
        <el-form-item label="用户名">
          <el-input v-model="form.username" disabled />
        </el-form-item>
        <el-form-item label="手机号" prop="mobile">
          <el-input v-model="form.mobile" maxlength="11" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" maxlength="40" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="身份">
          <el-tag :type="form.is_admin ? 'danger' : 'info'">{{ form.is_admin ? '管理员' : '会员' }}</el-tag>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="saveProfile">保存修改</el-button>
          <el-button style="margin-left:12px" @click="openPwdDialog">修改密码</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-dialog v-model="showPwdDialog" title="修改密码" width="420px" :close-on-click-modal="false">
      <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="100px" size="large">
        <el-form-item label="原密码" prop="oldPassword">
          <el-input v-model="pwdForm.oldPassword" type="password" show-password placeholder="请输入原密码" />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input v-model="pwdForm.newPassword" type="password" show-password placeholder="请输入新密码（至少8位）" />
        </el-form-item>
        <el-form-item label="确认新密码" prop="confirmPassword">
          <el-input v-model="pwdForm.confirmPassword" type="password" show-password placeholder="请再次输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPwdDialog = false">取消</el-button>
        <el-button type="primary" :loading="changingPwd" @click="changePwd">确认修改</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { authApi } from '@/api/auth'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import { Upload, UserFilled } from '@element-plus/icons-vue'

const userStore = useUserStore()
const form = reactive<any>({})
const original = reactive<any>({})
const formRef = ref()
const loading = ref(true)
const saving = ref(false)
const uploading = ref(false)
const fileInputRef = ref<HTMLInputElement>()

const showPwdDialog = ref(false)
const changingPwd = ref(false)
const pwdFormRef = ref()
const pwdForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

function fileUrl(path?: string) {
  if (!path) return ''
  if (path.startsWith('http://') || path.startsWith('https://')) return path
  return '/' + path.replace(/^\//, '')
}

function debounce<A extends any[]>(fn: (...args: A) => void, delay = 400) {
  let timer: ReturnType<typeof setTimeout> | null = null
  return (...args: A) => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }
}

const checkMobileAvailable = debounce(async (value: string, callback: any) => {
  try {
    const res = await authApi.checkExists({ field: 'mobile', value })
    if (res.data.exists) callback(new Error('该手机号已被其他用户使用'))
    else callback()
  } catch {
    callback()
  }
})

const validateMobile = (_rule: any, value: string, callback: any) => {
  if (!value) return callback()
  if (value.length > 11) return callback(new Error('手机号不能超过11位'))
  if (!/^1[3-9]\d{9}$/.test(value)) {
    return callback(new Error('请输入合法的手机号'))
  }
  if (value === original.mobile) return callback()
  checkMobileAvailable(value, callback)
}

const checkEmailAvailable = debounce(async (value: string, callback: any) => {
  try {
    const res = await authApi.checkExists({ field: 'email', value })
    if (res.data.exists) callback(new Error('该邮箱已被其他用户使用'))
    else callback()
  } catch {
    callback()
  }
})

const validateEmail = (_rule: any, value: string, callback: any) => {
  if (!value) return callback()
  if (value.length > 40) return callback(new Error('邮箱不能超过40位'))
  if (!/^[\w.+-]+@[\w-]+(\.[\w-]+)+$/.test(value)) {
    return callback(new Error('邮箱格式不正确'))
  }
  if (value === original.email) return callback()
  checkEmailAvailable(value, callback)
}

const rules = {
  mobile: [{ validator: validateMobile, trigger: 'blur' }],
  email: [{ validator: validateEmail, trigger: 'blur' }]
}

const pwdRules = {
  oldPassword: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '新密码至少8位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: any) => {
        if (value !== pwdForm.newPassword) callback(new Error('两次输入的密码不一致'))
        else callback()
      },
      trigger: 'blur'
    }
  ]
}

onMounted(async () => {
  try {
    const res = await authApi.getProfile()
    Object.assign(form, res.data)
    Object.assign(original, res.data)
  } catch {
  } finally {
    loading.value = false
  }
})

function triggerUpload() {
  fileInputRef.value?.click()
}

async function handleAvatarUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const maxSize = 10 * 1024 * 1024 // 10MB
  if (file.size > maxSize) {
    ElMessage.error('文件大小不能超过10MB')
    input.value = ''
    return
  }

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('dir', 'avatars')
    const res = await authApi.upload(formData)
    form.avatar = res.data.url
    ElMessage.success('头像已更新，请点击保存修改以确认')
  } catch {
    ElMessage.error('上传失败，请重试')
  } finally {
    uploading.value = false
    input.value = ''
  }
}

async function saveProfile() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    await authApi.updateProfile(form)
    Object.assign(original, form)
    await userStore.fetchUserInfo()
    ElMessage.success('保存成功')
  } catch {
  } finally {
    saving.value = false
  }
}

function openPwdDialog() {
  Object.assign(pwdForm, { oldPassword: '', newPassword: '', confirmPassword: '' })
  pwdFormRef.value?.clearValidate()
  showPwdDialog.value = true
}

async function changePwd() {
  const valid = await pwdFormRef.value?.validate().catch(() => false)
  if (!valid) return
  changingPwd.value = true
  try {
    await authApi.changePassword({ old_password: pwdForm.oldPassword, new_password: pwdForm.newPassword })
    ElMessage.success('密码修改成功')
    showPwdDialog.value = false
  } catch {
  } finally {
    changingPwd.value = false
  }
}
</script>

<style scoped lang="scss">
.admin-profile {
  max-width: 720px;
}
.profile-card {
  border-radius: 8px;
}
.profile-header {
  display: flex;
  align-items: center;
  gap: 24px;
  .avatar {
    background: #e5e7eb;
    flex-shrink: 0;
  }
  .avatar-actions {
    .form-hint {
      margin: 8px 0 0;
      color: #909399;
      font-size: 13px;
    }
  }
}
</style>
