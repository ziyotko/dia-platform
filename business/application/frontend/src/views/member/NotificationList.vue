<template>
  <div class="page-card">
    <div class="page-toolbar">
      <div>
        <el-tag v-if="unreadCount > 0" type="danger" effect="dark">{{ unreadCount }} 条未读</el-tag>
        <el-tag v-else type="info">全部已读</el-tag>
      </div>
      <el-button type="primary" :disabled="!unreadCount" :loading="markingAll" @click="markAllRead">全部已读</el-button>
    </div>

    <el-table :data="list" v-loading="loading">
      <el-table-column prop="title" label="标题" min-width="220" />
      <el-table-column label="内容" min-width="280">
        <template #default="{ row }">{{ row.content }}</template>
      </el-table-column>
      <el-table-column label="时间" width="170">
        <template #default="{ row }">{{ fmt(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.readAt ? 'info' : 'danger'" size="small">{{ row.readAt ? '已读' : '未读' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button v-if="!row.readAt" size="small" type="primary" @click="markRead(row)">标记已读</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      style="margin-top:16px;justify-content:flex-end"
      layout="total, prev, pager, next"
      :total="total" :page-size="pageSize" :current-page="page"
      @current-change="onPage"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { memberApi } from '@/api/member'
import { fmt } from '@/utils/constants'

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const loading = ref(false)
const markingAll = ref(false)
const unreadCount = ref(0)

async function fetch() {
  loading.value = true
  try {
    const res = await memberApi.getNotifications({ page: page.value, pageSize })
    list.value = res.data.list
    total.value = res.data.total
    const unread = await memberApi.getUnreadCount()
    unreadCount.value = unread.data?.count || 0
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }

async function markRead(row: any) {
  await memberApi.markRead(row.id)
  ElMessage.success('已标记已读')
  fetch()
}

// 一次把全部未读标记为已读，不需要逐条点击
async function markAllRead() {
  markingAll.value = true
  try {
    const res = await memberApi.markAllRead()
    ElMessage.success(`已将 ${res.data?.count ?? 0} 条通知标记为已读`)
    fetch()
  } finally {
    markingAll.value = false
  }
}

onMounted(fetch)
</script>
