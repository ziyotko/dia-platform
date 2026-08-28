<template>
  <div class="page-container">
    <!-- 静态化服务监控 -->
    <el-card shadow="hover" class="monitor-card">
      <template #header>
        <div class="card-header">
          <span>
            <el-icon><Monitor /></el-icon>
            静态化服务监控
          </span>
          <div class="monitor-actions">
            <el-tag :type="monitorData.online ? 'success' : 'danger'" effect="dark" class="status-tag">
              <el-icon v-if="monitorData.online"><CircleCheck /></el-icon>
              <el-icon v-else><CircleClose /></el-icon>
              {{ monitorData.online ? '运行中' : '已停止' }}
            </el-tag>
            <span class="last-check">最后更新时间：{{ monitorData.lastCheckTime }}</span>
            <el-button type="primary" :loading="monitorLoading" @click="fetchStaticMonitor">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>
    </el-card>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="24" :sm="12" :md="8" :lg="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon size="28"><DocumentChecked /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statData.todayCount }} 个文件</div>
              <div class="stat-label">今日静态化页面</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8" :lg="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon size="28"><Clock /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statData.lastTime }}</div>
              <div class="stat-label">全站最后静态化时间</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8" :lg="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon size="28"><HomeFilled /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statData.homeLastTime }}</div>
              <div class="stat-label">首页最后静态化时间</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8" :lg="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon size="28"><Menu /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statData.columnLastTime }}</div>
              <div class="stat-label">栏目页最后静态化时间</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8" :lg="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(155, 89, 182, 0.1); color: #9b59b6;">
              <el-icon size="28"><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statData.detailLastTime }}</div>
              <div class="stat-label">详情页最后静态化时间</div>
            </div>
          </div>
        </el-card>
      </el-col>
            <el-col :xs="24" :sm="12" :md="8" :lg="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon size="28"><Collection /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statData.topicLastTime }}</div>
              <div class="stat-label">专题页最后静态化时间</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 批量操作 / 单页操作 -->
    <el-card shadow="hover" class="tab-card">
      <template #header>
        <div class="card-header">
          <span>静态化管理</span>
          <div class="header-right">
            <el-button type="info" plain @click="openLogDialog">
              <el-icon><Document /></el-icon>
              静态化日志
            </el-button>
          </div>
        </div>
      </template>
      <el-tabs v-model="mainTab" @tab-change="handleMainTabChange">
        <!-- 批量操作 -->
        <el-tab-pane label="批量操作" name="batch">
          <div class="batch-actions">
            <div class="action-list">
              <el-button
                type="primary"
                size="large"
                :disabled="!monitorData.online"
                :loading="submitting === 'site'"
                @click="handleGenerateAll"
              >
                <el-icon><Refresh /></el-icon>
                生成全站
              </el-button>
              <el-button
                type="success"
                size="large"
                :disabled="!monitorData.online"
                :loading="submitting === 'pages'"
                @click="handleGenerateHome"
              >
                <el-icon><HomeFilled /></el-icon>
                生成首页
              </el-button>
              <el-button
                type="warning"
                size="large"
                :disabled="!monitorData.online"
                :loading="submitting === 'lists'"
                @click="handleGenerateColumn"
              >
                <el-icon><Menu /></el-icon>
                生成栏目页
              </el-button>
              <el-button
                type="danger"
                size="large"
                :disabled="!monitorData.online"
                :loading="submitting === 'articles'"
                @click="handleGenerateDetail"
              >
                <el-icon><Document /></el-icon>
                生成详情页
              </el-button>
                            <el-button type="info" size="large" :disabled="!monitorData.online" @click="handleGenerateTopic">
                <el-icon><Collection /></el-icon>
                生成专题页
              </el-button>
            </div>
            <div class="action-info">
              <el-tag v-if="hasActiveJob" type="warning" effect="dark">
                <el-icon class="is-loading"><Loading /></el-icon>
                有任务正在执行中...
              </el-tag>
              <span class="last-check">任务状态自动轮询（每 3 秒），成功/失败/中断将自动通知</span>
              <el-button type="primary" :loading="jobLoading" @click="refreshAllJobs(true)">
                <el-icon><Refresh /></el-icon>
                刷新任务
              </el-button>
            </div>
          </div>

          <!-- 静态化任务 -->
          <div class="job-section">
            <div class="job-header">
              <span>
                <el-icon><Monitor /></el-icon>
                静态化任务
              </span>
            </div>
            <el-table :data="jobList" v-loading="jobLoading" border stripe empty-text="暂无静态化任务，点击上方按钮发起批量操作">
              <el-table-column label="任务ID" min-width="150" show-overflow-tooltip>
                <template #default="{ row }">
                  <span class="job-id">{{ row.id }}</span>
                </template>
              </el-table-column>
              <el-table-column label="类型" width="120" align="center">
                <template #default="{ row }">
                  <el-tag :type="row.kindType" effect="plain">{{ row.kindText }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="120" align="center">
                <template #default="{ row }">
                  <el-tag :type="row.statusType" effect="dark">
                    <el-icon v-if="row.status === 'running'" class="is-loading"><Loading /></el-icon>
                    {{ row.statusText }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="进度" min-width="280">
                <template #default="{ row }">
                  <div class="job-progress">
                    <div class="progress-stage">
                      {{ row.progress?.stage || (row.status === 'succeeded' ? '已完成' : '等待执行') }}
                    </div>
                    <el-progress
                      :percentage="row.progressPercent"
                      :status="row.progressStatus"
                      :indeterminate="row.progressIndeterminate"
                      :stroke-width="10"
                    />
                    <div class="progress-meta">
                      <span>已处理 {{ row.progress?.processed ?? 0 }} / {{ row.progress?.total ?? 0 }}</span>
                      <span>已生成文件 {{ row.progress?.generated_files ?? 0 }}</span>
                    </div>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="创建时间" width="170">
                <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
              </el-table-column>
              <el-table-column label="更新时间" width="170">
                <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
              </el-table-column>
              <el-table-column label="操作" width="90" align="center" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" :loading="row.refreshing" @click="refreshJob(row.id)">
                    <el-icon><Refresh /></el-icon>刷新
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>

        <!-- 单页操作 -->
        <el-tab-pane label="单页操作" name="single">
          <el-tabs v-model="activeTab" @tab-change="handleTabChange">
            <!-- 首页 -->
            <el-tab-pane label="首页" name="home">
              <el-table :data="homePagedList" v-loading="loading" border stripe>
                <el-table-column type="index" width="60" align="center" />
                <el-table-column prop="id" label="ID" width="80" align="center" />
                <el-table-column prop="name" label="名称" min-width="160" />
                <el-table-column prop="template" label="模板名称" min-width="60" show-overflow-tooltip />
                <el-table-column prop="createTime" label="创建时间" width="170" />
                <el-table-column prop="updatedAt" label="更新时间" width="170" />
                <el-table-column label="操作" width="280" align="center" fixed="right">
                  <template #default="{ row }">
                    <el-button link type="primary" :loading="row.generating" :disabled="!monitorData.online" @click="handleGenerateSingle(row)">
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
                <el-table-column prop="name" label="栏目名称" min-width="160" />
                <el-table-column prop="template" label="模板名称" min-width="120" show-overflow-tooltip />
                <el-table-column prop="createTime" label="创建时间" width="170" />
                <el-table-column prop="updatedAt" label="更新时间" width="170" />
                <el-table-column label="操作" width="280" align="center" fixed="right">
                  <template #default="{ row }">
                    <el-button link type="primary" :loading="row.generating" :disabled="!monitorData.online" @click="handleGenerateSingle(row)">
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

            <!-- 详情页 -->
            <el-tab-pane label="详情页" name="detail">
              <el-table :data="detailPagedList" v-loading="loading" border stripe>
                <el-table-column type="index" width="60" align="center" />
                <el-table-column prop="id" label="ID" width="80" align="center" />
                <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
                <el-table-column prop="name" label="模板名称" min-width="60" show-overflow-tooltip />
                <el-table-column prop="author" label="作者" min-width="60" />
                <el-table-column prop="source" label="来源" min-width="100" />
                <el-table-column prop="createTime" label="创建时间" width="170" />
                <el-table-column prop="updatedAt" label="更新时间" width="170" />
                <el-table-column label="操作" width="340" align="center" fixed="right">
                  <template #default="{ row }">
                    <el-button link type="primary" :loading="row.generating" :disabled="!monitorData.online" @click="handleGenerateSingle(row)">
                      <el-icon><Refresh /></el-icon>重新生成
                    </el-button>
                    <el-button link type="success" @click="handlePreview(row)">
                      <el-icon><View /></el-icon>预览
                    </el-button>
                    <el-button link type="danger" :loading="row.deleting" :disabled="!monitorData.online" @click="handleDeleteArticle(row)">
                      <el-icon><Delete /></el-icon>删除
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

            <!-- 专题页 -->
            <el-tab-pane label="专题页" name="topic">
              <el-table :data="topicPagedList" v-loading="loading" border stripe>
                <el-table-column type="index" width="60" align="center" />
                <el-table-column prop="id" label="ID" width="80" align="center" />
                <el-table-column prop="name" label="名称" min-width="160" />
                <el-table-column prop="template" label="模板名称" min-width="60" show-overflow-tooltip />
                <el-table-column prop="createTime" label="创建时间" width="170" />
                <el-table-column prop="updatedAt" label="更新时间" width="170" />
                <el-table-column label="操作" width="280" align="center" fixed="right">
                  <template #default="{ row }">
                    <el-button link type="primary" :loading="row.generating" :disabled="!monitorData.online" @click="handleGenerateSingle(row)">
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

            
          </el-tabs>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 静态化日志弹窗 -->
    <el-dialog v-model="logDialogVisible" title="静态化日志" width="960px">
      <div class="log-dialog-body">
        <div class="log-toolbar">
          <el-button type="danger" link @click="clearLogs">
            <el-icon><Delete /></el-icon>清空日志
          </el-button>
        </div>
        <div class="log-scroll-container" v-loading="logLoading">
          <el-timeline v-if="logList.length">
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
          <el-empty v-else description="暂无静态化日志" :image-size="80" />
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
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed, watch, markRaw } from 'vue'
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
  CircleCheck,
  Warning,
  User,
  Monitor
} from '@element-plus/icons-vue'
import { getArticleColumnPublishes } from '@/api/article'
import { getColumnPublishes } from '@/api/column'
import { getStaticPages } from '@/api/static_page'
import { getStaticLogList, clearStaticLogs, getStaticLatestTimes } from '@/api/static_log'
import { getStaticMonitor } from '@/api/static_monitor'
import { startStaticJob, getStaticJob, startStaticPage, startStaticList, startStaticArticle, deleteStaticArticle } from '@/api/static_job'
import type { StaticJob, StaticJobStatus } from '@/api/static_job'
import { getSettings } from '@/api/settings'
import { getErrorMessage } from '@/utils/request'

