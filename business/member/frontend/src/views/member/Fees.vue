<template>
  <div class="fees-page" v-loading="loading">
    <div class="page-header">
      <h3>会费管理</h3>
      <el-select v-model="filterYear" placeholder="筛选年度" clearable style="width:140px" @change="fetchData">
        <el-option v-for="y in years" :key="y" :label="y + '年'" :value="y" />
      </el-select>
    </div>

    <el-card>
      <!-- Bank Info -->
      <el-alert type="info" :closable="false" show-icon class="bank-info" v-if="siteInfo">
        <template #title>
          <span>协会账户信息：{{ siteInfo.bank_name }} | 账号：{{ siteInfo.bank_account }} | 户名：{{ siteInfo.bank_account_name }}，汇款的时候，备注栏注明“XX 年度会费‌”及‌单位名称</span>
        </template>
      </el-alert>

      <el-table :data="fees" stripe>
        <el-table-column prop="org_name" label="入会信息" min-width="140" />
        <el-table-column prop="level_name" label="会员级别" width="110" />
        <el-table-column prop="year" label="年度" width="100">
          <template #default="{ row }">{{ row.year }}年</template>
        </el-table-column>
        <el-table-column prop="amount" label="金额（元）" width="120">
          <template #default="{ row }">¥{{ row.amount.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="paid_date" label="缴费日期" width="110">
          <template #default="{ row }">{{ row.paid_date || '-' }}</template>
        </el-table-column>
        <el-table-column prop="invoice_no" label="票据号码" />
        <el-table-column label="发票状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="!row.invoice_status" type="info" effect="plain" size="small">未申请</el-tag>
            <el-tag v-else-if="row.invoice_status === 'applied'" type="warning" effect="plain" size="small">已申请</el-tag>
            <el-tag v-else-if="row.invoice_status === 'issued'" type="success" effect="plain" size="small">已开票</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开票单位全称" min-width="160">
          <template #default="{ row }">{{ row.invoice_company || '-' }}</template>
        </el-table-column>
        <el-table-column label="统一社会信用代码" width="160">
          <template #default="{ row }">{{ row.invoice_tax_id || '-' }}</template>
        </el-table-column>
        <el-table-column label="开票金额" width="100">
          <template #default="{ row }">{{ row.invoice_amount ? '¥' + row.invoice_amount.toFixed(2) : '-' }}</template>
        </el-table-column>
        <el-table-column label="开票联系人" min-width="140">
          <template #default="{ row }">{{ row.invoice_contact || '-' }}</template>
        </el-table-column>
        <el-table-column prop="invoice_remark" label="开票备注" min-width="120">
          <template #default="{ row }">{{ row.invoice_remark || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button v-if="row.status === 'unpaid'" type="primary" size="small" @click="openPayDialog(row)">缴费</el-button>
            <el-button v-else-if="row.status === 'paid' && !row.invoice_status" type="success" size="small" @click="openInvoiceDialog(row)">申请开票</el-button>
            <el-tag v-else-if="row.status === 'pending'" type="warning" effect="plain" size="small">待确认</el-tag>
            <span v-else style="color:#9ca3af">-</span>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && fees.length === 0" description="暂无会费记录" />
    </el-card>

    <!-- 缴费对话框 -->
    <el-dialog v-model="showPayDialog" title="提交缴费凭证" width="500px">
      <div class="pay-info" v-if="payTarget">
        <div class="pay-info-row"><span class="pay-info-label">费用：</span>{{ payTarget.year }}年会费</div>
        <div class="pay-info-row"><span class="pay-info-label">金额：</span>¥{{ payTarget.amount?.toFixed(2) }}</div>
      </div>

      <el-form label-position="top" class="pay-form">
        <el-form-item label="缴费日期" required>
          <el-date-picker
            v-model="payDate"
            type="date"
            placeholder="选择缴费日期"
            value-format="YYYY-MM-DD"
            style="width:100%"
          />
        </el-form-item>
        <el-form-item label="上传缴费回执单" required>
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            accept=".pdf,.png,.jpg,.jpeg"
            :limit="1"
            :on-change="onFileChange"
            :file-list="fileList"
          >
            <el-button type="primary" plain>
              <el-icon><Upload /></el-icon> 选择文件
            </el-button>
            <template #tip>
              <div class="upload-tip">支持 PDF、PNG、JPG 格式</div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showPayDialog = false">取消</el-button>
        <el-button type="primary" :loading="submitting" :disabled="!canSubmit" @click="submitPay">
          提交审核
        </el-button>
      </template>
    </el-dialog>

    <!-- 申请开票对话框 -->
    <el-dialog v-model="showInvoiceDialog" title="申请开票" width="560px" :close-on-click-modal="false">
      <div class="pay-info" v-if="invoiceTarget">
        <div class="pay-info-row"><span class="pay-info-label">费用：</span>{{ invoiceTarget.year }}年会费</div>
        <div class="pay-info-row"><span class="pay-info-label">金额：</span>¥{{ invoiceTarget.amount?.toFixed(2) }}</div>
      </div>

      <el-form :model="invoiceForm" label-width="150px" class="invoice-form">
        <el-form-item label="开票单位全称" required>
          <el-input v-model="invoiceForm.invoice_company" placeholder="请输入开票单位全称" />
        </el-form-item>
        <el-form-item label="统一社会信用代码" required>
          <el-input v-model="invoiceForm.invoice_tax_id" placeholder="请输入统一社会信用代码" />
        </el-form-item>
        <el-form-item label="开票金额" required>
          <el-input-number v-model="invoiceForm.invoice_amount" :min="0" :precision="2" :step="100" style="width:100%" />
        </el-form-item>
        <el-form-item label="开票联系人及电话" required>
          <el-input v-model="invoiceForm.invoice_contact" placeholder="请输入联系人姓名和电话" />
        </el-form-item>
        <el-form-item label="开票备注">
          <el-input v-model="invoiceForm.invoice_remark" type="textarea" :rows="3" placeholder="可选填写开票备注信息" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showInvoiceDialog = false">取消</el-button>
        <el-button type="primary" :loading="invoiceSubmitting" :disabled="!canSubmitInvoice" @click="submitInvoice">
          提交开票申请
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { feeApi } from '@/api/index'
import { authApi } from '@/api/auth'
import { ElMessage } from 'element-plus'
import { Upload } from '@element-plus/icons-vue'
import type { UploadInstance, UploadFile, UploadProps } from 'element-plus'

