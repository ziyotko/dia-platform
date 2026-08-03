<template>
  <div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="meeting.title" label="会议名称" />
        <el-table-column label="报名时间" width="160"><template #default="{row}">{{ row.createdAt?.slice(0,16) }}</template></el-table-column>
        <el-table-column label="状态" width="120">
          <template #default="{row}">
            <el-tag :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
            <span v-if="row.status==='waitlist'" style="font-size:12px;color:#999">(第{{row.waitlistPosition}}位)</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{row}"><el-button size="small" type="danger" v-if="row.status==='pending'||row.status==='waitlist'" @click="cancel(row.id)">取消</el-button></template>
        </el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { memberApi } from '@/api/member'

const list = ref<any[]>([])
const page = ref(1)
const size = 10
const total = ref(0)

function statusType(s: string) { const m: any = { approved:'success', pending:'warning', waitlist:'info', rejected:'danger', cancelled:'info' }; return m[s] || 'info' }
function statusLabel(s: string) { const m: any = { approved:'已通过', pending:'待审核', waitlist:'候补', rejected:'已驳回', cancelled:'已取消' }; return m[s] || s }

async function load() {
  try { const res = await memberApi.getMyRegistrations({ page: page.value, pageSize: size }); list.value = res.data?.list || []; total.value = res.data?.total || 0 } catch (e) {}
}

async function cancel(id: number) {
  try { await memberApi.cancelRegistration(id); ElMessage.success('已取消'); load() } catch (e: any) {}
}

load()
</script>
