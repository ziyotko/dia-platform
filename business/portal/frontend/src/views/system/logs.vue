<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="操作人">
          <el-input v-model="queryForm.username" placeholder="请输入操作人" clearable />
        </el-form-item>
        <el-form-item label="操作类型">
          <el-select v-model="queryForm.type" placeholder="全部类型" clearable style="width: 140px">
            <el-option label="新增" value="CREATE" />
            <el-option label="修改" value="UPDATE" />
            <el-option label="删除" value="DELETE" />
            <el-option label="查询" value="READ" />
            <el-option label="登录" value="LOGIN" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作时间">
          <el-date-picker
            v-model="queryForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>查询
          </el-button>
          <el-button @click="resetQuery">
            <el-icon><RefreshRight /></el-icon>重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="hover" class="table-card">
      <template #header>
        <div class="card-header">
          <span>操作日志</span>
          <el-button type="danger" plain @click="handleClear">
            <el-icon><Delete /></el-icon>清空历史日志
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="username" label="操作人" width="120" />
        <el-table-column prop="type" label="操作类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getTypeColor(row.type)" size="small">{{ getTypeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="module" label="操作模块" width="120" />
        <el-table-column prop="description" label="操作描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="ip" label="IP地址" width="140" />
        <el-table-column prop="duration" label="耗时" width="90" align="center">
          <template #default="{ row }">
            <span :class="row.duration > 1000 ? 'text-danger' : 'text-success'">{{ row.duration }}ms</span>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="操作时间" width="170" />
        <el-table-column label="操作" width="100" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">
              <el-icon><View /></el-icon>详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="queryForm.page"
          v-model:page-size="queryForm.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="detailVisible" title="日志详情" width="600px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="操作人">{{ detailData.username }}</el-descriptions-item>
        <el-descriptions-item label="操作类型">
          <el-tag :type="getTypeColor(detailData.type)" size="small">{{ getTypeLabel(detailData.type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="操作模块">{{ detailData.module }}</el-descriptions-item>
        <el-descriptions-item label="操作描述">{{ detailData.description }}</el-descriptions-item>
        <el-descriptions-item label="请求方法">{{ detailData.method }}</el-descriptions-item>
        <el-descriptions-item label="请求路径">{{ detailData.path }}</el-descriptions-item>
        <el-descriptions-item label="请求参数">
          <pre class="code-block">{{ detailData.params }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ detailData.ip }}</el-descriptions-item>
        <el-descriptions-item label="User-Agent">{{ detailData.ua }}</el-descriptions-item>
        <el-descriptions-item label="操作时间">{{ detailData.createdAt }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshRight, Delete, View } from '@element-plus/icons-vue'
import { getLogList, clearLogs } from '@/api/log'
import { buildClearLogsConfirmText, buildClearLogsSuccessText } from '@/utils/format'

const loading = ref(false)
const detailVisible = ref(false)
const total = ref(0)

const queryForm = reactive({
  page: 1,
  pageSize: 10,
  username: '',
  type: '',
  dateRange: [] as string[]
})

const detailData = reactive({
  username: '',
  type: '',
  module: '',
  description: '',
  method: '',
  path: '',
  params: '',
  ip: '',
  ua: '',
  createdAt: ''
})

const tableData = ref<any[]>([])

const typeMap: Record<string, { label: string; color: string }> = {
  CREATE: { label: '新增', color: 'success' },
  UPDATE: { label: '修改', color: 'warning' },
  DELETE: { label: '删除', color: 'danger' },
  READ: { label: '查询', color: 'info' },
  LOGIN: { label: '登录', color: 'primary' }
}

const getTypeLabel = (type: string) => typeMap[type]?.label || type
const getTypeColor = (type: string) => typeMap[type]?.color || 'info'

const handleSearch = () => {
  queryForm.page = 1
  fetchData()
}

const resetQuery = () => {
  queryForm.username = ''
  queryForm.type = ''
  queryForm.dateRange = []
  queryForm.page = 1
  fetchData()
}

const fetchData = async () => {
  loading.value = true
  try {
    const params: any = {
      page: queryForm.page,
      pageSize: queryForm.pageSize,
      username: queryForm.username,
      type: queryForm.type
    }
    if (queryForm.dateRange && queryForm.dateRange.length === 2) {
      params.startDate = queryForm.dateRange[0]
      params.endDate = queryForm.dateRange[1]
    }
    const res: any = await getLogList(params)
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch (error) {
    ElMessage.error('获取日志列表失败')
  } finally {
    loading.value = false
  }
}

const handleClear = () => {
  ElMessageBox.confirm(buildClearLogsConfirmText(), '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res: any = await clearLogs()
      if (res.code === 0) {
        ElMessage.success(buildClearLogsSuccessText(res.data?.count ?? 0))
        fetchData()
      } else {
        ElMessage.error(res.message || '清空日志失败')
      }
    } catch (error) {
      ElMessage.error('清空日志失败')
    }
  })
}

const handleDetail = (row: any) => {
  Object.assign(detailData, row)
  detailVisible.value = true
}

const handleSizeChange = (val: number) => {
  queryForm.pageSize = val
  fetchData()
}

const handleCurrentChange = (val: number) => {
  queryForm.page = val
  fetchData()
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.page-container {
  .search-card {
    margin-bottom: 20px;
    border-radius: 12px;
    border: 1px solid #e6f2ff;
  }

  .table-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-weight: 600;
      color: #2c3e50;
    }
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  .text-danger {
    color: #f56c6c;
  }

  .text-success {
    color: #67c23a;
  }

  .code-block {
    background: #f5f7fa;
    padding: 12px;
    border-radius: 4px;
    font-size: 12px;
    overflow-x: auto;
    margin: 0;
  }

  :deep(.el-descriptions__label) {
    width: 100px;
    white-space: nowrap;
  }
}
</style>