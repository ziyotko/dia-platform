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
      <el-table-column label="申报时间" width="300">
        <template #default="{ row }">{{ fmt(row.applyStart) }} ~ {{ fmt(row.applyEnd) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }"><el-tag :type="batchStatusType[row.status]">{{ batchStatusMap[row.status] || row.status }}</el-tag></template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="viewDetail(row)">申报要求</el-button>
          <el-button size="small" type="primary" :disabled="!row.canApply" @click="goApply(row)">立即申报</el-button>
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

    <el-dialog v-model="detailVisible" :title="current.title" width="640px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="项目类别">{{ current.category?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="当前状态">{{ batchStatusMap[current.status] || current.status }}</el-descriptions-item>
        <el-descriptions-item label="申报时间">{{ fmt(current.applyStart) }} ~ {{ fmt(current.applyEnd) }}</el-descriptions-item>
      </el-descriptions>
      <el-alert
        v-if="!current.canApply"
        style="margin-top:12px"
        type="info"
        :closable="false"
        :title="`该批次当前${batchStatusMap[current.status] || current.status}，已不接受新的申报`"
      />
      <div class="detail-block">
        <h4>申报要求</h4>
        <div class="detail-text">{{ current.requirements || '无' }}</div>
      </div>
      <div class="detail-block">
        <h4>批次说明</h4>
        <div class="detail-text">{{ current.description || '无' }}</div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button type="primary" :disabled="!current.canApply" @click="goApply(current)">立即申报</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
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
    const res = await memberApi.getBatches({ page: page.value, pageSize, keyword: keyword.value })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }

const detailVisible = ref(false)
const current = ref<any>({})

function viewDetail(row: any) {
  current.value = row
  detailVisible.value = true
}

function goApply(row: any) {
  if (!row.canApply) return ElMessage.warning('该批次已不接受新的申报')
  router.push({ path: '/member/applications/create', query: { batchId: row.id } })
}

onMounted(fetch)
</script>

<style scoped>
.detail-block { margin-top: 16px; }
.detail-block h4 { margin-bottom: 8px; color: #374151; font-size: 14px; }
.detail-text { white-space: pre-wrap; line-height: 1.8; color: #4b5563; font-size: 14px; }
</style>
