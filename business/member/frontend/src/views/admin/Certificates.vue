<template>
  <div class="admin-certs">
    <div class="page-header"><h3>证书管理</h3><el-button type="primary" @click="showCreate=true">新增证书</el-button></div>
    <el-card><el-empty description="证书管理功能" /></el-card>
    <el-dialog v-model="showCreate" title="新增证书" width="420px">
      <el-form :model="certForm" size="large">
        <el-form-item label="会员ID"><el-input v-model.number="certForm.memberId" /></el-form-item>
        <el-form-item label="证书编号"><el-input v-model="certForm.certNo" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="showCreate=false">取消</el-button><el-button type="primary" @click="createCert">确认</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage } from 'element-plus'
const showCreate = ref(false)
const certForm = reactive({ memberId: 0, certNo: '' })
async function createCert() {
  try { await adminApi.createCertificate({ member_id: certForm.memberId, cert_no: certForm.certNo }); ElMessage.success('创建成功'); showCreate.value = false } catch {}
}
</script>

<style scoped lang="scss">
.admin-certs { max-width: 900px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
</style>