const loading = ref(false)
const activeTab = ref('home')
const mainTab = ref('batch') // 外层 Tab：批量操作 / 单页操作

// 批量操作参数
const grayEnabled = ref(true) // 首页整体变灰：1 开启 / 2 关闭
const staticPath = ref('') // 静态化输出路径（后端全局变量）
const submitting = ref<string | null>(null) // 当前提交中的任务类型

// 静态化任务列表与轮询
const jobLoading = ref(false)
const jobList = ref<any[]>([])
const activeJobIds = new Set<string>()
const POLL_INTERVAL = 3000
let pollTimer: any = null

const kindMap: Record<string, { text: string; type: string }> = {
  site: { text: '生成全站', type: 'primary' },
  pages: { text: '生成首页', type: 'success' },
  lists: { text: '生成栏目页', type: 'warning' },
  articles: { text: '生成详情页', type: 'danger' }
}

const jobStatusMap: Record<StaticJobStatus, { text: string; type: string }> = {
  queued: { text: '等待执行', type: 'info' },
  running: { text: '正在执行', type: 'warning' },
  succeeded: { text: '执行成功', type: 'success' },
  failed: { text: '执行失败', type: 'danger' },
  interrupted: { text: '已中断', type: 'danger' }
}

const hasActiveJob = computed(() =>
  jobList.value.some(j => j.status === 'queued' || j.status === 'running')
)

