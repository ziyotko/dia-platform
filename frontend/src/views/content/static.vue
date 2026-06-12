<template>
  <div class="page-container">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="24" :sm="12" :md="8">
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
      <el-col :xs="24" :sm="12" :md="8">
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
      <el-col :xs="24" :sm="12" :md="8">
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
        <el-button type="info" size="large" :loading="generating" @click="handleGenerateTopic">
          <el-icon><Collection /></el-icon>
          生成专题页
        </el-button>
        <el-button type="info" size="large" :loading="generating" @click="handleGenerateDetail">
          <el-icon><Document /></el-icon>
          生成详情页
        </el-button>
      </div>
    </el-card>

    <!-- Tab 切换区域 -->
    <el-card shadow="hover" class="tab-card">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- 首页 -->
        <el-tab-pane label="首页" name="home">
          <el-table :data="homePagedList" v-loading="loading" border stripe>
            <el-table-column type="index" width="60" align="center" />
            <el-table-column prop="id" label="ID" width="80" align="center" />
            <el-table-column prop="name" label="名称" min-width="160" />
            <el-table-column prop="code" label="编码" min-width="120" />
            <el-table-column prop="routePath" label="访问路径" min-width="180" show-overflow-tooltip />
            <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
            <el-table-column prop="createTime" label="创建时间" width="170" />
            <el-table-column prop="updatedAt" label="更新时间" width="170" />
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
              :total="homeTotal"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
            />
          </div>
        </el-tab-pane>

        <!-- 栏目页 -->
        <el-tab-pane label="栏目页" name="column">
          <el-table :data="columnPagedList" v-loading="loading" border stripe>
            <el-table-column type="index" width="60" align="center" />
            <el-table-column prop="id" label="ID" width="80" align="center" />
            <el-table-column prop="name" label="名称" min-width="160" />
            <el-table-column prop="code" label="编码" min-width="120" />
            <el-table-column prop="routePath" label="访问路径" min-width="180" show-overflow-tooltip />
            <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
            <el-table-column prop="createTime" label="创建时间" width="170" />
            <el-table-column prop="updatedAt" label="更新时间" width="170" />
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
              :total="columnTotal"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
            />
          </div>
        </el-tab-pane>

        <!-- 专题页 -->
        <el-tab-pane label="专题页" name="topic">
          <el-table :data="topicPagedList" v-loading="loading" border stripe>
            <el-table-column type="index" width="60" align="center" />
            <el-table-column prop="id" label="ID" width="80" align="center" />
            <el-table-column prop="name" label="名称" min-width="160" />
            <el-table-column prop="code" label="编码" min-width="120" />
            <el-table-column prop="routePath" label="访问路径" min-width="180" show-overflow-tooltip />
            <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
            <el-table-column prop="createTime" label="创建时间" width="170" />
            <el-table-column prop="updatedAt" label="更新时间" width="170" />
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
              :total="topicTotal"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
            />
          </div>
        </el-tab-pane>

        <!-- 详情页 -->
        <el-tab-pane label="详情页" name="detail">
          <el-table :data="detailPagedList" v-loading="loading" border stripe>
            <el-table-column type="index" width="60" align="center" />
            <el-table-column prop="id" label="ID" width="80" align="center" />
            <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
            <el-table-column prop="author" label="作者" min-width="60" />
            <el-table-column prop="source" label="来源" min-width="100" />
            <el-table-column prop="createTime" label="创建时间" width="170" />
            <el-table-column prop="updatedAt" label="更新时间" width="170" />
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
              :total="detailTotal"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
            />
          </div>
        </el-tab-pane>
        <el-tab-pane label="静态化日志" name="logs">
          <div class="tab-header-actions" style="justify-content: flex-end;">
            <el-button type="danger" link @click="clearLogs">
              <el-icon><Delete /></el-icon>清空日志
            </el-button>
          </div>
          <div class="log-scroll-container">
            <el-timeline v-loading="logLoading">
              <el-timeline-item
                v-for="log in logList"
                :key="log.id"
                :type="log.status"
                :icon="log.icon"
                :timestamp="log.time"
              >
                <div class="log-content">
                  <div class="log-header">
                    <span class="log-title">{{ log.operation }}</span>
                    <el-tag :type="log.status" size="small">{{ log.statusText }}</el-tag>
                  </div>
                  <div class="log-meta">
                    <span v-if="log.pageName" class="log-meta-item">
                      <el-icon size="12"><Document /></el-icon> {{ log.pageName }}
                    </span>
                    <span v-if="log.path" class="log-meta-item">
                      <el-icon size="12"><HomeFilled /></el-icon> {{ log.path }}
                    </span>
                    <span v-if="log.duration" class="log-meta-item">
                      <el-icon size="12"><Timer /></el-icon> {{ log.duration }}
                    </span>
                    <span v-if="log.fileSize" class="log-meta-item">
                      <el-icon size="12"><DocumentChecked /></el-icon> {{ log.fileSize }}
                    </span>
                    <span v-if="log.operator" class="log-meta-item">
                      <el-icon size="12"><User /></el-icon> {{ log.operator }}
                    </span>
                  </div>
                  <span class="log-detail">{{ log.message }}</span>
                </div>
              </el-timeline-item>
            </el-timeline>
          </div>
          <div class="pagination">
            <el-pagination
              v-model:current-page="logQueryForm.page"
              v-model:page-size="logQueryForm.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="logTotal"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleLogSizeChange"
              @current-change="handleLogCurrentChange"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  DocumentChecked,
  Timer,
  Clock,
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
  Warning,
  User
} from '@element-plus/icons-vue'
import { getArticleColumnPublishes } from '@/api/article'
import { getPages } from '@/api/page'
import { getStaticLogList, clearStaticLogs } from '@/api/static_log'

