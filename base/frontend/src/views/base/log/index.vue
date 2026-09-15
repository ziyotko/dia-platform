<template>
  <div class="log-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>审计日志</span>
          <div class="header-actions">
            <el-button v-if="can('base:operation-log:delete')" type="danger" :disabled="selectedIds.length === 0" @click="handleBatchDelete">批量删除</el-button>
            <el-button v-if="can('base:operation-log:clear')" type="warning" @click="handleClear">清理</el-button>
            <el-button v-if="can('base:operation-log:export')" type="success" @click="handleExport">导出</el-button>
          </div>
        </div>
      </template>

      <el-form :inline="true" :model="query" class="search-form">
        <el-form-item label="用户名">
          <el-input v-model="query.username" placeholder="请输入用户名" clearable />
        </el-form-item>
        <el-form-item label="模块">
          <el-input v-model="query.module" placeholder="请输入模块" clearable />
        </el-form-item>
        <el-form-item label="方法">
          <el-select v-model="query.method" placeholder="请选择" clearable style="width: 120px">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" placeholder="请选择" clearable style="width: 120px">
            <el-option label="成功" :value="1" />
            <el-option label="失败" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间">
          <el-date-picker
            v-model="dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DD HH:mm:ss"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchData">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" border @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="module" label="模块" width="200" show-overflow-tooltip />
        <el-table-column prop="action" label="操作" width="200" show-overflow-tooltip />
        <el-table-column prop="method" label="方法" width="80" />
        <el-table-column prop="path" label="路径" min-width="200" show-overflow-tooltip />
        <el-table-column prop="ip" label="IP" width="140" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '成功' : '失败' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="duration" label="耗时(ms)" width="100" />
        <el-table-column prop="operationAt" label="操作时间" width="180" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.size"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <el-dialog v-model="detailVisible" title="日志详情" width="700px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="用户名">{{ currentLog.username }}</el-descriptions-item>
        <el-descriptions-item label="模块">{{ currentLog.module }}</el-descriptions-item>
        <el-descriptions-item label="操作">{{ currentLog.action }}</el-descriptions-item>
        <el-descriptions-item label="方法">{{ currentLog.method }}</el-descriptions-item>
        <el-descriptions-item label="路径">{{ currentLog.path }}</el-descriptions-item>
        <el-descriptions-item label="IP">{{ currentLog.ip }}</el-descriptions-item>
        <el-descriptions-item label="请求参数">
          <pre>{{ formatJson(currentLog.params) }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="响应结果">
          <pre>{{ formatJson(currentLog.result) }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getLogList, deleteLogs, clearLogs, exportLogs } from '@/api/log'
import type { OperationLog, LogQuery } from '@/api/log'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
// 按钮级权限：与后端 base:operation-log:* 权限点对齐
const can = (code: string) => userStore.can(code)

const loading = ref(false)
const tableData = ref<OperationLog[]>([])
const total = ref(0)
const selectedIds = ref<number[]>([])
const dateRange = ref<[string, string] | null>(null)
const detailVisible = ref(false)
const currentLog = ref<Partial<OperationLog>>({})

const defaultQuery: LogQuery = {
  page: 1,
  size: 10,
  username: '',
  module: '',
  method: '',
  status: undefined,
  startAt: '',
  endAt: ''
}

const query = reactive<LogQuery>({ ...defaultQuery })

const fetchData = async () => {
  loading.value = true
  try {
    if (dateRange.value) {
      query.startAt = dateRange.value[0]
      query.endAt = dateRange.value[1]
    } else {
      query.startAt = ''
      query.endAt = ''
    }
    const res: any = await getLogList(query)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  Object.assign(query, defaultQuery)
  dateRange.value = null
  fetchData()
}

const handleSelectionChange = (rows: OperationLog[]) => {
  selectedIds.value = rows.map((row) => row.id)
}

const handleBatchDelete = () => {
  ElMessageBox.confirm('确认删除选中的日志？', '提示', { type: 'warning' }).then(async () => {
    await deleteLogs(selectedIds.value)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleClear = () => {
  ElMessageBox.prompt('请输入保留天数，早于该日期的日志将被清理', '清理日志', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputPattern: /^[1-9]\d*$/,
    inputErrorMessage: '请输入正整数',
    inputValue: '30'
  }).then(async ({ value }) => {
    await clearLogs(Number(value))
    ElMessage.success('清理成功')
    fetchData()
  })
}

const handleExport = async () => {
  try {
    const res: any = await exportLogs({
      username: query.username,
      module: query.module,
      method: query.method,
      path: query.path,
      status: query.status,
      startAt: query.startAt,
      endAt: query.endAt
    })
    const blob = new Blob([res], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `operation_logs_${new Date().getTime()}.csv`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

const handleDetail = (row: OperationLog) => {
  currentLog.value = row
  detailVisible.value = true
}

const formatJson = (str?: string) => {
  if (!str) return ''
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

onMounted(fetchData)
</script>

<style scoped lang="scss">
.log-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .header-actions {
    display: flex;
    gap: 10px;
  }
  .search-form {
    margin-bottom: 20px;
  }
  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
  pre {
    margin: 0;
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 200px;
    overflow: auto;
  }
}
</style>
