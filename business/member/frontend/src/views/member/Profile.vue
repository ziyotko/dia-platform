<template>
  <div class="profile-page" v-loading="loading">
    <el-card>
      <template #header><span>我的资料</span></template>
      <el-form :model="form" label-width="100px" size="large">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="用户名"><el-input v-model="form.username" disabled /></el-form-item>
            <el-form-item label="手机号"><el-input v-model="form.mobile" /></el-form-item>
            <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
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
            <el-form-item label="单位名称"><el-input v-model="form.company_name" /></el-form-item>
            <el-form-item label="信用代码"><el-input v-model="form.credit_code" /></el-form-item>
            <el-form-item label="法定代表人"><el-input v-model="form.legal_person" /></el-form-item>
            <el-form-item label="联系人"><el-input v-model="form.contact_person" /></el-form-item>
            <el-form-item label="单位地址"><el-input v-model="form.address" /></el-form-item>
            <el-form-item label="网站"><el-input v-model="form.website" /></el-form-item>
            <el-form-item label="组织机构证" v-if="form.member_type === 'unit'">
              <template v-if="form.cert_file">
                <el-link :href="fileUrl(form.cert_file)" target="_blank" type="primary" :underline="false">
                  <el-icon style="margin-right:4px"><Download /></el-icon>查看证照
                </el-link>
              </template>
              <span v-else style="color:#9ca3af">未上传</span>
            </el-form-item>
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
import { Download } from '@element-plus/icons-vue'

const form = reactive<any>({})
const loading = ref(true)
const saving = ref(false)
const showPwdDialog = ref(false)
const pwdForm = reactive({ oldPassword: '', newPassword: '' })

onMounted(async () => {
  try {
    const res = await authApi.getProfile()
    Object.assign(form, res.data)
  } catch {} finally { loading.value = false }
})

async function saveProfile() {
  saving.value = true
  try {
    await authApi.updateProfile(form)
    ElMessage.success('保存成功')
  } catch {} finally { saving.value = false }
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
.profile-page { max-width: 900px; margin: 0 auto; }
</style>
