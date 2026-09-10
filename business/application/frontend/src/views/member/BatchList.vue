<template>
  <div class="page-card">
    <div class="page-toolbar">
      <el-input v-model="keyword" placeholder="搜索项目名称" clearable style="width:260px" @keyup.enter="fetch" @clear="fetch" />
      <el-button type="primary" @click="fetch">查询</el-button>
    </div>

    <el-table :data="list" v-loading="loading">
      <el-table-column prop="title" label="项目名称" min-width="220" />
      <el-table-column label="项目类别" width="140">
        <template #default="{ row }">{{ row.category?.name || '-' }}</template>
      </el-table-column>
      <el-table-column label="申报截止" width="160">
        <template #default="{ row }">{{ fmt(row.applyEnd) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }"><el-tag :type="batchStatusType[row.status]">{{ batchStatusMap[row.status] }}</el-tag></template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="goApply(row)">立即申报</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      style="margin-top:16px;justify-content:flex-end"
      layout="total, prev, pager, next"
      :total="total"
      :page-size="pageSize"
      :current-page="page"
      @current-change="onPage"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { memberApi } from '@/api/member'
import { batchStatusMap, batchStatusType, fmt } from '@/utils/constants'

const router = useRouter()
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const keyword = ref('')
const loading = ref(false)

async function fetch() {
  loading.value = true
  try {
    const res = await memberApi.getOpenBatches({ page: page.value, pageSize, keyword: keyword.value })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }
function goApply(row: any) {
  router.push({ path: '/member/applications/create', query: { batchId: row.id } })
}

onMounted(fetch)
</script>
