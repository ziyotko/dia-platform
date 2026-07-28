<template>
  <div class="admin-fees" v-loading="loading">
    <div class="page-header">
      <h3>会费管理</h3>
      <el-button type="primary" @click="showCreate=true">新增费用</el-button>
    </div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="member.username" label="会员" />
        <el-table-column prop="year" label="年度" width="80" />
        <el-table-column prop="amount" label="金额" width="100"><template #default="{row}">¥{{ row.amount?.toFixed(2) }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="90"><template #default="{row}"><el-tag :type="row.status==='paid'?'success':'warning'">{{ row.status==='paid'?'已缴费':'未缴费' }}</el-tag></template></el-table-column>
        <el-table-column prop="paid_at" label="缴费时间" width="160"><template #default="{row}">{{ row.paid_at?.slice(0,16) || '-' }}</template></el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{row}">
            <el-button text size="small" @click="editFee(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination"><el-pagination background layout="prev, pager, next" :total="total" :page-size="size" v-model:current-page="page" @change="fetchData" /></div>
    </el-card>

    <el-dialog v-model="showCreate" title="新增费用" width="420px">
      <el-form :model="feeForm" size="large">
        <el-form-item label="会员ID" required><el-input v-model.number="feeForm.memberId" /></el-form-item>
        <el-form-item label="年度" required><el-input v-model.number="feeForm.year" /></el-form-item>
        <el-form-item label="金额" required><el-input v-model.number="feeForm.amount" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="feeForm.remark" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="showCreate=false">取消</el-button><el-button type="primary" @click="createFee">确认</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref<any[]>([]); const loading = ref(true); const showCreate = ref(false)
const page = ref(1); const size = ref(10); const total = ref(0)
const feeForm = reactive({ memberId: 0, year: new Date().getFullYear(), amount: 2000, remark: '' })

onMounted(() => fetchData())
async function fetchData() {
  loading.value = true
  try { const r = await adminApi.getFees({ page: page.value, size: size.value }); list.value = r.data?.list || []; total.value = r.data?.total || 0 } catch {} finally { loading.value = false }
}
async function createFee() {
  try { await adminApi.createFee({ member_id: feeForm.memberId, year: feeForm.year, amount: feeForm.amount, remark: feeForm.remark }); ElMessage.success('创建成功'); showCreate.value = false; fetchData() } catch {}
}
async function editFee(row: any) {
  try { const { value } = await ElMessageBox.prompt('输入状态(unpaid/paid)', '编辑费用', { inputValue: row.status }); if (value) { await adminApi.updateFee(row.id, { status: value }); ElMessage.success('已更新'); fetchData() } } catch {}
}
</script>

<style scoped lang="scss">
.admin-fees { max-width: 1100px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.pagination { display: flex; justify-content: center; margin-top: 24px; }
</style>
