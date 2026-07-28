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
        <el-table-column prop="org_name" label="申请入会" min-width="150" />
        <el-table-column prop="level_name" label="会员级别" width="120" />
        <el-table-column prop="year" label="年度" width="80" />
        <el-table-column prop="amount" label="金额" width="100"><template #default="{row}">¥{{ row.amount?.toFixed(2) }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="90"><template #default="{row}"><el-tag :type="row.status==='paid'?'success':'warning'">{{ row.status==='paid'?'已缴费':'未缴费' }}</el-tag></template></el-table-column>
        <el-table-column prop="paid_at" label="缴费时间" width="160"><template #default="{row}">{{ row.paid_at?.slice(0,16) || '-' }}</template></el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{row}">
            <el-button text size="small" @click="editFee(row)">编辑</el-button>
            <el-button v-if="row.status!=='paid'" text size="small" type="success" @click="markPaid(row)">标记已缴费</el-button>
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

    <el-dialog v-model="showEdit" title="编辑费用" width="420px">
      <el-form :model="editForm" size="large">
        <el-form-item label="金额" required>
          <el-input-number v-model="editForm.amount" :min="0" :precision="2" style="width:100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="editForm.status" style="width:100%">
            <el-option label="未缴费" value="unpaid" />
            <el-option label="已缴费" value="paid" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="editForm.remark" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="showEdit=false">取消</el-button><el-button type="primary" @click="saveEdit">保存</el-button></template>
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

const showEdit = ref(false)
const editForm = reactive({ id: 0, amount: 0, status: 'unpaid', remark: '' })

onMounted(() => fetchData())
async function fetchData() {
  loading.value = true
  try { const r = await adminApi.getFees({ page: page.value, size: size.value }); list.value = r.data?.list || []; total.value = r.data?.total || 0 } catch {} finally { loading.value = false }
}
async function createFee() {
  try { await adminApi.createFee({ member_id: feeForm.memberId, year: feeForm.year, amount: feeForm.amount, remark: feeForm.remark }); ElMessage.success('创建成功'); showCreate.value = false; fetchData() } catch {}
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
</script>

<style scoped lang="scss">
.admin-fees { max-width: 1100px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.pagination { display: flex; justify-content: center; margin-top: 24px; }
</style>
