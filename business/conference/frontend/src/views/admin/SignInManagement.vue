<template>
  <div>
    <el-card>
      <el-form :inline="true">
        <el-form-item><el-select v-model="filterMeetingId" placeholder="选择会议" clearable @change="load"><el-option v-for="m in meetings" :key="m.id" :label="m.title" :value="m.id" /></el-select></el-form-item>
      </el-form>
      <div v-if="stats.rate" style="margin-bottom:16px"><strong>签到 {{ stats.signed_in }}/{{ stats.total_registrations }} ({{ stats.rate }})</strong></div>
      <el-table :data="list" stripe>
        <el-table-column prop="user.realName" label="姓名" />
        <el-table-column label="签到时间" width="160"><template #default="{row}">{{ row.signInTime?.slice(0,16) }}</template></el-table-column>
        <el-table-column label="方式" width="80"><template #default="{row}">{{ row.method==='qrcode'?'扫码':'线上' }}</template></el-table-column>
        <el-table-column label="时长(分)" width="80"><template #default="{row}">{{ Math.floor((row.duration||0)/60) }}</template></el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([]); const meetings = ref<any[]>([]); const stats = ref<any>({})
const page = ref(1); const size = 10; const total = ref(0); const filterMeetingId = ref('')

async function load() {
  try {
    const res = await adminApi.getSignIns({ page: page.value, pageSize: size, meetingId: filterMeetingId.value })
    list.value = res.data?.list || []; total.value = res.data?.total || 0
    if (filterMeetingId.value) { const sr = await adminApi.getSignInStats(Number(filterMeetingId.value)); stats.value = sr.data }
  } catch (e) {}
}

onMounted(async () => {
  try { const res = await adminApi.getMeetings({ pageSize: 1000 }); meetings.value = res.data?.list || [] } catch (e) {}
  load()
})
</script>
