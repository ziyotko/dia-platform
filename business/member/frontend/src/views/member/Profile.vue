<template>
  <div class="profile-page" v-loading="loading">
    <el-card>
      <template #header><span>我的资料</span></template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" size="large">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="用户名"><el-input v-model="form.username" disabled /></el-form-item>
            <el-form-item label="手机号" prop="mobile"><el-input v-model="form.mobile" maxlength="11" /></el-form-item>
            <el-form-item label="邮箱" prop="email"><el-input v-model="form.email" maxlength="40" /></el-form-item>
            <el-form-item label="会员类型">
              <el-tag>{{ form.member_type === 'unit' ? '单位会员' : '个人会员' }}</el-tag>
            </el-form-item>
            <el-form-item label="会员等级">
              <el-tag :type="!form.member_level || form.member_level === 'normal' ? 'info' : 'warning'">
                {{ !form.member_level || form.member_level === 'normal' ? '暂无' : form.member_level }}
              </el-tag>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <template v-if="form.member_type === 'unit'">
            <el-form-item label="单位名称"><el-input v-model="form.company_name" /></el-form-item>
            <el-form-item label="组织机构代码证"><el-input v-model="form.credit_code" /></el-form-item>
            <el-form-item label="法定代表人"><el-input v-model="form.legal_person" /></el-form-item>
            <el-form-item label="联系人"><el-input v-model="form.contact_person" /></el-form-item>
            <el-form-item label="单位地址"><el-input v-model="form.address" /></el-form-item>
            <el-form-item label="网站"><el-input v-model="form.website" /></el-form-item>
            <el-form-item label="组织机构证">
              <div class="cert-file-row">
                <template v-if="form.cert_file">
                  <el-link :href="fileUrl(form.cert_file)" target="_blank" type="primary" :underline="false">
                    <el-icon style="margin-right:4px"><Download /></el-icon>下载证照
                  </el-link>
                  <el-button size="small" style="margin-left:8px" @click="triggerUpload">
                    <el-icon><Upload /></el-icon>重新上传
                  </el-button>
                </template>
                <template v-else>
                  <el-button size="small" type="primary" @click="triggerUpload">
                    <el-icon><Upload /></el-icon>上传组织机构证
                  </el-button>
                </template>
                <input ref="fileInputRef" type="file" accept=".jpg,.jpeg,.png,.pdf" style="display:none" @change="handleCertUpload" />
                <span v-if="uploading" style="margin-left:8px;color:#409eff">上传中...</span>
              </div>
            </el-form-item>
            </template>
            <template v-else>
              <el-form-item label="姓名"><el-input v-model="form.name" /></el-form-item>
              <el-form-item label="身份证号" prop="id_card"><el-input v-model="form.id_card" maxlength="18" /></el-form-item>
            </template>
          </el-col>
        </el-row>
        <el-form-item label="简介">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="saveProfile">保存修改</el-button>
          <el-button style="margin-left:12px" @click="showPwdDialog = true">修改密码</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-dialog v-model="showPwdDialog" title="修改密码" width="420px">
      <el-form :model="pwdForm" size="large">
        <el-form-item><el-input v-model="pwdForm.oldPassword" type="password" placeholder="原密码" show-password /></el-form-item>
        <el-form-item><el-input v-model="pwdForm.newPassword" type="password" placeholder="新密码" show-password /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPwdDialog = false">取消</el-button>
        <el-button type="primary" @click="changePwd">确认修改</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { authApi } from '@/api/auth'
import { ElMessage } from 'element-plus'
import { Download, Upload } from '@element-plus/icons-vue'

const form = reactive<any>({})
const original = reactive<any>({})
const formRef = ref()
const loading = ref(true)
const saving = ref(false)
const showPwdDialog = ref(false)
const pwdForm = reactive({ oldPassword: '', newPassword: '' })
const uploading = ref(false)
const fileInputRef = ref<HTMLInputElement>()

// 防抖：避免输入过程中频繁请求后端查重
function debounce<A extends any[]>(fn: (...args: A) => void, delay = 400) {
  let timer: ReturnType<typeof setTimeout> | null = null
  return (...args: A) => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }
}

// 手机号：与注册一致（合法性 + 查重，未改动时跳过查重避免误报）
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
  if (!value) return callback(new Error('请输入手机号'))
  if (value.length > 11) return callback(new Error('手机号不能超过11位'))
  if (!/^1[3-9]\d{9}$/.test(value)) {
    return callback(new Error('请输入合法的手机号'))
  }
  if (value === original.mobile) return callback()
  checkMobileAvailable(value, callback)
}

// 邮箱：与注册一致（合法性 + 查重，未改动时跳过查重避免误报）
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
  if (!value) return callback(new Error('请输入邮箱'))
  if (value.length > 40) return callback(new Error('邮箱不能超过40位'))
  if (!/^[\w.+-]+@[\w-]+(\.[\w-]+)+$/.test(value)) {
    return callback(new Error('邮箱格式不正确'))
  }
  if (value === original.email) return callback()
  checkEmailAvailable(value, callback)
}

// 身份证号校验（18位，出生日期 + 校验码）
const validateIdCard = (_rule: any, value: string, callback: any) => {
  if (!value) return callback(new Error('请输入身份证号'))
  const id = value.trim().toUpperCase()
  if (!/^\d{17}[\dX]$/.test(id)) {
    return callback(new Error('身份证号应为18位，末位可为X'))
  }
  const year = +id.slice(6, 10)
  const month = +id.slice(10, 12)
  const day = +id.slice(12, 14)
  const date = new Date(year, month - 1, day)
  if (date.getFullYear() !== year || date.getMonth() + 1 !== month || date.getDate() !== day) {
    return callback(new Error('身份证号出生日期不合法'))
  }
  const weights = [7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2]
  const codes = '10X98765432'
  let sum = 0
  for (let i = 0; i < 17; i++) {
    sum += +id[i] * weights[i]
  }
  if (codes[sum % 11] !== id[17]) {
    return callback(new Error('身份证号校验不通过'))
  }
  callback()
}

const rules = {
  mobile: [{ validator: validateMobile, trigger: 'blur' }],
  email: [{ validator: validateEmail, trigger: 'blur' }],
  id_card: [{ validator: validateIdCard, trigger: 'blur' }]
}

onMounted(async () => {
  try {
    const res = await authApi.getProfile()
    Object.assign(form, res.data)
    Object.assign(original, res.data)
  } catch {} finally { loading.value = false }
})

async function saveProfile() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    await authApi.updateProfile(form)
    Object.assign(original, form)
    ElMessage.success('保存成功')
  } catch {} finally { saving.value = false }
}

function triggerUpload() {
  fileInputRef.value?.click()
}

async function handleCertUpload(event: Event) {
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
    formData.append('dir', 'certs')
    const res = await authApi.upload(formData)
    form.cert_file = res.data.url
    ElMessage.success('证照上传成功，请点击保存修改以确认')
  } catch {
    ElMessage.error('上传失败，请重试')
  } finally {
    uploading.value = false
    input.value = ''
  }
}

function fileUrl(path: string) {
  if (!path) return ''
  if (path.startsWith('http://') || path.startsWith('https://')) return path
  return '/' + path.replace(/^\//, '')
}

async function changePwd() {
  if (!pwdForm.oldPassword || !pwdForm.newPassword) return
  try {
    await authApi.changePassword({ old_password: pwdForm.oldPassword, new_password: pwdForm.newPassword })
    ElMessage.success('密码修改成功')
    showPwdDialog.value = false
  } catch {}
}
</script>

<style scoped lang="scss">
.profile-page {  width: 100%;}
</style>
