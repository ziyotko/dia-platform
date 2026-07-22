<template>
  <div class="login-log-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>登录日志</span>
          <div class="header-actions">
            <el-button type="danger" :disabled="selectedIds.length === 0" @click="handleBatchDelete">批量删除</el-button>
            <el-button type="warning" @click="handleClear">清理</el-button>
            <el-button type="success" @click="handleExport">导出</el-button>
          </div>
        </div>
      </template>

      <el-form :inline="true" :model="query" class="search-form">
        <el-form-item label="用户名">
          <el-input v-model="query.username" placeholder="请输入用户名" clearable />
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
        <el-table-column prop="ip" label="IP" width="140" />
        <el-table-column prop="agent" label="UserAgent" min-width="250" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '成功' : '失败' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="消息" min-width="150" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="登录时间" width="180" />
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
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getLoginLogList, deleteLoginLogs, clearLoginLogs, exportLoginLogs } from '@/api/login-log'
import type { LoginLog, LoginLogQuery } from '@/api/login-log'

const loading = ref(false)
const tableData = ref<LoginLog[]>([])
const total = ref(0)
const selectedIds = ref<number[]>([])
const dateRange = ref<[string, string] | null>(null)

const defaultQuery: LoginLogQuery = {
  page: 1,
  size: 10,
  username: '',
  status: undefined,
  startAt: '',
  endAt: ''
}

const query = reactive<LoginLogQuery>({ ...defaultQuery })

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
    const res: any = await getLoginLogList(query)
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

const handleSelectionChange = (rows: LoginLog[]) => {
  selectedIds.value = rows.map((row) => row.id)
}

const handleBatchDelete = () => {
  ElMessageBox.confirm('确认删除选中的登录日志？', '提示', { type: 'warning' }).then(async () => {
    await deleteLoginLogs(selectedIds.value)
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
    await clearLoginLogs(Number(value))
    ElMessage.success('清理成功')
    fetchData()
  })
}

const handleExport = async () => {
  try {
    const res: any = await exportLoginLogs({
      username: query.username,
      status: query.status,
      startAt: query.startAt,
      endAt: query.endAt
    })
    const blob = new Blob([res], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `login_logs_${new Date().getTime()}.csv`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

onMounted(fetchData)
</script>

<style scoped lang="scss">
.login-log-page {
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
}
</style>