const loading = ref(false)
const generating = ref(false)
const activeTab = ref('home')

const statData = reactive({
  generated: 128,
  pending: 12,
  lastTime: '2026-06-05 10:30'
})

const queryForm = reactive({
  page: 1,
  pageSize: 10
})

const homeList = ref<any[]>([])
const homeTotal = computed(() => homeList.value.length)
const homePagedList = computed(() => {
  const start = (queryForm.page - 1) * queryForm.pageSize
  const end = start + queryForm.pageSize
  return homeList.value.slice(start, end)
})

const columnList = ref<any[]>([])
const columnTotal = computed(() => columnList.value.length)
const columnPagedList = computed(() => {
  const start = (queryForm.page - 1) * queryForm.pageSize
  const end = start + queryForm.pageSize
  return columnList.value.slice(start, end)
})

const detailList = ref<any[]>([])
const detailTotal = ref(0)
const detailPagedList = computed(() => detailList.value)

const topicList = ref<any[]>([])
const topicTotal = computed(() => topicList.value.length)
const topicPagedList = computed(() => {
  const start = (queryForm.page - 1) * queryForm.pageSize
  const end = start + queryForm.pageSize
  return topicList.value.slice(start, end)
})

const pageList = ref<any[]>([
  // 首页
  { id: 1, name: '网站首页', path: '/', type: '首页', fileSize: '32 KB', generating: false },
  // 栏目页
  { id: 2, name: '关于我们', path: '/about', type: '单页', fileSize: '18 KB', generating: false },
  { id: 3, name: '新闻中心', path: '/news', type: '列表', fileSize: '45 KB', generating: false },
  { id: 4, name: '产品分类-电子产品', path: '/category/electronics', type: '分类', fileSize: '28 KB', generating: false },
  { id: 5, name: '产品分类-家居用品', path: '/category/home', type: '分类', fileSize: '-', generating: false },
  { id: 8, name: '标签-Vue', path: '/tag/vue', type: '标签', fileSize: '-', generating: false },
  { id: 9, name: '标签-Go', path: '/tag/go', type: '标签', fileSize: '22 KB', generating: false },
  { id: 10, name: '联系我们', path: '/contact', type: '单页', fileSize: '15 KB', generating: false },
  // 详情页
  { id: 6, name: '文章-2026年行业趋势分析', path: '/article/1001', type: '文章', fileSize: '52 KB', generating: false },
  { id: 7, name: '文章-新技术应用案例', path: '/article/1002', type: '文章', fileSize: '38 KB', generating: false },
  // 专题页
  { id: 11, name: '年中大促专题', path: '/topic/2026-mid', type: '专题', fileSize: '48 KB', generating: false },
  { id: 12, name: '品牌故事专题', path: '/topic/brand', type: '专题', fileSize: '35 KB', generating: false }
])

watch(activeTab, (val) => {
  if (val !== 'logs') {
    queryForm.page = 1
  }
})

const logList = ref<any[]>([])
const logTotal = ref(0)
const logLoading = ref(false)
const logQueryForm = reactive({
  page: 1,
  pageSize: 10
})

const statusIconMap: Record<string, any> = {
  success: Check,
  warning: Warning,
  danger: CircleClose,
  primary: DocumentChecked
}

const statusTextMap: Record<string, string> = {
  success: '成功',
  warning: '警告',
  danger: '失败',
  primary: '信息'
}

const fetchLogList = async () => {
  logLoading.value = true
  try {
    const res: any = await getStaticLogList({
      page: logQueryForm.page,
      pageSize: logQueryForm.pageSize
    })
    if (res.code === 0 || res.code === 200) {
      logList.value = (res.data.list || []).map((item: any) => ({
        ...item,
        statusText: statusTextMap[item.status] || item.status,
        icon: statusIconMap[item.status] || DocumentChecked,
        time: item.createTime
      }))
      logTotal.value = res.data.total || 0
    }
  } catch (error) {
    console.error(error)
  } finally {
    logLoading.value = false
  }
}

