<template>
  <div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="orderNo" label="订单号" width="200" />
        <el-table-column prop="meeting.title" label="会议" />
        <el-table-column prop="amount" label="金额" width="100"><template #default="{row}">¥{{ row.amount }}</template></el-table-column>
        <el-table-column label="状态" width="100"><template #default="{row}"><el-tag :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag></template></el-table-column>
        <el-table-column label="时间" width="160"><template #default="{row}">{{ row.createdAt?.slice(0,16) }}</template></el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{row}">
            <el-button size="small" type="primary" v-if="row.status==='unpaid'" @click="pay(row)">支付</el-button>
            <el-button size="small" type="warning" v-if="row.status==='paid'" @click="refund(row)">申请退款</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>

    <el-dialog v-model="payVisible" title="选择支付方式" width="400px">
      <el-radio-group v-model="payMethod"><el-radio label="wechat">微信支付</el-radio><el-radio label="alipay">支付宝</el-radio></el-radio-group>
      <template #footer><el-button @click="payVisible=false">取消</el-button><el-button type="primary" @click="confirmPay">确认支付</el-button></template>
    </el-dialog>

    <el-dialog v-model="refundVisible" title="申请退款" width="400px">
      <el-input v-model="refundReason" placeholder="退款原因" type="textarea" />
      <template #footer><el-button @click="refundVisible=false">取消</el-button><el-button type="primary" @click="confirmRefund">提交</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { memberApi } from '@/api/member'

const list = ref<any[]>([])
const page = ref(1); const size = 10; const total = ref(0)
const payVisible = ref(false); const payMethod = ref('wechat'); const payOrderId = ref(0)
const refundVisible = ref(false); const refundReason = ref(''); const refundOrderId = ref(0)

function statusType(s: string) { const m: any = { paid:'success', unpaid:'warning', refunding:'info', refunded:'info' }; return m[s] || 'info' }
function statusLabel(s: string) { const m: any = { paid:'已支付', unpaid:'未支付', refunding:'退款中', refunded:'已退款' }; return m[s] || s }

async function load() {
  try { const res = await memberApi.getMyOrders({ page: page.value, pageSize: size }); list.value = res.data?.list || []; total.value = res.data?.total || 0 } catch (e) {}
}

function pay(row: any) { payOrderId.value = row.id; payVisible.value = true }
async function confirmPay() {
  try { await memberApi.payOrder(payOrderId.value, { payMethod: payMethod.value }); ElMessage.success('支付成功'); payVisible.value = false; load() } catch (e: any) {}
}

function refund(row: any) { refundOrderId.value = row.id; refundVisible.value = true }
async function confirmRefund() {
  try { await memberApi.applyRefund(refundOrderId.value, { reason: refundReason.value }); ElMessage.success('申请已提交'); refundVisible.value = false; load() } catch (e: any) {}
}

load()
</script>