const fees = ref<any[]>([])
const loading = ref(true)
const filterYear = ref<number | ''>('')
const siteInfo = reactive<any>({})
const years = ref<number[]>([])

// 缴费弹窗
const showPayDialog = ref(false)
const payTarget = ref<any>(null)
const payDate = ref('')
const fileList = ref<any[]>([])
const uploadRef = ref<UploadInstance>()
const uploadedUrl = ref('')
const submitting = ref(false)
const uploadFileRef = ref<File | null>(null)

const canSubmit = computed(() => !!payDate.value && (!!uploadedUrl.value || !!uploadFileRef.value))

onMounted(async () => {
  try {
    const [feeRes, siteRes] = await Promise.all([
      feeApi.getMyFees(),
      authApi.getSiteInfo()
    ])
    fees.value = feeRes.data || []
    Object.assign(siteInfo, siteRes.data)
    const currentYear = new Date().getFullYear()
    for (let y = currentYear; y >= currentYear - 5; y--) years.value.push(y)
  } catch {} finally { loading.value = false }
})

async function fetchData() {
  loading.value = true
  try {
    const params: any = {}
    if (filterYear.value) params.year = filterYear.value
    const res = await feeApi.getMyFees(params)
    fees.value = res.data || []
  } catch {} finally { loading.value = false }
}

