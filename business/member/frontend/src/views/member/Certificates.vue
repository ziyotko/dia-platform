<template>
  <div class="certs-page" v-loading="loading">
    <div class="page-header"><h3>我的证书</h3></div>
    <el-card>
      <el-table :data="certificates" stripe>
        <el-table-column prop="cert_no" label="证书编号" />
        <el-table-column prop="issued_at" label="颁发日期" width="120">
          <template #default="{ row }">{{ formatDate(row.issued_at) }}</template>
        </el-table-column>
        <el-table-column prop="expire_at" label="有效期至" width="120">
          <template #default="{ row }">{{ formatDate(row.expire_at) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'">
              {{ row.status === 'active' ? '有效' : '已过期' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button text type="primary" @click="downloadCert(row)">
              <el-icon><Download /></el-icon> 下载证书
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && certificates.length === 0" description="暂无证书" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { certificateApi } from '@/api/index'
import { ElMessage } from 'element-plus'

const certificates = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await certificateApi.getMyCertificates()
    certificates.value = res.data || []
  } catch {} finally { loading.value = false }
})

function downloadCert(row: any) {
  if (row.file_path) {
    window.open(row.file_path, '_blank')
  } else {
    ElMessage.info('证书文件暂未生成，请联系管理员')
  }
}

function formatDate(d: string) { return d ? d.slice(0, 10) : '' }
</script>

<style scoped lang="scss">
.certs-page { max-width: 900px; margin: 0 auto; }
.page-header { margin-bottom: 20px; }
</style>
