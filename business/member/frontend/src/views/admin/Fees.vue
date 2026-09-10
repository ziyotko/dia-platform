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
      <div class="stat-card stat-card--pending">
        <div class="stat-icon"><el-icon :size="28"><Clock /></el-icon></div>
        <div class="stat-body">
          <span class="stat-label">待确认</span>
          <span class="stat-value">{{ pendingCount }}</span>
        </div>
      </div>
      <div class="stat-card stat-card--amount">
        <div class="stat-icon"><el-icon :size="28"><Money /></el-icon></div>
        <div class="stat-body">
          <span class="stat-label">总金额</span>
          <span class="stat-value">¥{{ totalAmount.toFixed(2) }}</span>
        </div>
      </div>
      <div class="stat-card stat-card--paid-amount">
        <div class="stat-icon"><el-icon :size="28"><CircleCheckFilled /></el-icon></div>
        <div class="stat-body">
          <span class="stat-label">实缴总金额</span>
          <span class="stat-value">¥{{ paidAmount.toFixed(2) }}</span>
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
          <el-option label="待确认" value="pending" />
          <el-option label="未缴费" value="unpaid" />
        </el-select>
        <el-select v-model="filterType" placeholder="会员类型" clearable style="width:140px" @change="fetchData">
          <el-option label="单位会员" value="unit" />
          <el-option label="个人会员" value="personal" />
        </el-select>
        <span class="filter-hint">共 {{ total }} 条记录</span>
      </div>
    </el-card>

    <!-- Table -->
    <el-card shadow="never" class="table-card">
      <el-table :data="list" stripe highlight-current-row>
        <el-table-column type="index" label="#" width="50" align="center" />
        <el-table-column prop="member.username" label="会员信息" min-width="200">
          <template #default="{row}">
            <div class="cell-member">
              <span v-if="row.member?.member_type === 'personal'" class="member-unit member-unit--personal">
                <el-icon class="member-type-icon"><User /></el-icon>
                {{ row.member?.name || row.member?.username }}
              </span>
              <span v-else-if="row.member?.company_name" class="member-unit">{{ row.member.company_name }}</span>
              <span v-else class="member-unit">{{ row.member?.username || '-' }}</span>
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
            <el-tag v-else-if="row.status==='pending'" type="warning" effect="dark" round>
              <el-icon style="vertical-align:-2px;margin-right:3px"><Clock /></el-icon>待确认
            </el-tag>
            <el-tag v-else type="info" effect="dark" round>
              <el-icon style="vertical-align:-2px;margin-right:3px"><WarningFilled /></el-icon>未缴费
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="paid_date" label="缴费日期" width="110">
          <template #default="{row}">
            <span>{{ row.paid_date || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="paid_amount" label="实缴金额" width="110" align="right">
          <template #default="{row}">
            <span v-if="row.paid_amount > 0" class="cell-amount">¥{{ row.paid_amount?.toFixed(2) }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="confirmed_at" label="确认时间" width="170">
          <template #default="{row}">
            <span>{{ row.confirmed_at?.slice(0,16).replace('T',' ') || '-' }}</span>
          </template>
        </el-table-column>
        <!-- Invoice columns -->
        <el-table-column prop="invoice_no" label="票据号码" width="120" />
        <el-table-column label="发票状态" width="90" align="center">
          <template #default="{row}">
            <el-tag v-if="!row.invoice_status" type="info" effect="plain" size="small">未申请</el-tag>
            <el-tag v-else-if="row.invoice_status === 'applied'" type="warning" effect="plain" size="small">已申请</el-tag>
            <el-tag v-else-if="row.invoice_status === 'issued'" type="success" effect="plain" size="small">已开票</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开票单位" min-width="160">
          <template #default="{row}">{{ row.invoice_company || '-' }}</template>
        </el-table-column>
        <el-table-column label="开票金额" width="100" align="right">
          <template #default="{row}">{{ row.invoice_amount ? '¥' + row.invoice_amount.toFixed(2) : '-' }}</template>
        </el-table-column>
        <el-table-column prop="invoice_remark" label="开票备注" min-width="120">
          <template #default="{row}">{{ row.invoice_remark || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="360" fixed="right">
          <template #default="{row}">
            <div class="action-btns">
              <el-button v-if="row.status==='unpaid'" text size="small" :icon="Edit" @click="editFee(row)">编辑</el-button>
              <el-button v-if="row.receipt_file" text size="small" type="primary" :icon="Download" @click="viewReceipt(row)">缴费回执</el-button>
              <el-button v-if="row.status==='pending'" text size="small" type="success" :icon="CircleCheck" @click="confirmPay(row)">确认缴费</el-button>
              <el-tooltip
                v-if="row.status==='unpaid' && !hasLevel(row)"
                content="会员级别为空，请先修改会员级别"
                placement="top"
              >
                <el-button text size="small" type="success" :icon="CircleCheck" disabled>免缴确认</el-button>
              </el-tooltip>
              <el-button v-else-if="row.status==='unpaid'" text size="small" type="success" :icon="CircleCheck" @click="markPaid(row)">免缴确认</el-button>
              <el-button v-if="row.status==='unpaid'" text size="small" type="danger" :icon="Delete" @click="deleteFee(row)">删除</el-button>
              <el-button v-if="row.invoice_status==='applied'" text size="small" type="warning" :icon="Coin" @click="openIssueInvoice(row)">开票</el-button>
              <el-button v-if="row.invoice_status==='issued'" text size="small" type="warning" :icon="Upload" @click="openIssueInvoice(row)">重新上传发票</el-button>
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
    <el-dialog v-model="showEdit" title="编辑费用" width="560px" :close-on-click-modal="false" class="fee-dialog">
      <el-form :model="editForm" label-width="90px" class="fee-form">
        <el-form-item label="会员单位">
          <span class="form-value">{{ editForm.memberName || '-' }}</span>
        </el-form-item>
        <el-form-item label="入会信息">
          <div v-if="editForm.orgName" class="member-info-tag">
            <el-tag type="info" round>{{ editForm.orgName }}</el-tag>
          </div>
          <span v-else class="form-hint">-</span>
        </el-form-item>
        <el-form-item label="会员级别" required>
          <el-select
            v-model="editForm.levelId"
            style="width:100%"
            :loading="editLevelLoading"
            @change="onEditLevelChange"
            placeholder="选择会员级别"
          >
            <el-option
              v-for="l in editLevelOptions"
              :key="l.id"
              :label="l.name"
              :value="l.id"
            />
          </el-select>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="年度">
              <span class="form-value">{{ editForm.year }}年</span>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="金额" required>
              <el-input-number v-model="editForm.amount" :min="0" :precision="2" :step="500" style="width:100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="状态">
          <el-tag v-if="editForm.status==='paid'" type="success" effect="dark" round>
            <el-icon style="vertical-align:-2px;margin-right:3px"><CircleCheckFilled /></el-icon>已缴费
          </el-tag>
          <el-tag v-else type="warning" effect="dark" round>
            <el-icon style="vertical-align:-2px;margin-right:3px"><WarningFilled /></el-icon>未缴费
          </el-tag>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editForm.remark" type="textarea" :rows="2" placeholder="可选填写备注信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEdit=false">取消</el-button>
        <el-button type="primary" :icon="Check" @click="saveEdit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 开票对话框 / 重新上传发票 -->
    <el-dialog v-model="showInvoice" :title="invoiceTarget?.invoice_status === 'issued' ? '重新上传发票' : '开票'" width="560px" :close-on-click-modal="false">
      <div class="confirm-info" v-if="invoiceTarget">
        <div class="confirm-info-row">会员：<strong>{{ invoiceTarget.member?.company_name || invoiceTarget.member?.username }}</strong></div>
        <div class="confirm-info-row">{{ invoiceTarget.year }}年 · {{ invoiceTarget.org_name }} · {{ invoiceTarget.level_name }}</div>
        <div class="confirm-info-row" style="margin-top:8px">
          开票单位：{{ invoiceTarget.invoice_company || '-' }}<br>
          信用代码：{{ invoiceTarget.invoice_tax_id || '-' }}<br>
          开票金额：¥{{ (invoiceTarget.invoice_amount || invoiceTarget.amount)?.toFixed(2) }}<br>
          联系人：{{ invoiceTarget.invoice_contact || '-' }}<br>
          备注：{{ invoiceTarget.invoice_remark || '-' }}
          <template v-if="invoiceTarget.invoice_status === 'issued'">
            <br>当前票据号码：{{ invoiceTarget.invoice_no || '-' }}
            <br>当前发票文件：<a v-if="invoiceTarget.invoice_file" :href="invoiceTarget.invoice_file" target="_blank" style="color:#002fa7">查看</a><span v-else>-</span>
          </template>
        </div>
      </div>
      <el-form label-width="100px" class="confirm-form">
        <el-form-item label="票据号码" required>
          <el-input v-model="invoiceNo" placeholder="请输入票据号码" />
        </el-form-item>
        <el-form-item label="发票文件">
          <el-upload
            ref="invoiceUploadRef"
            :auto-upload="false"
            accept=".pdf"
            :limit="1"
            :on-change="onInvoiceFileChange"
            :file-list="invoiceFileList"
          >
            <el-button type="primary" plain>
              <el-icon><Upload /></el-icon> 选择 PDF 文件
            </el-button>
            <template #tip>
              <div class="upload-tip">仅支持 PDF 格式，如不更换文件可不选</div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showInvoice=false">取消</el-button>
        <el-button type="warning" :loading="issuingInvoice" :icon="Coin" @click="submitIssueInvoice">{{ invoiceTarget?.invoice_status === 'issued' ? '确认更新' : '确认开票' }}</el-button>
      </template>
    </el-dialog>

    <!-- 确认缴费对话框 -->
    <el-dialog v-model="showConfirm" title="确认缴费" width="480px" :close-on-click-modal="false">
      <div class="confirm-info" v-if="confirmTarget">
        <div class="confirm-info-row">会员：<strong>{{ confirmTarget.member?.company_name || confirmTarget.member?.username }}</strong></div>
        <div class="confirm-info-row">{{ confirmTarget.year }}年 · {{ confirmTarget.org_name }} · {{ confirmTarget.level_name }}</div>
      </div>
      <el-form label-width="100px" class="confirm-form">
        <el-form-item label="缴费金额" required>
          <el-input-number
            v-model="confirmAmount"
            :min="0"
            :precision="2"
            :step="100"
            style="width:100%"
          />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="confirmRemark" type="textarea" :rows="3" placeholder="可选填写备注信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showConfirm=false">取消</el-button>
        <el-button type="primary" :loading="confirming" :icon="CircleCheck" @click="submitConfirm">确认缴费</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete, CircleCheck, CircleCheckFilled, WarningFilled, Coin, List, Money, Check, Clock, Download, Upload, User } from '@element-plus/icons-vue'

const list = ref<any[]>([]); const loading = ref(true); const showCreate = ref(false)
const page = ref(1); const size = ref(10); const total = ref(0)

const memberOptions = ref<any[]>([])
const searchLoading = ref(false)
const saving = ref(false)
const yearOptions = ref<number[]>([])
const selectedMemberInfo = reactive({ orgId: 0, levelId: 0, orgName: '', levelName: '' })
const feeForm = reactive({ memberId: null as number | null, year: new Date().getFullYear(), amount: 2000, remark: '' })

const showEdit = ref(false)
const showConfirm = ref(false)
const confirmTarget = ref<any>(null)
const confirmAmount = ref(0)
const confirmRemark = ref('')
const confirming = ref(false)

// Invoice issuing (admin)
const showInvoice = ref(false)
const invoiceTarget = ref<any>(null)
const invoiceNo = ref('')
const invoiceFileList = ref<any[]>([])
const invoiceUploadRef = ref<any>(null)
const issuingInvoice = ref(false)
const editForm = reactive({ id: 0, memberId: 0, memberName: '', orgId: 0, orgName: '', levelId: null as number | null, levelName: '', year: 0, amount: 0, status: 'unpaid', remark: '' })
const editLevelOptions = ref<any[]>([])
const editLevelLoading = ref(false)

const filterYear = ref<number | null>(null)
const filterStatus = ref<string | null>(null)
const filterType = ref<string | null>(null)

const paidCount = computed(() => list.value.filter(r => r.status === 'paid').length)
const pendingCount = computed(() => list.value.filter(r => r.status === 'pending').length)
const unpaidCount = computed(() => list.value.filter(r => r.status === 'unpaid').length)
const totalAmount = computed(() => list.value.reduce((s, r) => s + (r.amount || 0), 0))
const paidAmount = computed(() => list.value.filter(r => r.status === 'paid').reduce((s, r) => s + ((r.paid_amount || 0)), 0))

onMounted(() => {
  fetchData()
  const y = new Date().getFullYear()
  // 往前 3 年 + 当年 + 往后 9 年，共 13 个年度可选
  yearOptions.value = Array.from({ length: 13 }, (_, i) => y + 3 - i)
})
async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: page.value, size: size.value }
    if (filterYear.value) params.year = filterYear.value
    if (filterStatus.value) params.status = filterStatus.value
    if (filterType.value) params.member_type = filterType.value
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
  editForm.memberId = row.member_id
  editForm.memberName = row.member?.company_name || row.member?.username || ''
  editForm.orgId = row.org_id
  editForm.orgName = row.org_name || ''
  editForm.levelId = row.level_id || null
  editForm.year = row.year
  editForm.amount = row.amount
  editForm.status = row.status
  editForm.remark = row.remark || ''
  showEdit.value = true
  // 异步加载该会员所属组织的所有级别
  fetchEditLevels(row.org_id, row.level_id)
}
async function fetchEditLevels(orgId: number, currentLevelId: number) {
  editLevelOptions.value = []
  if (!orgId) return
  editLevelLoading.value = true
  try {
    const r = await adminApi.getOrgLevels(orgId)
    // 接口返回 MemberOrgLevel[]，每个元素有 level 嵌套对象
    const items: any[] = r.data || []
    editLevelOptions.value = items.map((item: any) => item.level || item).filter(Boolean)
    // 若当前级别不在列表中，仍保留显示
  } catch {} finally { editLevelLoading.value = false }
}
async function onEditLevelChange(levelId: number) {
  // 找到级别名称
  const found = editLevelOptions.value.find((l: any) => l.id === levelId)
  editForm.levelName = found?.name || ''
  // 根据新级别 + 当前年度自动查询会费标准，填充金额
  if (!levelId || !editForm.year) return
  try {
    const r = await adminApi.getFeeStandardsByLevel(levelId)
    const standards: any[] = r.data || []
    const std = standards.find((s: any) => s.year === editForm.year)
    if (std && std.amount > 0) {
      editForm.amount = std.amount
    }
  } catch {}
}
async function saveEdit() {
  if (!editForm.levelId) { ElMessage.warning('请选择会员级别'); return }
  try {
    await adminApi.updateFee(editForm.id, {
      amount: editForm.amount,
      status: editForm.status,
      level_id: editForm.levelId || undefined,
      level_name: editForm.levelName || undefined
    })
    ElMessage.success('保存成功')
    showEdit.value = false
    fetchData()
  } catch {}
}
function hasLevel(row: any) {
  return !!row.level_name || Number(row.level_id) > 0
}
async function markPaid(row: any) {
  if (!hasLevel(row)) {
    ElMessage.warning('会员级别为空，请先联系管理员修改会员级别')
    return
  }
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
function viewReceipt(row: any) {
  if (row.receipt_file) {
    window.open(row.receipt_file, '_blank')
  }
}
async function confirmPay(row: any) {
  confirmTarget.value = row
  confirmAmount.value = row.amount || 0
  confirmRemark.value = ''
  showConfirm.value = true
}
async function submitConfirm() {
  if (!confirmTarget.value) return
  confirming.value = true
  try {
    await adminApi.confirmFee(confirmTarget.value.id, {
      amount: confirmAmount.value,
      remark: confirmRemark.value
    })
    ElMessage.success('已确认缴费')
    showConfirm.value = false
    fetchData()
  } catch {} finally { confirming.value = false }
}
// ─── Invoice issuing ───
function openIssueInvoice(row: any) {
  invoiceTarget.value = row
  invoiceNo.value = row.invoice_no || ''
  invoiceFileList.value = []
  showInvoice.value = true
}
function onInvoiceFileChange(file: any) {
  invoiceFileList.value = [file]
}
async function submitIssueInvoice() {
  if (!invoiceNo.value) { ElMessage.warning('请输入票据号码'); return }
  if (!invoiceTarget.value) return
  issuingInvoice.value = true
  try {
    const fd = new FormData()
    fd.append('invoice_no', invoiceNo.value)
    if (invoiceFileList.value.length > 0 && invoiceFileList.value[0].raw) {
      fd.append('file', invoiceFileList.value[0].raw)
    }
    await adminApi.issueInvoice(invoiceTarget.value.id, fd)
    ElMessage.success('开票成功')
    showInvoice.value = false
    fetchData()
  } catch {} finally { issuingInvoice.value = false }
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
 width: 100%;
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
      color: #1d2739;
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
  grid-template-columns: repeat(6, 1fr);
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
    .stat-value { font-size: 22px; font-weight: 700; color: #1d2739; }
  }
  &--total .stat-icon { background: #eef2ff; color: #4361ee; }
  &--paid .stat-icon { background: #ecfdf5; color: #10b981; }
  &--unpaid .stat-icon { background: #fffbeb; color: #f59e0b; }
  &--pending .stat-icon { background: #fef3c7; color: #d97706; }
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
    color: #1d2739;
    line-height: 1.4;
  }
  .member-unit {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    color: #94a3b8;
    line-height: 1.3;
    .member-type-icon {
      color: #f59e0b;
      font-size: 14px;
    }
  }
}
.cell-amount {
  font-weight: 600;
  color: #1d2739;
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
  .form-value { color: #1d2739; font-size: 14px; }
}
.member-info-tag {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.confirm-info {
  background: #f5f7fb;
  border-radius: 8px;
  padding: 14px 16px;
  margin-bottom: 20px;
  font-size: 14px;
  line-height: 1.8;
  color: #303133;
}
.confirm-form {
  :deep(.el-form-item) {
    margin-bottom: 18px;
  }
}
</style>