// 格式化 ISO 时间字符串：去掉中间的 T，仅保留到秒（例如 2026-08-24 12:34:56）
const formatTime = (t?: string) =>
  t ? t.replace('T', ' ').slice(0, 19) : '-'

// 格式化时间字符串：仅保留到分钟，不显示秒（例如 2026-08-24 12:34）
const formatShortTime = (t?: string) =>
  t ? t.replace('T', ' ').slice(0, 16) : '-'

// 将后端任务结构标准化为前端展示结构
const normalizeJob = (job: StaticJob) => {
  const kindInfo = kindMap[job.kind] || { text: job.kind, type: 'info' }
  const statusInfo = jobStatusMap[job.status] || { text: job.status, type: 'info' }
  const progress: any = job.progress || {}
  const total = Number(progress.total) || 0
  const processed = Number(progress.processed) || 0
  const isActive = job.status === 'queued' || job.status === 'running'
  const percent = job.status === 'succeeded'
    ? 100
    : total > 0
      ? Math.min(99, Math.round((processed / total) * 100))
      : isActive ? 0 : 100
  return {
    ...job,
    kindText: kindInfo.text,
    kindType: kindInfo.type,
    statusText: statusInfo.text,
    statusType: statusInfo.type,
    progressPercent: percent,
    progressStatus: job.status === 'failed' || job.status === 'interrupted'
      ? 'exception'
      : job.status === 'succeeded' ? 'success' : undefined,
    progressIndeterminate: isActive && total === 0,
    refreshing: false
  }
}

