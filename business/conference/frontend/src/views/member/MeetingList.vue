<template>
  <div>
    <el-card><el-tabs v-model="activeTab" @tab-change="loadData">
      <el-tab-pane label="可参加会议" name="open" />
      <el-tab-pane label="已结束会议" name="closed" />
    </el-tabs>
    <el-row :gutter="16">
      <el-col :span="8" v-for="item in list" :key="item.id" style="margin-bottom:16px">
        <el-card shadow="hover" class="meeting-card" @click="$router.push(`/member/meetings/${item.id}`)">
          <el-tag :type="item.type==='online'?'success':item.type==='offline'?'':'warning'" size="small">{{ item.type==='online'?'线上':item.type==='offline'?'线下':'混合' }}</el-tag>
          <h3 style="margin:12px 0">{{ item.title }}</h3>
          <p><el-icon><Clock /></el-icon> {{ formatTime(item.startTime) }} <template v-if="item.endTime">~ {{ formatTime(item.endTime) }}</template></p>
          <p v-if="item.location"><el-icon><Location /></el-icon> {{ item.location }}</p>
          <p><el-icon><User /></el-icon> {{ item.capacity || '不限' }}人 {{ item.fee > 0 ? `¥${item.fee}` : '免费' }}</p>
        </el-card>
      </el-col>
    </el-row>
    <el-empty v-if="list.length===0" description="暂无会议" />
    <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev, pager, next" @current-change="loadData" style="margin-top:16px" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { memberApi } from '@/api/member'

const activeTab = ref('open')
const list = ref<any[]>([])
const page = ref(1)
const size = 10
const total = ref(0)

function formatTime(t: string) { return (t || '').replace('T', ' ').slice(0, 16) }

async function loadData() {
  try {
    const res = await memberApi.getMeetings({ page: page.value, pageSize: size, status: activeTab.value === 'open' ? 'open' : 'closed' })
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {}
}

loadData()
</script>

<style scoped>
.meeting-card { cursor: pointer; }
.meeting-card p { color: #666; font-size: 13px; margin: 4px 0; display: flex; align-items: center; gap: 4px; }
</style>
