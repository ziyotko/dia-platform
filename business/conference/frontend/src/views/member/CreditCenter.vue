<template>
  <div>
    <el-card>
      <div style="text-align:center;padding:40px 0">
        <div style="font-size:48px;font-weight:bold;color:#2563eb">{{ totalCredits }}</div>
        <div style="color:#666">总学分</div>
      </div>
    </el-card>
    <el-card style="margin-top:16px">
      <h3>学分记录</h3>
      <el-table :data="list" stripe>
        <el-table-column prop="meeting.title" label="会议" />
        <el-table-column prop="credits" label="学分" width="100" />
        <el-table-column label="来源" width="100"><template #default="{row}">{{ row.source==='auto_assign'?'自动发放':'手动调整' }}</template></el-table-column>
        <el-table-column prop="remark" label="备注" />
        <el-table-column label="时间" width="160"><template #default="{row}">{{ row.createdAt?.slice(0,16) }}</template></el-table-column>
      </el-table>
      <el-pagination v-if="recordCount>size" v-model:current-page="page" :page-size="size" :total="recordCount" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { memberApi } from '@/api/member'

const totalCredits = ref(0)
const list = ref<any[]>([])
const page = ref(1); const size = 10; const recordCount = ref(0)

async function load() {
  try {
    const res = await memberApi.getMyCredits({ page: page.value, pageSize: size })
    totalCredits.value = res.data?.total_credits || 0
    list.value = res.data?.records || []
    recordCount.value = res.data?.record_count || 0
  } catch (e) {}
}

load()
</script>
