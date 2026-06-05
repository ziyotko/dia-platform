<template>
  <div class="page-container">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(64, 158, 255, 0.1); color: #409eff;">
              <el-icon size="28"><DocumentChecked /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statData.generated }}</div>
              <div class="stat-label">已静态化页面</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(230, 162, 60, 0.1); color: #e6a23c;">
              <el-icon size="28"><Timer /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statData.pending }}</div>
              <div class="stat-label">待生成页面</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(103, 194, 58, 0.1); color: #67c23a;">
              <el-icon size="28"><Clock /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statData.lastTime }}</div>
              <div class="stat-label">上次生成时间</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(245, 108, 108, 0.1); color: #f56c6c;">
              <el-icon size="28"><FolderOpened /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statData.cacheSize }}</div>
              <div class="stat-label">缓存大小</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 操作区域 -->
    <el-card shadow="hover" class="action-card">
      <template #header>
        <div class="card-header">
          <span>批量操作</span>
          <el-tag v-if="generating" type="warning" effect="dark">
            <el-icon class="is-loading"><Loading /></el-icon>
            正在生成中...
          </el-tag>
        </div>
      </template>
      <div class="action-list">
        <el-button type="primary" size="large" :loading="generating" @click="handleGenerateAll">
          <el-icon><Refresh /></el-icon>
          全站静态化
        </el-button>
        <el-button type="success" size="large" :loading="generating" @click="handleGenerateHome">
          <el-icon><HomeFilled /></el-icon>
          生成首页
        </el-button>
        <el-button type="info" size="large" :loading="generating" @click="handleGenerateColumn">
          <el-icon><Menu /></el-icon>
          生成栏目页
        </el-button>
        <el-button type="info" size="large" :loading="generating" @click="handleGenerateDetail">
          <el-icon><Document /></el-icon>
          生成详情页
        </el-button>
        <el-button type="info" size="large" :loading="generating" @click="handleGenerateTopic">
          <el-icon><Collection /></el-icon>
          生成专题页
        </el-button>
        <el-button type="danger" size="large" :disabled="generating" @click="handleClearCache">
          <el-icon><Delete /></el-icon>
          清理缓存
        </el-button>
      </div>
    </el-card>

    <!-- 页面静态化状态 -->
    <el-card shadow="hover" class="table-card">
      <template #header>
        <div class="card-header">
          <span>页面静态化状态</span>
          <el-button type="primary" link @click="fetchPageList">
            <el-icon><Refresh /></el-icon>刷新
          </el-button>
        </div>
      </template>
      <el-table :data="pageList" v-loading="loading" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="name" label="页面名称" min-width="180" />
        <el-table-column prop="path" label="访问路径" min-width="200" show-overflow-tooltip />
        <el-table-column prop="type" label="页面类型" width="120">
          <template #default="{ row }">
            <el-tag :type="row.type === '首页' ? 'primary' : row.type === '文章' ? 'success' : 'info'">
              {{ row.type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="静态化状态" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === '已生成' ? 'success' : 'warning'">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="generateTime" label="生成时间" width="170" />
        <el-table-column prop="fileSize" label="文件大小" width="120" align="center" />
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" :loading="row.generating" @click="handleGenerateSingle(row)">
              <el-icon><Refresh /></el-icon>重新生成
            </el-button>
            <el-button link type="success" @click="handlePreview(row)">
              <el-icon><View /></el-icon>预览
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

    <!-- 生成日志 -->
    <el-card shadow="hover" class="log-card">
      <template #header>
        <div class="card-header">
          <span>生成日志</span>
          <el-button type="danger" link @click="clearLogs">
            <el-icon><Delete /></el-icon>清空日志
          </el-button>
        </div>
      </template>
      <el-timeline>
        <el-timeline-item
          v-for="(log, index) in logList"
          :key="index"
          :type="log.type"
          :icon="log.icon"
          :timestamp="log.time"
        >
          <div class="log-content">
            <span class="log-title">{{ log.title }}</span>
            <span class="log-detail">{{ log.detail }}</span>
          </div>
        </el-timeline-item>
      </el-timeline>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  DocumentChecked,
  Timer,
  Clock,
  FolderOpened,
  Refresh,
  HomeFilled,
  Document,
  Menu,
  Collection,
  Delete,
  View,
  Loading,
  Check,
  CircleClose,
  Warning
} from '@element-plus/icons-vue'

const loading = ref(false)
const generating = ref(false)
const total = ref(0)

const statData = reactive({
  generated: 128,
  pending: 12,
  lastTime: '2026-06-05 10:30',
  cacheSize: '256 MB'
})

const queryForm = reactive({
  page: 1,
  pageSize: 10
})

const pageList = ref<any[]>([
  { id: 1, name: '网站首页', path: '/', type: '首页', status: '已生成', generateTime: '2026-06-05 10:00', fileSize: '32 KB', generating: false },
  { id: 2, name: '关于我们', path: '/about', type: '单页', status: '已生成', generateTime: '2026-06-05 10:01', fileSize: '18 KB', generating: false },
  { id: 3, name: '新闻中心', path: '/news', type: '列表', status: '已生成', generateTime: '2026-06-05 10:02', fileSize: '45 KB', generating: false },
  { id: 4, name: '产品分类-电子产品', path: '/category/electronics', type: '分类', status: '已生成', generateTime: '2026-06-05 10:03', fileSize: '28 KB', generating: false },
  { id: 5, name: '产品分类-家居用品', path: '/category/home', type: '分类', status: '待生成', generateTime: '-', fileSize: '-', generating: false },
  { id: 6, name: '文章-2026年行业趋势分析', path: '/article/1001', type: '文章', status: '已生成', generateTime: '2026-06-05 10:05', fileSize: '52 KB', generating: false },
  { id: 7, name: '文章-新技术应用案例', path: '/article/1002', type: '文章', status: '已生成', generateTime: '2026-06-05 10:06', fileSize: '38 KB', generating: false },
  { id: 8, name: '标签-Vue', path: '/tag/vue', type: '标签', status: '待生成', generateTime: '-', fileSize: '-', generating: false },
  { id: 9, name: '标签-Go', path: '/tag/go', type: '标签', status: '已生成', generateTime: '2026-06-05 10:08', fileSize: '22 KB', generating: false },
  { id: 10, name: '联系我们', path: '/contact', type: '单页', status: '已生成', generateTime: '2026-06-05 10:09', fileSize: '15 KB', generating: false }
])

const logList = ref<any[]>([
  { type: 'success', icon: Check, time: '2026-06-05 10:30:15', title: '全站静态化完成', detail: '共生成 128 个页面，耗时 12.5 秒' },
  { type: 'primary', icon: Refresh, time: '2026-06-05 10:15:02', title: '首页重新生成', detail: '文件大小 32 KB，生成耗时 0.8 秒' },
  { type: 'warning', icon: Warning, time: '2026-06-05 09:45:30', title: '栏目页生成警告', detail: '部分栏目下无内容，已跳过空栏目页面' },
  { type: 'success', icon: Check, time: '2026-06-05 09:30:00', title: '详情页批量生成完成', detail: '共生成 56 个详情页面，耗时 8.2 秒' },
  { type: 'danger', icon: CircleClose, time: '2026-06-05 09:00:10', title: '缓存清理完成', detail: '已清理过期静态文件，释放 128 MB 空间' }
])

const fetchPageList = async () => {
  loading.value = true
  try {
    // TODO: 调用后端接口获取页面列表
    await new Promise(resolve => setTimeout(resolve, 500))
    total.value = pageList.value.length
  } finally {
    loading.value = false
  }
}

const handleSizeChange = (val: number) => {
  queryForm.pageSize = val
  fetchPageList()
}

const handleCurrentChange = (val: number) => {
  queryForm.page = val
  fetchPageList()
}

const simulateGenerate = async (title: string) => {
  generating.value = true
  try {
    // TODO: 调用后端静态化接口
    await new Promise(resolve => setTimeout(resolve, 2000))
    ElMessage.success(`${title}成功`)
    addLog('success', Check, `${title}完成`, '操作执行成功')
    fetchPageList()
  } catch (error) {
    ElMessage.error(`${title}失败`)
    addLog('danger', CircleClose, `${title}失败`, '操作执行失败')
  } finally {
    generating.value = false
  }
}

const handleGenerateAll = () => {
  ElMessageBox.confirm('确定要执行全站静态化吗？此操作可能需要较长时间。', '确认操作', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    simulateGenerate('全站静态化')
  }).catch(() => {})
}

