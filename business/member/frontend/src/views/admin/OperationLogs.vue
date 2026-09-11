<template>
  <div class="admin-op-logs" v-loading="loading">
    <div class="page-header">
      <h3>操作日志</h3>
      <div class="filters">
        <el-input v-model="keyword" placeholder="搜索操作人/路径/IP" clearable style="width:240px" @clear="search" @keyup.enter="search" />
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </div>

    <el-card>
      <el-table :data="list" stripe style="width: 100%" class="op-logs-table" :row-class-name="tableRowClassName" @row-click="openDetail">
        <el-table-column label="操作时间" min-width="160">
          <template #default="{ row }">{{ fmt(row.operation_at) }}</template>
        </el-table-column>
        <el-table-column prop="username" label="操作人" min-width="110">
          <template #default="{ row }">{{ row.username || '-' }}</template>
        </el-table-column>
        <el-table-column prop="method" label="方法" width="90">
          <template #default="{ row }">
            <el-tag :type="methodTag(row.method)" size="small">{{ row.method }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" min-width="220" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '成功' : '失败' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP" min-width="130" />
        <el-table-column label="耗时" width="90">
          <template #default="{ row }">{{ row.duration }}ms</template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length === 0" description="暂无操作日志" />
      <div class="pagination">
        <el-pagination background layout="prev, pager, next, total" :total="total" :page-size="size" v-model:current-page="page" @change="fetchData" />
      </div>
    </el-card>

    <!-- 日志详情弹窗 -->
    <el-dialog v-model="detailVisible" title="操作日志详情" width="720px" append-to-body>
      <el-descriptions :column="1" border v-if="detail">
        <el-descriptions-item label="操作人">{{ detail.username || '-' }}</el-descriptions-item>
        <el-descriptions-item label="方法">{{ detail.method }}</el-descriptions-item>
        <el-descriptions-item label="路径">{{ detail.path }}</el-descriptions-item>
        <el-descriptions-item label="模块">{{ detail.module || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="detail.status === 1 ? 'success' : 'danger'" size="small">{{ detail.status === 1 ? '成功' : '失败' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="IP">{{ detail.ip }}</el-descriptions-item>
        <el-descriptions-item label="耗时">{{ detail.duration }}ms</el-descriptions-item>
        <el-descriptions-item label="操作时间">{{ fmt(detail.operation_at) }}</el-descriptions-item>
        <el-descriptions-item label="请求参数">
          <pre class="log-pre">{{ detail.params || '-' }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="响应结果">
          <pre class="log-pre">{{ detail.result || '-' }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([])
const loading = ref(true)
const page = ref(1); const size = ref(10); const total = ref(0)
const keyword = ref('')

const detailVisible = ref(false)
const detail = ref<any>(null)

function fmt(d: string) { return d ? d.replace('T', ' ').slice(0, 19) : '-' }
function methodTag(m: string) {
  const map: Record<string, string> = { POST: 'primary', PUT: 'warning', DELETE: 'danger' }
  return map[m] || 'info'
}
function tableRowClassName() { return 'op-logs-row' }

onMounted(() => fetchData())

async function fetchData() {
  loading.value = true
  try {
    const res = await adminApi.getOperationLogs({
      page: page.value,
      size: size.value,
      keyword: keyword.value
    })
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch {} finally { loading.value = false }
}

function search() { page.value = 1; fetchData() }

function openDetail(row: any) {
  detail.value = row
  detailVisible.value = true
}
</script>

<style scoped lang="scss">
.admin-op-logs {
  width: 100%;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 12px;
  h3 {
    font-size: 22px;
    font-weight: 600;
    color: #1d2739;
    margin: 0;
  }
}
.filters { display: flex; gap: 12px; }
.el-card { border-radius: 10px; }
.pagination { display: flex; justify-content: center; padding: 20px 0; }
.op-logs-table :deep(.op-logs-row) {
  cursor: pointer;
}
.op-logs-table :deep(.op-logs-row:hover td) {
  background: #f5f7fb !important;
}
.log-pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 200px;
  overflow: auto;
  font-size: 12px;
  color: #606266;
  background: #fafbfc;
  padding: 8px;
  border-radius: 6px;
}
</style>