// 读取静态化参数配置（输出路径 / 首页整体变灰）
const loadStaticSettings = async () => {
  try {
    const res: any = await getSettings()
    if (res.code === 0 || res.code === 200) {
      const s = res.data || {}
      staticPath.value = s.staticPath || ''
      grayEnabled.value = !!s.homeGray
    }
  } catch (error) {
    // 忽略，不影响页面其他功能
  }
}

// 发起批量操作任务（后端透传静态化程序 202 + 任务信息）
const runStaticJob = async (kind: 'site' | 'pages' | 'lists' | 'articles', title: string) => {
  submitting.value = kind
  try {
    const res = await startStaticJob(kind, grayEnabled.value ? 1 : 2)
    if (res.status === 202 && res.data?.ok && res.data.job) {
      const job = normalizeJob(res.data.job)
      const idx = jobList.value.findIndex(j => j.id === job.id)
      if (idx >= 0) {
        jobList.value[idx] = job
      } else {
        jobList.value.unshift(job)
      }
      if (job.status === 'queued' || job.status === 'running') {
        activeJobIds.add(job.id)
        startPolling()
      }
      ElMessage.success(`${title}任务已提交，请等待处理结果`)
    } else {
      const data: any = res.data || {}
      const msg = data.message || data.msg || '任务提交失败'
      ElMessage.error(`${title}失败：${msg}`)
    }
  } catch (error: any) {
    const msg = getErrorMessage(error)
    ElMessage.error(`${title}失败：${msg}`)
  } finally {
    submitting.value = null
  }
}

// 通用确认：每个静态化操作先弹出确认框，确认后再发起任务
const confirmRun = (kind: 'site' | 'pages' | 'lists' | 'articles', title: string, message?: string) => {
  ElMessageBox.confirm(
    message || `确定要执行${title}吗？`,
    '确认操作',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    runStaticJob(kind, title)
  }).catch(() => {})
}

