<template>
  <div class="page-card">
    <div class="page-toolbar">
      <div style="display:flex;gap:12px;flex-wrap:wrap">
        <el-input v-model="keyword" placeholder="搜索项目名称" clearable style="width:200px" @keyup.enter="search" @clear="search" />
        <el-select v-model="batchId" placeholder="全部批次" clearable filterable style="width:220px" @change="search">
          <el-option v-for="b in batches" :key="b.id" :label="b.title" :value="b.id" />
        </el-select>
        <el-select v-model="reviewerId" placeholder="全部评审人" clearable filterable style="width:180px" @change="search">
          <el-option v-for="r in reviewers" :key="r.id" :label="r.realName || r.username" :value="r.id" />
        </el-select>
        <el-select v-model="status" placeholder="全部状态" clearable style="width:140px" @change="search">
          <el-option label="待评审" value="pending" />
          <el-option label="已评分" value="scored" />
        </el-select>
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </div>

    <el-table :data="list" v-loading="loading">
      <el-table-column prop="id" label="任务ID" width="90" />
      <el-table-column label="项目名称" min-width="220">
        <template #default="{ row }">{{ row.application?.title || '-' }}</template>
      </el-table-column>
      <el-table-column label="申报人" width="110">
        <template #default="{ row }">{{ row.application?.user?.realName || '-' }}</template>
      </el-table-column>
      <el-table-column label="所属批次" min-width="150">
        <template #default="{ row }">{{ row.application?.batch?.title || '-' }}</template>
      </el-table-column>
      <el-table-column label="评审人" width="120">
        <template #default="{ row }">{{ row.reviewer?.realName || row.reviewer?.username || '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'scored' ? 'success' : 'warning'" size="small">{{ reviewStatusMap[row.status] || row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="评分" width="90">
        <template #default="{ row }">{{ row.status === 'scored' ? row.score : '-' }}</template>
      </el-table-column>
      <el-table-column label="评审时间" width="170">
        <template #default="{ row }">{{ fmt(row.reviewedAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="$router.push(`/admin/applications/${row.applicationId}`)">查看申报</el-button>
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
import { adminApi } from '@/api/admin'
import { reviewStatusMap, fmt } from '@/utils/constants'

const list = ref<any[]>([])
const batches = ref<any[]>([])
const reviewers = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const keyword = ref('')
const batchId = ref<any>('')
const reviewerId = ref<any>('')
const status = ref('')
const loading = ref(false)

async function fetch() {
  loading.value = true
  try {
    const res = await adminApi.getReviewTasks({
      page: page.value,
      pageSize,
      keyword: keyword.value,
      batchId: batchId.value || undefined,
      reviewerId: reviewerId.value || undefined,
      status: status.value,
    })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function search() { page.value = 1; fetch() }
function onPage(p: number) { page.value = p; fetch() }

onMounted(async () => {
  fetch()
  const [b, r] = await Promise.all([
    adminApi.getBatches({ page: 1, pageSize: 200 }),
    adminApi.getReviewers(),
  ])
  batches.value = b.data.list
  reviewers.value = r.data
})
</script>