const handleLogSizeChange = (val: number) => {
  logQueryForm.pageSize = val
  logQueryForm.page = 1
  fetchLogList()
}

const handleLogCurrentChange = (val: number) => {
  logQueryForm.page = val
  fetchLogList()
}

const fetchPageList = async () => {
  loading.value = true
  try {
    if (activeTab.value === 'home') {
      const res: any = await getPages({ pageType: 'home' })
      homeList.value = res.data || []
    } else if (activeTab.value === 'column') {
      const res: any = await getPages({ pageType: 'column' })
      columnList.value = res.data || []
    } else if (activeTab.value === 'detail') {
      const res: any = await getArticleColumnPublishes({
        page: queryForm.page,
        pageSize: queryForm.pageSize
      })
      detailList.value = res.data?.list || []
      detailTotal.value = res.data?.total || 0
    } else if (activeTab.value === 'topic') {
      const res: any = await getPages({ pageType: 'special' })
      topicList.value = res.data || []
    } else {
      const res: any = await getArticleColumnPublishes({
        page: queryForm.page,
        pageSize: queryForm.pageSize
      })
      if (res.data) {
        pageList.value = res.data.list || []
      }
    }
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const handleTabChange = () => {
  if (activeTab.value === 'logs') {
    logQueryForm.page = 1
    fetchLogList()
  } else {
    queryForm.page = 1
    fetchPageList()
  }
}

const handleSizeChange = (val: number) => {
  queryForm.pageSize = val
  queryForm.page = 1
  if (activeTab.value === 'detail') {
    fetchPageList()
  }
}

const handleCurrentChange = (val: number) => {
  queryForm.page = val
  if (activeTab.value === 'detail') {
    fetchPageList()
  }
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

const handleGenerateSingle = async (row: any) => {
  row.generating = true
  const name = row.title || row.name
  try {
    // TODO: 调用单页生成接口
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success(`「${name}」生成成功`)
  } catch (error) {
    ElMessage.error(`「${name}」生成失败`)
  } finally {
    row.generating = false
  }
}

const handlePreview = (row: any) => {
  window.open(row.path, '_blank')
}

const addLog = (status: string, icon: any, operation: string, message: string, pageName = '-', path = '-', duration = '-', fileSize = '-', operator = 'admin') => {
  const statusTextMap: Record<string, string> = { success: '成功', warning: '警告', danger: '失败', primary: '信息' }
  logList.value = [
    { id: Date.now(), operation, status, statusText: statusTextMap[status] || status, icon, time: new Date().toLocaleString(), pageName, path, duration, fileSize, operator, message },
    ...logList.value
  ]
}

const clearLogs = () => {
  ElMessageBox.confirm('确定要清空所有静态化日志吗？', '确认清空', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res: any = await clearStaticLogs()
      if (res.code === 0 || res.code === 200) {
        logList.value = []
        logTotal.value = 0
        ElMessage.success('日志已清空')
      } else {
        ElMessage.error(res.msg || '清空失败')
      }
    } catch (error) {
      ElMessage.error('清空日志失败')
    }
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

  .tab-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .tab-header-actions {
      display: flex;
      justify-content: flex-end;
      margin-bottom: 12px;
    }

    .pagination {
      margin-top: 20px;
      display: flex;
      justify-content: flex-end;
    }

    .log-scroll-container {
      max-height: 480px;
      overflow-y: auto;
      padding-right: 8px;

      &::-webkit-scrollbar {
        width: 6px;
      }

      &::-webkit-scrollbar-thumb {
        background: #dcdfe6;
        border-radius: 3px;
      }

      &::-webkit-scrollbar-track {
        background: transparent;
      }
    }

    .log-loading-more,
    .log-no-more {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 6px;
      padding: 12px 0;
      font-size: 13px;
      color: #909399;
    }

    .log-content {
      display: flex;
      flex-direction: column;
      gap: 6px;

      .log-header {
        display: flex;
        align-items: center;
        gap: 8px;

        .log-title {
          font-weight: 600;
          color: #2c3e50;
        }
      }

      .log-meta {
        display: flex;
        flex-wrap: wrap;
        gap: 10px;

        .log-meta-item {
          display: inline-flex;
          align-items: center;
          gap: 4px;
          font-size: 12px;
          color: #606266;
          background: #f5f7fa;
          padding: 2px 8px;
          border-radius: 4px;
        }
      }

      .log-detail {
        font-size: 13px;
        color: #909399;
      }
    }
  }
}
</style>
