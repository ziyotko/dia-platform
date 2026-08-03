<template>
  <div>
    <el-card>
      <h3>个人资料</h3>
      <el-form :model="form" label-width="80px">
        <el-form-item label="用户名"><el-input v-model="userInfo.username" disabled /></el-form-item>
        <el-form-item label="姓名"><el-input v-model="form.realName" /></el-form-item>
        <el-form-item label="手机号"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
        <el-form-item label="单位"><el-input v-model="form.company" /></el-form-item>
        <el-form-item label="分会"><el-input v-model="form.branch" disabled /></el-form-item>
        <el-form-item label="会员等级"><el-input v-model="form.memberLevel" disabled /></el-form-item>
        <el-form-item><el-button type="primary" @click="save">保存</el-button></el-form-item>
      </el-form>
    </el-card>

    <el-card style="margin-top:16px">
      <h3>修改密码</h3>
      <el-form :model="pwdForm" label-width="80px">
        <el-form-item label="原密码"><el-input v-model="pwdForm.oldPassword" type="password" show-password /></el-form-item>
        <el-form-item label="新密码"><el-input v-model="pwdForm.newPassword" type="password" show-password /></el-form-item>
        <el-form-item><el-button type="primary" @click="changePwd">修改密码</el-button></el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { authApi } from '@/api/auth'

const userStore = useUserStore()
const userInfo = ref<any>({})

const form = reactive({ realName: '', phone: '', email: '', company: '', branch: '', memberLevel: '' })
const pwdForm = reactive({ oldPassword: '', newPassword: '' })

onMounted(async () => {
  await userStore.fetchUserInfo()
  userInfo.value = userStore.userInfo || {}
  Object.assign(form, userInfo.value)
})

async function save() {
  try { await authApi.updateMemberProfile(form); ElMessage.success('保存成功') } catch (e: any) {}
}

async function changePwd() {
  try { await authApi.changeMemberPassword(pwdForm); ElMessage.success('密码已修改') } catch (e: any) {}
}
</script>
