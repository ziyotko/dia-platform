<template>
  <div>
    <el-card>
      <el-form :inline="true">
        <el-form-item><el-select v-model="filterMeetingId" placeholder="选择会议" clearable @change="load"><el-option v-for="m in meetings" :key="m.id" :label="m.title" :value="m.id" /></el-select></el-form-item>
        <el-form-item><el-select v-model="filterStatus" placeholder="状态" clearable @change="load"><el-option label="未支付" value="unpaid" /><el-option label="已支付" value="paid" /><el-option label="退款中" value="refunding" /><el-option label="已退款" value="refunded" /></el-select></el-form-item>
      </el-form>
      <el-table :data="list" stripe>
        <el-table-column prop="orderNo" label="订单号" width="180" />
        <el-table-column prop="user.realName" label="用户" />
        <el-table-column prop="meeting.title" label="会议" />
        <el-table-column prop="amount" label="金额" width="100"><template #default="{row}">¥{{ row.amount }}</template></el-table-column>
        <el-table-column label="状态" width="80"><template #default="{row}"><el-tag :type="st(row.status)" size="small">{{ sl(row.status) }}</el-tag></template></el-table-column>
        <el-table-column label="时间" width="160"><template #default="{row}">{{ row.createdAt?.slice(0,16) }}</template></el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>

    <el-card style="margin-top:16px">
      <h4>退款申请</h4>
      <el-table :data="refunds" stripe>
        <el-table-column prop="user.realName" label="用户" />
        <el-table-column prop="amount" label="金额" width="100" />
        <el-table-column prop="reason" label="原因" />
        <el-table-column label="操作" width="160"><template #default="{row}">
          <el-button size="small" type="success" @click="process(row,true)">通过</el-button>
          <el-button size="small" type="danger" @click="process(row,false)">拒绝</el-button>
        </template></el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([]); const refunds = ref<any[]>([]); const meetings = ref<any[]>([])
const page = ref(1); const size = 10; const total = ref(0); const filterMeetingId = ref(''); const filterStatus = ref('')

function st(s: string) { const m: any = { paid:'success', unpaid:'warning', refunding:'info', refunded:'info' }; return m[s] || 'info' }
function sl(s: string) { const m: any = { paid:'已支付', unpaid:'未支付', refunding:'退款中', refunded:'已退款' }; return m[s] || s }

async function load() {
  try { const res = await adminApi.getOrders({ page: page.value, pageSize: size, meetingId: filterMeetingId.value, status: filterStatus.value }); list.value = res.data?.list || []; total.value = res.data?.total || 0 } catch (e) {}
  try { const res = await adminApi.getRefunds({ status: 'pending' }); refunds.value = res.data?.list || [] } catch (e) {}
}

async function process(row: any, approved: boolean) {
  try { await adminApi.processRefund(row.id, { approved }); ElMessage.success(approved ? '已通过' : '已拒绝'); load() } catch (e: any) {}
}

onMounted(async () => {
  try { const res = await adminApi.getMeetings({ pageSize: 1000 }); meetings.value = res.data?.list || [] } catch (e) {}
  load()
})
</script>
