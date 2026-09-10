<template>
  <div class="page-card">
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

    <el-pagination
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

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const loading = ref(false)
const dialogVisible = ref(false)
const current = ref<any>({})

async function fetch() {
  loading.value = true
  try {
    const res = await memberApi.getAnnouncements({ page: page.value, pageSize })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }
function view(row: any) { current.value = row; dialogVisible.value = true }

onMounted(fetch)
</script>
