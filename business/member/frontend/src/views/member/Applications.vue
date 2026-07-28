<template>
  <div class="applications-page" v-loading="loading">
    <div class="page-header">
      <h3>我的申请</h3>
      <el-button type="primary" @click="showCreate = true" v-if="!hasPending">发起入会申请</el-button>
    </div>

    <!-- Application List -->
    <el-card v-if="applications.length">
      <el-table :data="applications" stripe>
        <el-table-column prop="id" label="编号" width="80" />
        <el-table-column prop="org.name" label="申请分会" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="review_comment" label="审核意见" />
        <el-table-column prop="created_at" label="申请时间" width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button text type="primary" @click="viewDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-empty v-else description="暂无申请记录" />

    <!-- Create Application Dialog -->
    <el-dialog v-model="showCreate" title="发起入会申请" width="560px">
      <el-form :model="appForm" label-width="100px" size="large">
        <el-form-item label="申请分会" required>
          <el-select v-model="appForm.orgId" placeholder="选择分会" style="width:100%">
            <el-option v-for="org in orgs" :key="org.id" :label="org.name" :value="org.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="单位名称"><el-input v-model="appForm.companyName" /></el-form-item>
        <el-form-item label="信用代码"><el-input v-model="appForm.creditCode" /></el-form-item>
        <el-form-item label="联系人"><el-input v-model="appForm.contactPerson" /></el-form-item>
        <el-form-item label="单位地址"><el-input v-model="appForm.address" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitApp">提交申请</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { applicationApi, orgApi } from '@/api/index'
import { ElMessage } from 'element-plus'

const applications = ref<any[]>([])
const orgs = ref<any[]>([])
const loading = ref(true)
const showCreate = ref(false)
const submitting = ref(false)
const appForm = reactive({ orgId: null as number | null, companyName: '', creditCode: '', contactPerson: '', address: '' })

const hasPending = computed(() => applications.value.some((a: any) => !['rejected', 'draft'].includes(a.status)))

const statusMap: Record<string, { label: string; tag: string }> = {
  draft: { label: '草稿', tag: 'info' },
  pending_review: { label: '待审核', tag: 'warning' },
  approved: { label: '已通过', tag: 'success' },
  rejected: { label: '已拒绝', tag: 'danger' }
}

function statusLabel(s: string) { return statusMap[s]?.label || s }
function statusTag(s: string) { return statusMap[s]?.tag || 'info' as any }

onMounted(async () => {
  try {
    const [apps, orgData] = await Promise.all([applicationApi.getMyApplications(), orgApi.getTree()])
    applications.value = apps.data || []
    orgs.value = flattenOrgs(orgData.data || [])
  } catch {} finally { loading.value = false }
})

function flattenOrgs(nodes: any[]): any[] {
  let r: any[] = []
  for (const n of nodes) { r.push(n); if (n.children) r = r.concat(flattenOrgs(n.children)) }
  return r
}

function viewDetail(row: any) {
  ElMessage.info(`申请详情 - 编号: ${row.id}`)
}

async function submitApp() {
  if (!appForm.orgId) { ElMessage.warning('请选择分会'); return }
  submitting.value = true
  try {
    await applicationApi.createApplication({
      org_id: appForm.orgId,
      company_name: appForm.companyName,
      credit_code: appForm.creditCode,
      contact_person: appForm.contactPerson,
      address: appForm.address
    })
    ElMessage.success('申请提交成功')
    showCreate.value = false
    const res = await applicationApi.getMyApplications()
    applications.value = res.data || []
  } catch {} finally { submitting.value = false }
}

function formatDate(d: string) { return d ? d.slice(0, 16) : '' }
</script>

<style scoped lang="scss">
.applications-page { max-width: 1000px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
</style>
