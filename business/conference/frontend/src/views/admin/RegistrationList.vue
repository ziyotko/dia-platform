<template>
  <div>
    <el-card>
      <el-form :inline="true">
        <el-form-item><el-select v-model="filterMeetingId" placeholder="选择会议" clearable @change="load"><el-option v-for="m in meetings" :key="m.id" :label="m.title" :value="m.id" /></el-select></el-form-item>
        <el-form-item><el-select v-model="filterStatus" placeholder="状态" clearable @change="load"><el-option label="待审核" value="pending" /><el-option label="已通过" value="approved" /><el-option label="已驳回" value="rejected" /><el-option label="候补" value="waitlist" /></el-select></el-form-item>
      </el-form>
      <el-table :data="list" stripe>
        <el-table-column prop="user.realName" label="姓名" />
        <el-table-column prop="user.phone" label="手机号" />
        <el-table-column prop="meeting.title" label="会议" />
        <el-table-column label="状态" width="100"><template #default="{row}"><el-tag :type="st(row.status)" size="small">{{ sl(row.status) }}</el-tag></template></el-table-column>
        <el-table-column label="时间" width="160"><template #default="{row}">{{ row.createdAt?.slice(0,16) }}</template></el-table-column>
        <el-table-column label="操作" width="180"><template #default="{row}">
          <el-button size="small" type="success" v-if="row.status==='pending'||row.status==='waitlist'" @click="approve(row)">通过</el-button>
          <el-button size="small" type="danger" v-if="row.status==='pending'" @click="reject(row)">驳回</el-button>
          <el-button size="small" type="warning" v-if="row.status==='waitlist'" @click="promote(row)">转正</el-button>
        </template></el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([]); const meetings = ref<any[]>([]); const page = ref(1); const size = 10; const total = ref(0)
const filterMeetingId = ref(''); const filterStatus = ref('')

function st(s: string) { const m: any = { approved:'success', pending:'warning', waitlist:'info', rejected:'danger' }; return m[s] || 'info' }
function sl(s: string) { const m: any = { approved:'已通过', pending:'待审核', waitlist:'候补', rejected:'已驳回' }; return m[s] || s }

async function load() {
  try {
    const res = await adminApi.getRegistrations({ page: page.value, pageSize: size, meetingId: filterMeetingId.value, status: filterStatus.value })
    list.value = res.data?.list || []; total.value = res.data?.total || 0
  } catch (e) {}
}

onMounted(async () => {
  try { const res = await adminApi.getMeetings({ pageSize: 1000 }); meetings.value = res.data?.list || [] } catch (e) {}
  load()
})

async function approve(row: any) { try { await adminApi.approveRegistration(row.id); ElMessage.success('已通过'); load() } catch (e: any) {} }
async function reject(row: any) {
  try { const comment = prompt('驳回原因：'); if (comment) { await adminApi.rejectRegistration(row.id, { comment }); ElMessage.success('已驳回'); load() } } catch (e) {}
}
async function promote(row: any) { try { await adminApi.promoteWaitlist(row.id); ElMessage.success('已转正'); load() } catch (e: any) {} }
</script>
