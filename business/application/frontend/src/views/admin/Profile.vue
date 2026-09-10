<template>
  <div class="page-card" style="max-width:700px">
    <h3 style="margin-bottom:16px">个人资料</h3>
    <el-form :model="form" label-width="90px">
      <el-form-item label="用户名"><el-input v-model="form.username" disabled /></el-form-item>
      <el-form-item label="姓名"><el-input v-model="form.realName" /></el-form-item>
      <el-form-item label="角色"><el-input :model-value="roleName(form.roleCode)" disabled /></el-form-item>
      <el-form-item label="电话"><el-input v-model="form.phone" /></el-form-item>
      <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
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

const form = reactive({ username: '', realName: '', roleCode: '', phone: '', email: '' })
const pwdForm = reactive({ oldPassword: '', newPassword: '' })

function roleName(code: string) {
  const map: Record<string, string> = { super_admin: '超级管理员', manager: '管理人', reviewer: '评审人' }
  return map[code] || code
}

async function fetch() {
  const res = await authApi.getAdminProfile()
  Object.assign(form, {
    username: res.data.username, realName: res.data.realName, roleCode: res.data.roleCode,
    phone: res.data.phone, email: res.data.email,
  })
}

async function save() {
  await authApi.updateAdminProfile({ realName: form.realName, phone: form.phone, email: form.email })
  ElMessage.success('保存成功')
}

async function changePwd() {
  if (!pwdForm.oldPassword || !pwdForm.newPassword) return ElMessage.warning('请填写完整')
  await authApi.changeAdminPassword(pwdForm)
  ElMessage.success('密码修改成功')
  pwdForm.oldPassword = ''
  pwdForm.newPassword = ''
}

onMounted(fetch)
</script>
