<template>
  <div>
    <el-card><h3>个人资料</h3><p>用户名：{{ adminInfo.username }}</p><p>角色：{{ adminInfo.roleCode }}</p></el-card>
    <el-card style="margin-top:16px"><h3>修改密码</h3>
      <el-form :model="pwdForm" label-width="80px">
        <el-form-item label="原密码"><el-input v-model="pwdForm.oldPassword" type="password" /></el-form-item>
        <el-form-item label="新密码"><el-input v-model="pwdForm.newPassword" type="password" /></el-form-item>
        <el-form-item><el-button type="primary" @click="changePwd">修改密码</el-button></el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAdminStore } from '@/stores/admin'
import { authApi } from '@/api/auth'

const adminStore = useAdminStore()
const adminInfo = ref<any>({})
const pwdForm = reactive({ oldPassword: '', newPassword: '' })

onMounted(async () => { await adminStore.fetchAdminInfo(); adminInfo.value = adminStore.adminInfo || {} })

async function changePwd() { try { await authApi.changeAdminPassword(pwdForm); ElMessage.success('密码已修改') } catch (e: any) {} }
</script>