function statusType(status: string) {
  return status === 'paid' ? 'success' : status === 'pending' ? 'warning' : 'info'
}
function statusLabel(status: string) {
  return status === 'paid' ? '已缴费' : status === 'pending' ? '待确认' : '未缴费'
}

function openPayDialog(row: any) {
  payTarget.value = row
  payDate.value = ''
  fileList.value = []
  uploadFileRef.value = null
  uploadedUrl.value = ''
  showPayDialog.value = true
}

function onFileChange(uploadFile: UploadFile) {
  if (uploadFile.raw) {
    uploadFileRef.value = uploadFile.raw
  }
}

async function submitPay() {
  if (!payTarget.value || !payDate.value) return
  submitting.value = true
  try {
    let receiptFile = uploadedUrl.value

    // Upload file if new one selected
    if (uploadFileRef.value) {
      const formData = new FormData()
      formData.append('file', uploadFileRef.value)
      const uploadRes = await authApi.upload(formData)
      receiptFile = uploadRes.data?.url || uploadRes.data || ''
    }

    if (!receiptFile) {
      ElMessage.warning('请上传缴费回执单')
      return
    }

    await feeApi.payFee(payTarget.value.id, {
      receipt_file: receiptFile,
      paid_date: payDate.value
    })
    ElMessage.success('缴费信息已提交，等待管理员确认')
    showPayDialog.value = false
    fetchData()
  } catch {} finally { submitting.value = false }
}

// ─── Invoice ───
const showInvoiceDialog = ref(false)
const invoiceTarget = ref<any>(null)
const invoiceSubmitting = ref(false)
const invoiceForm = reactive({
  invoice_company: '',
  invoice_tax_id: '',
  invoice_amount: 0,
  invoice_contact: '',
  invoice_remark: ''
})

const canSubmitInvoice = computed(() => {
  return invoiceForm.invoice_company
    && invoiceForm.invoice_tax_id
    && invoiceForm.invoice_amount > 0
    && invoiceForm.invoice_contact
})

async function openInvoiceDialog(row: any) {
  invoiceTarget.value = row
  invoiceForm.invoice_amount = row.amount || 0
  invoiceForm.invoice_remark = ''

  // 从申请人资料自动获取单位全称、社会信用代码、联系人及电话
  try {
    const res = await authApi.getProfile()
    if (res.data) {
      invoiceForm.invoice_company = res.data.company_name || ''
      invoiceForm.invoice_tax_id = res.data.credit_code || ''
      const contact = res.data.contact_person || ''
      const phone = res.data.mobile || ''
      invoiceForm.invoice_contact = [contact, phone].filter(Boolean).join(' / ')
    } else {
      invoiceForm.invoice_company = ''
      invoiceForm.invoice_tax_id = ''
      invoiceForm.invoice_contact = ''
    }
  } catch {
    invoiceForm.invoice_company = ''
    invoiceForm.invoice_tax_id = ''
    invoiceForm.invoice_contact = ''
  }

  showInvoiceDialog.value = true
}

async function submitInvoice() {
  if (!invoiceTarget.value) return
  invoiceSubmitting.value = true
  try {
    await feeApi.applyInvoice(invoiceTarget.value.id, { ...invoiceForm })
    ElMessage.success('开票申请已提交')
    showInvoiceDialog.value = false
    fetchData()
  } catch {} finally { invoiceSubmitting.value = false }
}

function formatDate(d: string) { return d ? d.slice(0, 16) : '' }
</script>

<style scoped lang="scss">
.fees-page { width: 100%; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.bank-info { margin-bottom: 16px; }

.pay-info {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 14px 16px;
  margin-bottom: 20px;
}
.pay-info-row {
  font-size: 14px;
  color: #303133;
  line-height: 1.8;
}
.pay-info-label {
  color: #909399;
  margin-right: 4px;
}
.pay-form {
  :deep(.el-form-item) {
    margin-bottom: 18px;
  }
}
.upload-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.invoice-form {
  :deep(.el-form-item) {
    margin-bottom: 20px;
  }
}
</style>