const handleGenerateHome = () => simulateGenerate('首页生成')
const handleGenerateColumn = () => simulateGenerate('栏目页生成')
const handleGenerateDetail = () => simulateGenerate('详情页生成')
const handleGenerateTopic = () => simulateGenerate('专题页生成')

const handleClearCache = () => {
  ElMessageBox.confirm('确定要清理所有静态缓存吗？清理后需要重新生成页面。', '确认清理', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'error'
  }).then(() => {
    simulateGenerate('缓存清理')
  }).catch(() => {})
}

const handleGenerateSingle = async (row: any) => {
  row.generating = true
  try {
    // TODO: 调用单页生成接口
    await new Promise(resolve => setTimeout(resolve, 1000))
    row.status = '已生成'
    row.generateTime = new Date().toLocaleString()
    ElMessage.success(`「${row.name}」生成成功`)
  } catch (error) {
    ElMessage.error(`「${row.name}」生成失败`)
  } finally {
    row.generating = false
  }
}

const handlePreview = (row: any) => {
  window.open(row.path, '_blank')
}

const addLog = (type: string, icon: any, title: string, detail: string) => {
  logList.value.unshift({
    type,
    icon,
    time: new Date().toLocaleString(),
    title,
    detail
  })
}

const clearLogs = () => {
  ElMessageBox.confirm('确定要清空所有生成日志吗？', '确认清空', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    logList.value = []
    ElMessage.success('日志已清空')
  }).catch(() => {})
}

onMounted(() => {
  fetchPageList()
})
</script>

<style scoped lang="scss">
.page-container {
  .stat-row {
    margin-bottom: 20px;

    .stat-card {
      border-radius: 12px;
      border: 1px solid #e6f2ff;

      :deep(.el-card__body) {
        padding: 20px;
      }
    }

    .stat-content {
      display: flex;
      align-items: center;
      gap: 16px;
    }

    .stat-icon {
      width: 56px;
      height: 56px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .stat-value {
      font-size: 24px;
      font-weight: 700;
      color: #2c3e50;
      line-height: 1.2;
    }

    .stat-label {
      font-size: 13px;
      color: #909399;
      margin-top: 4px;
    }
  }

  .action-card {
    margin-bottom: 20px;
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-weight: 600;
      color: #2c3e50;
    }

    .action-list {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;
      padding: 8px 0;

      .el-button {
        min-width: 140px;
      }
    }
  }

  .table-card {
    margin-bottom: 20px;
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-weight: 600;
      color: #2c3e50;
    }

    .pagination {
      margin-top: 20px;
      display: flex;
      justify-content: flex-end;
    }
  }

  .log-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-weight: 600;
      color: #2c3e50;
    }

    .log-content {
      display: flex;
      flex-direction: column;
      gap: 4px;

      .log-title {
        font-weight: 600;
        color: #2c3e50;
      }

      .log-detail {
        font-size: 13px;
        color: #909399;
      }
    }
  }
}
</style>
