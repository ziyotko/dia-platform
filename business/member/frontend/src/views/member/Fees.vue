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
          <span>协会账户信息：{{ siteInfo.bank_name }} | 账号：{{ siteInfo.bank_account }} | 户名：{{ siteInfo.bank_account_name }}</span>
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
            <el-tag :type="row.status === 'paid' ? 'success' : 'warning'">
              {{ row.status === 'paid' ? '已缴费' : '未缴费' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="paid_at" label="缴费时间" width="170">
          <template #default="{ row }">{{ formatDate(row.paid_at) || '-' }}</template>
        </el-table-column>
        <el-table-column prop="transaction_id" label="交易流水号" />
        <el-table-column prop="invoice_no" label="发票号" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button v-if="row.status === 'unpaid'" type="primary" size="small" @click="payFee(row)">缴费</el-button>
            <span v-else style="color:#9ca3af">-</span>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && fees.length === 0" description="暂无会费记录" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { feeApi } from '@/api/index'
import { authApi } from '@/api/auth'
import { ElMessage, ElMessageBox } from 'element-plus'

const fees = ref<any[]>([])
const loading = ref(true)
const filterYear = ref<number | ''>('')
const siteInfo = reactive<any>({})
const years = ref<number[]>([])

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

async function payFee(row: any) {
  try {
    await ElMessageBox.confirm(`确认缴纳 ${row.year} 年会费 ¥${row.amount.toFixed(2)}？`, '确认缴费', {
      confirmButtonText: '确认缴费',
      type: 'warning'
    })
    await feeApi.payFee(row.id)
    ElMessage.success('缴费成功')
    fetchData()
  } catch {}
}

function formatDate(d: string) { return d ? d.slice(0, 16) : '' }
</script>

<style scoped lang="scss">
.fees-page { max-width: 1100px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.bank-info { margin-bottom: 16px; }
</style>
