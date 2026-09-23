<template>
  <div class="admin-apps" v-loading="loading">
    <div class="page-header"><h3>入会审核</h3></div>
    <div class="filter-bar">
      <el-radio-group v-model="statusFilter" @change="onFilterChange" size="small">
        <el-radio-button value="">全部</el-radio-button>
        <el-radio-button value="pending_review">待审核</el-radio-button>
        <el-radio-button value="approved">已通过</el-radio-button>
        <el-radio-button value="rejected">已拒绝</el-radio-button>
      </el-radio-group>
    </div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="申请单位/个人">
          <template #default="{row}">{{ row.member?.member_type === 'personal' ? (row.member?.name || '-') : (row.member?.company_name || '-') }}</template>
        </el-table-column>
        <el-table-column prop="member.username" label="用户名" />
        <el-table-column prop="org.name" label="申请总会/分会" />
        <el-table-column label="类别" width="90">
          <template #default="{row}">{{ row.member?.member_type === 'personal' ? '个人会员' : '单位会员' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{row}"><el-tag :type="row.status==='pending_review'?'warning':row.status==='approved'?'success':'danger'">{{ statusLabel(row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="created_at" label="申请时间" width="170"><template #default="{row}">{{ row.created_at?.slice(0,16).replace('T', ' ') }}</template></el-table-column>
        <el-table-column label="操作" width="280" v-if="list.length">
          <template #default="{row}">
            <el-button size="small" @click="viewMember(row)">查看</el-button>
            <template v-if="row.status === 'pending_review'">
              <el-button size="small" type="success" @click="review(row, true)">通过</el-button>
              <el-button size="small" type="danger" @click="review(row, false)">拒绝</el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length===0" />
      <div class="pagination"><el-pagination background layout="prev, pager, next" :total="total" :page-size="size" v-model:current-page="page" @change="fetchData" /></div>
    </el-card>

    <!-- Member Detail Dialog -->
    <el-dialog v-model="showMemberDetail" title="申请单位信息" width="640px">
      <el-descriptions :column="1" border v-if="currentMember">
        <el-descriptions-item label="单位名称/姓名">{{ currentMember.member_type === 'personal' ? (currentMember.name || '-') : (currentMember.company_name || '-') }}</el-descriptions-item>
        <template v-if="currentMember.member_type === 'unit'">
        <el-descriptions-item label="统一社会信用代码">
          {{ currentMember.credit_code || '-' }}
          <el-link v-if="currentMember.cert_file" :href="fileUrl(currentMember.cert_file)" target="_blank" type="primary" underline="never" style="margin-left:12px">下载证照</el-link>
        </el-descriptions-item>
        <el-descriptions-item label="法定代表人">{{ currentMember.legal_person || '-' }}</el-descriptions-item>
        <el-descriptions-item label="联系人">{{ currentMember.contact_person || '-' }}</el-descriptions-item>
        <el-descriptions-item label="单位地址">{{ currentMember.address || '-' }}</el-descriptions-item>
        <el-descriptions-item label="网站">{{ currentMember.website || '-' }}</el-descriptions-item>
        </template>
        <el-descriptions-item label="会员类型">{{ currentMember.member_type === 'personal' ? '个人会员' : '单位会员' }}</el-descriptions-item>
        <el-descriptions-item label="入会申请书">
          <el-link v-if="currentRow?.signed_file" :href="fileUrl(currentRow.signed_file)" target="_blank" type="primary" underline="never">下载申请书</el-link>
          <span v-else>-</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import { fileUrl } from '@/utils/fileUrl'

const list = ref<any[]>([]); const loading = ref(true)
const page = ref(1); const size = ref(10); const total = ref(0)
const statusFilter = ref('')
const showMemberDetail = ref(false)
const currentMember = ref<any>(null)
const currentRow = ref<any>(null)

const sm: Record<string,string> = { pending_review:'待审核', approved:'已通过', rejected:'已拒绝' }
function statusLabel(s: string) { return sm[s] || s }

onMounted(() => fetchData())
function onFilterChange() {
  page.value = 1
  fetchData()
}

async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: page.value, size: size.value }
    if (statusFilter.value) params.status = statusFilter.value
    const r = await adminApi.getApplications(params)
    list.value = r.data?.list || []
    total.value = r.data?.total || 0
  } catch {} finally { loading.value = false }
}
function viewMember(row: any) {
  currentMember.value = row.member || null
  currentRow.value = row
  showMemberDetail.value = true
}

async function review(row: any, approved: boolean) {
  try {
    const title = approved ? '通过申请' : '拒绝申请'
    const tip = approved ? '请输入审核意见（选填）' : '请输入拒绝理由'
    const { value: comment } = await ElMessageBox.prompt(tip, title, {
      inputType: 'textarea',
      inputPlaceholder: approved ? '可选，输入审核意见...' : '请输入拒绝理由...',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputValidator: (v: string) => (!approved && !(v && v.trim()) ? '请填写拒绝理由' : true)
    })
    await adminApi.reviewApplication(row.id, { approved, comment: (comment || '').trim() })
    ElMessage.success(approved ? '已通过' : '已拒绝')
    fetchData()
  } catch {}
}
</script>

<style scoped lang="scss">
.admin-apps {
 width: 100%;
}
.page-header {
  margin-bottom: 24px;
  h3 {
    font-size: 22px;
    font-weight: 600;
    color: #1d2739;
    margin: 0;
  }
}
.filter-bar { margin-bottom: 16px; }
.el-card {
  border-radius: 10px;
}
.pagination { display: flex; justify-content: center; padding: 20px 0; }
</style>