const handleGenerateAll = () => confirmRun('site', '生成全站', '确定要执行生成全站吗？此操作可能需要较长时间。')
const handleGenerateHome = () => confirmRun('pages', '生成首页')
const handleGenerateColumn = () => confirmRun('lists', '生成栏目页')
const handleGenerateDetail = () => confirmRun('articles', '生成详情页')
// 专题页生成功能暂未实现，点击仅提示
const handleGenerateTopic = () => {
  ElMessage.info('专题页生成功能暂未实现，敬请期待')
}

// 任务轮询：有活动任务时定时查询状态
const startPolling = () => {
  if (pollTimer) return
  pollTimer = setInterval(() => {
    refreshAllJobs()
  }, POLL_INTERVAL)
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

// 刷新所有活动任务（自动轮询不显示加载遮罩，仅手动刷新时显示）
const refreshAllJobs = async (showLoading = false) => {
  const ids = Array.from(activeJobIds)
  if (ids.length === 0) {
    stopPolling()
    return
  }
  if (showLoading) jobLoading.value = true
  try {
    await Promise.all(ids.map(id => refreshJob(id)))
  } finally {
    if (showLoading) jobLoading.value = false
  }
}

// 刷新单个任务状态
const refreshJob = async (id: string) => {
  const row = jobList.value.find(j => j.id === id)
  if (row) row.refreshing = true
  try {
    const res = await getStaticJob(id)
    // 任务状态查询接口返回 200（仅发起任务时返回 202），此处兼容 2xx 成功码
    if ((res.status === 200 || res.status === 202) && res.data?.ok && res.data.job) {
      const prevStatus = row?.status
      const updated = normalizeJob(res.data.job)
      const idx = jobList.value.findIndex(j => j.id === id)
      if (idx >= 0) jobList.value[idx] = updated

      if (prevStatus && prevStatus !== updated.status
        && ['succeeded', 'failed', 'interrupted'].includes(updated.status)) {
        notifyJobDone(updated)
      }

      if (updated.status === 'queued' || updated.status === 'running') {
        activeJobIds.add(id)
        startPolling()
      } else {
        activeJobIds.delete(id)
        if (activeJobIds.size === 0) stopPolling()
      }
    }
  } catch (error) {
    // 单个任务刷新失败忽略，等待下次轮询
  } finally {
    const r = jobList.value.find(j => j.id === id)
    if (r) r.refreshing = false
  }
}

// 任务结束通知
const notifyJobDone = (job: any) => {
  if (job.status === 'succeeded') {
    ElMessage.success(`「${job.kindText}」任务执行成功`)
    fetchPageList()
    fetchStaticStat()
  } else if (job.status === 'failed') {
    ElMessage.error(`「${job.kindText}」任务执行失败`)
  } else if (job.status === 'interrupted') {
    ElMessage.warning(`「${job.kindText}」任务已中断`)
  }
}

const monitorLoading = ref(false)
const monitorData = reactive({
  online: false,
  address: '',
  httpStatus: 0,
  lastCheckTime: '-',
  message: '尚未获取'
})

const fetchStaticMonitor = async () => {
  monitorLoading.value = true
  try {
    const res: any = await getStaticMonitor()
    if (res.code === 0 || res.code === 200) {
      monitorData.online = !!res.data?.online
      monitorData.address = res.data?.address || ''
      monitorData.httpStatus = res.data?.httpStatus || 0
      monitorData.lastCheckTime = res.data?.lastCheckTime || new Date().toLocaleString()
      monitorData.message = res.data?.message || ''
    } else {
      ElMessage.error(res.msg || '获取静态化服务状态失败')
    }
  } catch (error) {
    ElMessage.error('获取静态化服务状态失败')
  } finally {
    monitorLoading.value = false
  }
}

const statData = reactive({
  todayCount: 0,
  lastTime: '-',
  homeLastTime: '-',
  columnLastTime: '-',
  topicLastTime: '-',
  detailLastTime: '-'
})

// 读取静态化日志最新成功时间及今日成功文件数，填充统计卡片
const fetchStaticStat = async () => {
  try {
    const res: any = await getStaticLatestTimes()
    if (res.code === 0 || res.code === 200) {
      const d = res.data || {}
      statData.todayCount = d.todayFiles || 0
      statData.lastTime = formatShortTime(d.site)
      statData.homeLastTime = formatShortTime(d.home)
      statData.columnLastTime = formatShortTime(d.column)
      statData.topicLastTime = formatShortTime(d.topic)
      statData.detailLastTime = formatShortTime(d.detail)
    }
  } catch (error) {
    // 忽略，保持默认 '-'
  }
}

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

watch(activeTab, () => {
  queryForm.page = 1
})

const logDialogVisible = ref(false)
const logList = ref<any[]>([])
const logTotal = ref(0)
const logLoading = ref(false)
const logQueryForm = reactive({
  page: 1,
  pageSize: 10
})

// 打开静态化日志弹窗：重置到第一页并加载
const openLogDialog = () => {
  logQueryForm.page = 1
  logDialogVisible.value = true
  fetchLogList()
}

const statusIconMap: Record<string, any> = {
  success: markRaw(Check),
  warning: markRaw(Warning),
  danger: markRaw(CircleClose),
  primary: markRaw(DocumentChecked)
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
        icon: markRaw(statusIconMap[item.status] || DocumentChecked),
        time: item.createTime
      }))
      logTotal.value = res.data.total || 0
    }
  } catch (error) {
    ElMessage.error('获取静态化日志失败')
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
      const res: any = await getStaticPages({ pageType: 'home' })
      homeList.value = res.data || []
    } else if (activeTab.value === 'column') {
      const res: any = await getColumnPublishes()
      columnList.value = res.data?.list || []
    } else if (activeTab.value === 'detail') {
      const res: any = await getArticleColumnPublishes({
        page: queryForm.page,
        pageSize: queryForm.pageSize
      })
      detailList.value = res.data?.list || []
      detailTotal.value = res.data?.total || 0
    } else if (activeTab.value === 'topic') {
      const res: any = await getStaticPages({ pageType: 'special' })
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
    ElMessage.error('获取静态化页面失败')
  } finally {
    loading.value = false
  }
}

