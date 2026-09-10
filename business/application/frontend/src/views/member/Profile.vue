<template>
  <div class="page-card" style="max-width:700px">
    <h3 style="margin-bottom:16px">个人资料</h3>
    <el-form :model="form" label-width="90px">
      <el-form-item label="用户名"><el-input v-model="form.username" disabled /></el-form-item>
      <el-form-item label="真实姓名"><el-input v-model="form.realName" /></el-form-item>
      <el-form-item label="联系电话"><el-input v-model="form.phone" /></el-form-item>
      <el-form-item label="电子邮箱"><el-input v-model="form.email" /></el-form-item>
      <el-form-item label="身份证号"><el-input v-model="form.idCard" /></el-form-item>
      <el-form-item label="所在单位"><el-input v-model="form.organization" /></el-form-item>
      <el-form-item label="职务"><el-input v-model="form.position" /></el-form-item>
      <el-form-item>
        <el-button type="primary" @click="save">保存资料</el-button>
      </el-form-item>
    </el-form>

    <el-divider />
    <h3 style="margin-bottom:16px">修改密码</h3>
    <el-form :model="pwdForm" label-width="90px">
      <el-form-item label="原密码"><el-input v-model="pwdForm.oldPassword" type="password" show-password /></el-form-item>
      <el-form-item label="新密码"><el-input v-model="pwdForm.newPassword" type="password" show-password /></el-form-item>
      <el-form-item>
        <el-button type="primary" @click="changePwd">修改密码</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { authApi } from '@/api/auth'

const form = reactive({ username: '', realName: '', phone: '', email: '', idCard: '', organization: '', position: '' })
const pwdForm = reactive({ oldPassword: '', newPassword: '' })

async function fetch() {
  const res = await authApi.getUserProfile()
  Object.assign(form, {
    username: res.data.username, realName: res.data.realName, phone: res.data.phone,
    email: res.data.email, idCard: res.data.idCard, organization: res.data.organization, position: res.data.position,
  })
}

async function save() {
  await authApi.updateUserProfile({
    realName: form.realName, phone: form.phone, email: form.email,
    idCard: form.idCard, organization: form.organization, position: form.position,
  })
  ElMessage.success('保存成功')
}

async function changePwd() {
  if (!pwdForm.oldPassword || !pwdForm.newPassword) return ElMessage.warning('请填写完整')
  await authApi.changeUserPassword(pwdForm)
  ElMessage.success('密码修改成功')
  pwdForm.oldPassword = ''
  pwdForm.newPassword = ''
}

onMounted(fetch)
</script>
