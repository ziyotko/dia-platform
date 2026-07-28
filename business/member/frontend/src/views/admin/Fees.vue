<template>
  <div class="admin-fees" v-loading="loading">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <h3>会费管理</h3>
      </div>
      <el-button type="primary" :icon="Plus" size="large" @click="showCreate=true">新增费用</el-button>
    </div>

    <!-- Stats Cards -->
    <div class="stats-row">
      <div class="stat-card stat-card--total">
        <div class="stat-icon"><el-icon :size="28"><List /></el-icon></div>
        <div class="stat-body">
          <span class="stat-label">总记录数</span>
          <span class="stat-value">{{ total }}</span>
        </div>
      </div>
      <div class="stat-card stat-card--paid">
        <div class="stat-icon"><el-icon :size="28"><CircleCheckFilled /></el-icon></div>
        <div class="stat-body">
          <span class="stat-label">已缴费</span>
          <span class="stat-value">{{ paidCount }}</span>
        </div>
      </div>
      <div class="stat-card stat-card--unpaid">
        <div class="stat-icon"><el-icon :size="28"><WarningFilled /></el-icon></div>
        <div class="stat-body">
          <span class="stat-label">未缴费</span>
          <span class="stat-value">{{ unpaidCount }}</span>
        </div>
      </div>
      <div class="stat-card stat-card--amount">
        <div class="stat-icon"><el-icon :size="28"><Money /></el-icon></div>
        <div class="stat-body">
          <span class="stat-label">总金额</span>
          <span class="stat-value">¥{{ totalAmount.toFixed(2) }}</span>
        </div>
      </div>
    </div>

    <!-- Filter Bar -->
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-select v-model="filterYear" placeholder="选择年度" clearable style="width:140px" @change="fetchData">
          <el-option v-for="y in yearOptions" :key="y" :label="y+'年'" :value="y" />
        </el-select>
        <el-select v-model="filterStatus" placeholder="缴费状态" clearable style="width:140px" @change="fetchData">
          <el-option label="已缴费" value="paid" />
          <el-option label="未缴费" value="unpaid" />
        </el-select>
        <span class="filter-hint">共 {{ total }} 条记录</span>
      </div>
    </el-card>

    <!-- Table -->
    <el-card shadow="never" class="table-card">
      <el-table :data="list" stripe highlight-current-row>
        <el-table-column type="index" label="#" width="50" align="center" />
        <el-table-column prop="member.username" label="会员" min-width="200">
          <template #default="{row}">
            <div class="cell-member">
              <span class="member-name">{{ row.member?.username }}</span>
              <span v-if="row.member?.company_name" class="member-unit">{{ row.member.company_name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="org_name" label="入会信息" min-width="140" />
        <el-table-column prop="level_name" label="会员级别" width="110" />
        <el-table-column prop="year" label="年度" width="80" align="center" />
        <el-table-column prop="amount" label="金额" width="120" align="right">
          <template #default="{row}">
            <span class="cell-amount">¥{{ row.amount?.toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{row}">
            <el-tag v-if="row.status==='paid'" type="success" effect="dark" round>
              <el-icon style="vertical-align:-2px;margin-right:3px"><CircleCheckFilled /></el-icon>已缴费
            </el-tag>
            <el-tag v-else type="warning" effect="dark" round>
              <el-icon style="vertical-align:-2px;margin-right:3px"><WarningFilled /></el-icon>未缴费
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="paid_at" label="缴费时间" width="170">
          <template #default="{row}">
            <span class="cell-time">{{ row.paid_at?.slice(0,16).replace('T',' ') || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{row}">
            <div class="action-btns">
              <el-button text size="small" :icon="Edit" @click="editFee(row)">编辑</el-button>
              <el-button v-if="row.status!=='paid'" text size="small" type="success" :icon="CircleCheck" @click="markPaid(row)">缴费</el-button>
              <el-button v-if="row.status!=='paid'" text size="small" type="danger" :icon="Delete" @click="deleteFee(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination">
        <el-pagination background layout="total, prev, pager, next" :total="total" :page-size="size" v-model:current-page="page" @current-change="fetchData" />
      </div>
    </el-card>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreate" title="新增费用" width="560px" :close-on-click-modal="false" class="fee-dialog">
      <el-form :model="feeForm" label-width="90px" class="fee-form">
        <el-form-item label="会员单位" required>
          <el-select v-model="feeForm.memberId" filterable remote :remote-method="searchMember" :loading="searchLoading" placeholder="按名称搜索会员单位" style="width:100%" @change="onMemberSelect">
            <el-option v-for="m in memberOptions" :key="m.id" :label="`${m.company_name || m.username}${m.company_name ? ' ('+m.username+')' : ''}`" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="入会信息">
          <div v-if="selectedMemberInfo.orgName" class="member-info-tag">
            <el-tag type="info" round>{{ selectedMemberInfo.orgName }}</el-tag>
            <el-tag v-if="selectedMemberInfo.levelName" type="primary" round effect="plain">{{ selectedMemberInfo.levelName }}</el-tag>
          </div>
          <span v-else class="form-hint">选择会员后自动带出</span>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="年度" required>
              <el-select v-model="feeForm.year" style="width:100%">
                <el-option v-for="y in yearOptions" :key="y" :label="y+'年'" :value="y" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="金额" required>
              <el-input-number v-model="feeForm.amount" :min="0" :precision="2" :step="500" style="width:100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="备注">
          <el-input v-model="feeForm.remark" type="textarea" :rows="2" placeholder="可选填写备注信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate=false">取消</el-button>
        <el-button type="primary" :loading="saving" :icon="Plus" @click="createFee">确认创建</el-button>
      </template>
    </el-dialog>

    <!-- Edit Dialog -->
    <el-dialog v-model="showEdit" title="编辑费用" width="480px" :close-on-click-modal="false" class="fee-dialog">
      <el-form :model="editForm" label-width="80px" class="fee-form">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="金额" required>
              <el-input-number v-model="editForm.amount" :min="0" :precision="2" :step="500" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-select v-model="editForm.status" style="width:100%">
                <el-option label="未缴费" value="unpaid" />
                <el-option label="已缴费" value="paid" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="备注">
          <el-input v-model="editForm.remark" type="textarea" :rows="2" placeholder="可选填写备注信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEdit=false">取消</el-button>
        <el-button type="primary" :icon="Check" @click="saveEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete, CircleCheck, CircleCheckFilled, WarningFilled, Coin, List, Money, Check } from '@element-plus/icons-vue'

const list = ref<any[]>([]); const loading = ref(true); const showCreate = ref(false)
const page = ref(1); const size = ref(10); const total = ref(0)

const memberOptions = ref<any[]>([])
const searchLoading = ref(false)
const saving = ref(false)
const yearOptions = ref<number[]>([])
const selectedMemberInfo = reactive({ orgId: 0, levelId: 0, orgName: '', levelName: '' })
const feeForm = reactive({ memberId: null as number | null, year: new Date().getFullYear(), amount: 2000, remark: '' })

const showEdit = ref(false)
const editForm = reactive({ id: 0, amount: 0, status: 'unpaid', remark: '' })

const filterYear = ref<number | null>(null)
const filterStatus = ref<string | null>(null)

const paidCount = computed(() => list.value.filter(r => r.status === 'paid').length)
const unpaidCount = computed(() => list.value.filter(r => r.status !== 'paid').length)
const totalAmount = computed(() => list.value.reduce((s, r) => s + (r.amount || 0), 0))

onMounted(() => {
  fetchData()
  const y = new Date().getFullYear()
  yearOptions.value = Array.from({ length: 10 }, (_, i) => y - i)
})
async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: page.value, size: size.value }
    if (filterYear.value) params.year = filterYear.value
    if (filterStatus.value) params.status = filterStatus.value
    const r = await adminApi.getFees(params)
    list.value = r.data?.list || []
    total.value = r.data?.total || 0
  } catch {} finally { loading.value = false }
}
async function searchMember(query: string) {
  if (!query) return
  searchLoading.value = true
  try {
    const r = await adminApi.getMembers({ keyword: query, page: 1, size: 20 })
    memberOptions.value = r.data?.list || []
  } catch {} finally { searchLoading.value = false }
}
async function onMemberSelect(id: number) {
  selectedMemberInfo.orgId = 0
  selectedMemberInfo.levelId = 0
  selectedMemberInfo.orgName = ''
  selectedMemberInfo.levelName = ''
  try {
    const r = await adminApi.getMemberFeeInfo(id)
    const d = r.data || {}
    selectedMemberInfo.orgId = d.org_id || 0
    selectedMemberInfo.levelId = d.level_id || 0
    selectedMemberInfo.orgName = d.org_name || ''
    selectedMemberInfo.levelName = d.level_name || ''
  } catch {}
}
async function createFee() {
  if (!feeForm.memberId) { ElMessage.warning('请选择会员单位'); return }
  saving.value = true
  try {
    await adminApi.createFee({
      member_id: feeForm.memberId,
      year: feeForm.year,
      amount: feeForm.amount,
      remark: feeForm.remark,
      org_id: selectedMemberInfo.orgId,
      org_name: selectedMemberInfo.orgName,
      level_id: selectedMemberInfo.levelId,
      level_name: selectedMemberInfo.levelName
    })
    ElMessage.success('创建成功')
    showCreate.value = false
    fetchData()
  } catch {} finally { saving.value = false }
}
function editFee(row: any) {
  editForm.id = row.id
  editForm.amount = row.amount
  editForm.status = row.status
  editForm.remark = row.remark || ''
  showEdit.value = true
}
async function saveEdit() {
  try {
    await adminApi.updateFee(editForm.id, { amount: editForm.amount, status: editForm.status })
    ElMessage.success('保存成功')
    showEdit.value = false
    fetchData()
  } catch {}
}
async function markPaid(row: any) {
  try {
    const { value } = await ElMessageBox.prompt(`确定将会费（¥${row.amount?.toFixed(2)} - ${row.year}年）标记为已缴费？`, '确认缴费', {
      type: 'warning',
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      inputPlaceholder: '请填写备注说明原因',
      inputValidator: (v: string) => { if (!v) return '备注不能为空'; return true }
    })
    await adminApi.updateFee(row.id, { status: 'paid', remark: value })
    ElMessage.success('已标记为已缴费')
    fetchData()
  } catch {}
}
async function deleteFee(row: any) {
  try {
    await ElMessageBox.confirm(
      `确定要删除该费用记录吗？<br><small>会员：${row.member?.username || '-'} · ${row.year}年 · ¥${row.amount?.toFixed(2)}</small>`,
      '确认删除',
      { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消', dangerouslyUseHTMLString: true }
    )
    await adminApi.deleteFee(row.id)
    ElMessage.success('已删除')
    fetchData()
  } catch {}
}
</script>

<style scoped lang="scss">
.admin-fees {
  max-width: 1400px;
  margin: 0 auto;
  padding: 24px 0;
}

// ─── Header ───
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
    h3 {
      font-size: 22px;
      font-weight: 600;
      color: #1a1a2e;
      display: flex;
      align-items: center;
      gap: 8px;
      margin: 0;
      .el-icon { color: #4361ee; }
    }
    .header-subtitle {
      font-size: 14px;
      color: #94a3b8;
    }
  }
}

// ─── Stats Cards ───
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}
.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 24px;
  border-radius: 12px;
  background: #fff;
  border: 1px solid #eef2f6;
  box-shadow: 0 1px 3px rgba(0,0,0,.04);
  transition: transform .2s, box-shadow .2s;
  &:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,.08); }
  .stat-icon {
    width: 52px; height: 52px;
    display: flex; align-items: center; justify-content: center;
    border-radius: 12px;
  }
  .stat-body {
    display: flex; flex-direction: column;
    .stat-label { font-size: 13px; color: #94a3b8; margin-bottom: 4px; }
    .stat-value { font-size: 22px; font-weight: 700; color: #1a1a2e; }
  }
  &--total .stat-icon { background: #eef2ff; color: #4361ee; }
  &--paid .stat-icon { background: #ecfdf5; color: #10b981; }
  &--unpaid .stat-icon { background: #fffbeb; color: #f59e0b; }
  &--amount .stat-icon { background: #f0fdf4; color: #22c55e; }
}

// ─── Filter Card ───
.filter-card {
  margin-bottom: 16px;
  border-radius: 10px;
  :deep(.el-card__body) { padding: 14px 20px; }
}
.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  .filter-hint {
    margin-left: auto;
    font-size: 13px;
    color: #94a3b8;
  }
}

// ─── Table Card ───
.table-card {
  border-radius: 10px;
  :deep(.el-card__body) { padding: 0; }
}
.cell-member {
  display: flex;
  flex-direction: column;
  gap: 2px;
  .member-name {
    font-weight: 500;
    color: #1a1a2e;
    line-height: 1.4;
  }
  .member-unit {
    font-size: 12px;
    color: #94a3b8;
    line-height: 1.3;
  }
}
.cell-amount {
  font-weight: 600;
  color: #1a1a2e;
  font-variant-numeric: tabular-nums;
}
.cell-time {
  color: #64748b;
  font-size: 13px;
}
.action-btns {
  display: flex;
  gap: 2px;
  .el-button { padding: 4px 8px; }
}

// ─── Pagination ───
.pagination {
  display: flex;
  justify-content: center;
  padding: 20px 0;
}

// ─── Dialogs ───
.fee-dialog {
  :deep(.el-dialog__body) { padding-top: 12px; }
}
.fee-form {
  .form-hint { color: #94a3b8; font-size: 13px; }
}
.member-info-tag {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