const handleTabChange = () => {
  queryForm.page = 1
  fetchPageList()
}

// 外层 Tab 切换：进入「单页操作」时加载页面列表
const handleMainTabChange = () => {
  if (mainTab.value === 'single') {
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

const handleGenerateSingle = async (row: any) => {
  // 静态化服务停止时禁止重新生成
  if (!monitorData.online) {
    ElMessage.warning('静态化服务已停止，无法重新生成页面')
    return
  }
  row.generating = true
  const name = row.title || row.name
  try {
    let res: { status: number; data: any }
    if (activeTab.value === 'home') {
      // 首页重新生成：调用后端代理 /static/page（自动附带验证头）
      // 输出目录与首页整体变灰由后端全局变量决定，前端仅传页面名
      res = await startStaticPage(name)
    } else if (activeTab.value === 'column') {
      // 栏目页重新生成：调用后端代理 /static/list（自动附带验证头）
      // 输出目录由后端全局变量决定，前端仅传栏目名称
      res = await startStaticList(name)
    } else if (activeTab.value === 'detail') {
      // 详情页重新生成：调用后端代理 /static/article（自动附带验证头）
      // 输出目录由后端全局变量决定，前端仅传文章ID
      res = await startStaticArticle(row.id)
    } else if (activeTab.value === 'topic') {
      // 专题页生成功能暂未实现，点击仅提示
      ElMessage.info('专题页生成功能暂未实现，敬请期待')
      return
    } else {
      ElMessage.info('该类型暂未接入单页生成接口')
      return
    }
    const data: any = res.data || {}
    if (res.status === 200 && data.ok && data.result) {
      const r = data.result
      const generatedAt = r.generated_at
        ? String(r.generated_at).replace('T', ' ').slice(0, 19)
        : '-'
      ElMessageBox.alert(
        `生成时间：${generatedAt}`,
        `「${name}」生成成功`,
        { confirmButtonText: '确定', type: 'success' }
      )
      fetchPageList()
      fetchStaticStat()
    } else {
      const msg = data.message || data.msg || '生成失败'
      ElMessage.error(`「${name}」生成失败：${msg}`)
    }
  } catch (error: any) {
    const msg = getErrorMessage(error)
    ElMessage.error(`「${name}」生成失败：${msg}`)
  } finally {
    row.generating = false
  }
}

// 删除详情页静态文件：调用后端代理 DELETE /static/article?id={文章ID}&path={输出目录}
// 输出目录取自后端全局变量「静态化输出路径」，确认后执行删除并刷新列表。
const handleDeleteArticle = (row: any) => {
  // 静态化服务停止时禁止删除静态文件
  if (!monitorData.online) {
    ElMessage.warning('静态化服务已停止，无法删除静态文件')
    return
  }
  const name = row.title || row.name || row.id
  ElMessageBox.confirm(
    `确定要删除文章「${name}」(ID: ${row.id}) 的静态文件吗？删除后需重新生成才能恢复。`,
    '确认删除',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    row.deleting = true
    try {
      const res = await deleteStaticArticle(row.id, staticPath.value)
      const data: any = res.data || {}
      if (res.status === 200 && data.ok) {
        ElMessage.success(`「${name}」静态文件已删除`)
        fetchPageList()
        fetchStaticStat()
      } else {
        const msg = data.message || data.msg || '删除失败'
        ElMessage.error(`「${name}」删除失败：${msg}`)
      }
    } catch (error: any) {
      const msg = getErrorMessage(error)
      ElMessage.error(`「${name}」删除失败：${msg}`)
    } finally {
      row.deleting = false
    }
  }).catch(() => {})
}

const handlePreview = (row: any) => {
  window.open(row.path, '_blank')
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
  fetchStaticMonitor()
  fetchStaticStat()
  loadStaticSettings()
})

