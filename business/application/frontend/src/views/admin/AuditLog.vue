<template>
  <div class="page-card">
    <div class="page-toolbar">
      <el-input v-model="keyword" placeholder="搜索操作/管理员/详情" clearable style="width:240px" @keyup.enter="fetch" @clear="fetch" />
      <el-select v-model="module" placeholder="全部模块" clearable style="width:160px" @change="onFilterChange">
        <el-option v-for="m in modules" :key="m" :label="m" :value="m" />
      </el-select>
      <el-button type="primary" @click="fetch">查询</el-button>
      <el-button :loading="exporting" @click="exportCsv">导出 CSV</el-button>
    </div>

    <el-table :data="list" v-loading="loading">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="admin" label="操作人" width="120" />
      <el-table-column prop="module" label="模块" width="110" />
      <el-table-column prop="action" label="操作" min-width="150" />
      <el-table-column prop="detail" label="详情" min-width="220" show-overflow-tooltip>
        <template #default="{ row }">{{ row.detail || '-' }}</template>
      </el-table-column>
      <el-table-column prop="ip" label="IP" width="130" />
      <el-table-column label="时间" width="170">
        <template #default="{ row }">{{ fmt(row.createdAt) }}</template>
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
import { fmt } from '@/utils/constants'
import { exportAuditLogs } from '@/utils/file'

// 与后端 recordAudit 的 module 取值保持一致（新增模块时同步这里）
const modules = [
  '批次管理',
  '类别管理',
  '初审管理',
  '评审管理',
  '专家库',
  '专家评审',
  '结果公示',
  '证书管理',
  '通知管理',
  '用户管理',
  '账号管理',
  '系统日志',
]

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const keyword = ref('')
const module = ref('')
const loading = ref(false)
const exporting = ref(false)

async function fetch() {
  loading.value = true
  try {
    const res = await adminApi.getAuditLogs({ page: page.value, pageSize, keyword: keyword.value, module: module.value })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }

// 切换筛选条件回到第 1 页，否则在第 2 页切筛选会看到空列表
function onFilterChange() { page.value = 1; fetch() }

// 导出走后端鉴权接口（最多 10000 条，带当前筛选条件）
async function exportCsv() {
  exporting.value = true
  try {
    await exportAuditLogs({ module: module.value, keyword: keyword.value })
  } finally {
    exporting.value = false
  }
}

onMounted(fetch)
</script>
