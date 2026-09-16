<template>
  <div class="certs-page" v-loading="loading">
    <div class="page-header"><h3>我的证书</h3></div>
    <el-card>
      <el-table :data="certificates" stripe style="width:100%">
        <el-table-column prop="cert_no" label="证书编号" min-width="160" />
        <el-table-column prop="issued_at" label="颁发日期" min-width="110">
          <template #default="{ row }">{{ formatDate(row.issued_at) }}</template>
        </el-table-column>
        <el-table-column prop="level_name" label="会员级别" min-width="110" />
        <el-table-column prop="expire_at" label="有效期至" min-width="110">
          <template #default="{ row }">{{ formatDate(row.expire_at) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" min-width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'">
              {{ row.status === 'active' ? '有效' : '已过期' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="260">
          <template #default="{ row }">
            <el-button text type="primary" :disabled="!isActiveMember" @click="isActiveMember ? downloadCert(row) : undefined">
              <el-icon><Download /></el-icon> 下载证书
            </el-button>
            <el-button
              v-if="isActiveMember && row.status !== 'active' && !hasActiveCert"
              text
              type="warning"
              @click="renewCert"
            >
              <el-icon><Refresh /></el-icon> 刷新
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && certificates.length === 0" description="暂无证书" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { certificateApi } from '@/api/index'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download, Refresh } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const isActiveMember = computed(() => userStore.userInfo?.status === 'active')

const certificates = ref<any[]>([])
const loading = ref(true)
const renewing = ref(false)

const hasActiveCert = computed(() => certificates.value.some(c => c.status === 'active'))

onMounted(() => fetchCertificates())

async function fetchCertificates() {
  loading.value = true
  try {
    const res = await certificateApi.getMyCertificates()
    certificates.value = res.data || []
  } catch {} finally { loading.value = false }
}

function downloadCert(row: any) {
  if (!isActiveMember.value) {
    ElMessage.warning('仅正式会员可下载证书')
    return
  }
  if (row.file_path) {
    window.open(fileUrl(row.file_path), '_blank')
  } else {
    ElMessage.info('证书文件暂未生成，请联系管理员')
  }
}

function fileUrl(path?: string) {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path
  // 兼容 `uploads/...`（相对）、`/uploads/...`（旧绝对）与带部署前缀的绝对路径，统一指向当前部署子路径
  const base = import.meta.env.BASE_URL || '/'
  const clean = path.replace(/^\.?\//, '')
  return clean.startsWith('uploads/') ? `${base}${clean}` : `/${clean}`
}

async function renewCert() {
  if (!isActiveMember.value) {
    ElMessage.warning('仅正式会员可更新证书')
    return
  }
  try {
    await ElMessageBox.confirm(
      '当前证书已失效，确认重新生成证书？',
      '证书刷新',
      { type: 'warning', confirmButtonText: '确认刷新', cancelButtonText: '取消' }
    )
    renewing.value = true
    await certificateApi.renewCertificate()
    ElMessage.success('证书已重新生成')
    await fetchCertificates()
  } catch {} finally { renewing.value = false }
}

function formatDate(d: string) { return d ? d.slice(0, 10) : '' }
</script>

<style scoped lang="scss">
.certs-page {  width: 100%;}
.page-header { margin-bottom: 20px; }
</style>