onUnmounted(() => {
  stopPolling()
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

  .monitor-card {
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

    .monitor-actions {
      display: flex;
      align-items: center;
      gap: 12px;

      .last-check {
        font-size: 13px;
        font-weight: 400;
        color: #909399;
      }

      .status-tag {
        font-size: 14px;
        padding: 6px 16px;
      }
    }

    :deep(.el-card__body) {
      padding: 0;
    }
  }

  .tab-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .card-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      font-weight: 600;
      color: #2c3e50;

      .header-right {
        display: flex;
        align-items: center;
        gap: 12px;
      }
    }

    .batch-actions {
      display: flex;
      align-items: center;
      justify-content: space-between;
      flex-wrap: wrap;
      gap: 12px;
      padding: 8px 0 4px;

      .action-list {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;
        padding: 12px 0 8px;

        .el-button {
          min-width: 140px;
        }
      }

      .action-info {
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        justify-content: flex-end;
        gap: 12px;
        padding: 12px 0 8px;

        .last-check {
          font-size: 13px;
          color: #909399;
        }
      }
    }

    .job-section {
      margin-top: 8px;

      .job-header {
        display: flex;
        align-items: center;
        gap: 12px;
        font-weight: 600;
        color: #2c3e50;
        padding: 12px 0;
      }

      .job-id {
        font-family: 'Consolas', 'Courier New', monospace;
        font-size: 13px;
        color: #409eff;
      }

      .job-progress {
        display: flex;
        flex-direction: column;
        gap: 4px;
        padding: 4px 0;

        .progress-stage {
          font-size: 13px;
          color: #606266;
        }

        .progress-meta {
          display: flex;
          gap: 16px;
          font-size: 12px;
          color: #909399;
        }
      }
    }

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
  }

  .log-dialog-body {
    .log-toolbar {
      display: flex;
      justify-content: flex-end;
      margin-bottom: 12px;
    }

    .log-scroll-container {
      max-height: 560px;
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

    .pagination {
      margin-top: 20px;
      display: flex;
      justify-content: flex-end;
    }
  }
}
</style>
