<template>
  <div class="page-card">
    <el-tabs v-model="tab" @tab-change="onTabChange">
      <el-tab-pane label="公示公告" name="announcement" />
      <el-tab-pane label="评审结果" name="result" />
    </el-tabs>

    <!-- 文本公示公告 -->
    <template v-if="tab === 'announcement'">
      <el-table :data="list" v-loading="loading">
        <el-table-column prop="title" label="公示标题" min-width="240" />
        <el-table-column label="所属批次" min-width="180">
          <template #default="{ row }">{{ row.batch?.title || '-' }}</template>
        </el-table-column>
        <el-table-column label="发布时间" width="170">
          <template #default="{ row }">{{ fmt(row.publishedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="view(row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>
    </template>

    <!-- 逐条评审结果公示：与公告同页展示，避免两种公示各看一半 -->
    <template v-else>
      <el-table :data="results" v-loading="loading">
        <el-table-column prop="title" label="项目名称" min-width="220" />
        <el-table-column label="申报人" width="120">
          <template #default="{ row }">{{ row.userRealName || '-' }}</template>
        </el-table-column>
        <el-table-column label="所属批次" min-width="160">
          <template #default="{ row }">{{ row.batch?.title || '-' }}</template>
        </el-table-column>
        <el-table-column label="评审结果" width="110">
          <template #default="{ row }">
            <el-tag :type="resultType(row.status)" size="small">{{ resultText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="公示时间" width="170">
          <template #default="{ row }">{{ fmt(row.publishedAt) }}</template>
        </el-table-column>
      </el-table>
    </template>

    <el-pagination
      v-if="total"
      style="margin-top:16px;justify-content:flex-end"
      layout="total, prev, pager, next"
      :total="total" :page-size="pageSize" :current-page="page"
      @current-change="onPage"
    />

    <el-dialog v-model="dialogVisible" :title="current.title" width="700px">
      <div style="white-space:pre-wrap;line-height:1.8">{{ current.content }}</div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { memberApi } from '@/api/member'
import { fmt } from '@/utils/constants'

const tab = ref<'announcement' | 'result'>('announcement')
const list = ref<any[]>([])
const results = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const loading = ref(false)
const dialogVisible = ref(false)
const current = ref<any>({})

// 通过的走 published/certified，未通过的保持 rejected（公示时间仍会写入）
function resultText(status: string) {
  return status === 'published' || status === 'certified' ? '通过' : '未通过'
}
function resultType(status: string) {
  return status === 'published' || status === 'certified' ? 'success' : 'danger'
}

async function fetch() {
  loading.value = true
  try {
    if (tab.value === 'announcement') {
      const res = await memberApi.getAnnouncements({ page: page.value, pageSize })
      list.value = res.data.list
      total.value = res.data.total
    } else {
      const res = await memberApi.getPublishedResults({ page: page.value, pageSize })
      results.value = res.data.list
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }

function onTabChange() {
  page.value = 1
  total.value = 0
  fetch()
}

function view(row: any) { current.value = row; dialogVisible.value = true }

onMounted(fetch)
</script>
