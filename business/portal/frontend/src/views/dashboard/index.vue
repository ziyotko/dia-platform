<template>
  <div class="dashboard">
    <el-row :gutter="20" class="stat-cards">
      <el-col :xs="24" :sm="12" :md="6" :lg="4">
        <el-card class="stat-card" shadow="never">
          <div class="stat-body">
            <div class="stat-info">
              <div class="stat-label">今日访问量</div>
              <div class="stat-value">{{ stats.todayVisit.toLocaleString() }}</div>
            </div>
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon :size="26"><View /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6" :lg="4">
        <el-card class="stat-card" shadow="never">
          <div class="stat-body">
            <div class="stat-info">
              <div class="stat-label">文章总数</div>
              <div class="stat-value">{{ stats.articleCount.toLocaleString() }}</div>
            </div>
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon :size="26"><DocumentChecked /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6" :lg="4">
        <el-card class="stat-card" shadow="never">
          <div class="stat-body">
            <div class="stat-info">
              <div class="stat-label">我的草稿文章</div>
              <div class="stat-value">{{ stats.todayStaticCount.toLocaleString() }}</div>
            </div>
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon :size="26"><Document /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6" :lg="4">
        <el-card class="stat-card" shadow="never">
          <div class="stat-body">
            <div class="stat-info">
              <div class="stat-label">我的已发布文章</div>
              <div class="stat-value">{{ stats.myArticleCount.toLocaleString() }}</div>
            </div>
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon :size="26"><EditPen /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="8">
        <el-card shadow="never">
          <div class="quick-links">
            <div v-for="(item, index) in quickLinks" :key="index" class="quick-item" @click="$router.push(item.path)">
              <div class="quick-icon" :style="{ background: item.bg, color: item.color }">
                <el-icon size="24"><component :is="item.icon" /></el-icon>
              </div>
              <div class="quick-name">{{ item.name }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="dashboard-main">
      <el-col :xs="24" :lg="16">
        <el-card class="chart-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>访问趋势</span>
              <div class="header-controls">
                <el-select v-model="visitYear" size="small" class="year-select">
                  <el-option v-for="y in yearOptions" :key="y" :label="`${y}年`" :value="y" />
                </el-select>
                <el-radio-group v-model="chartPeriod" size="small">
                  <el-radio-button value="week" :disabled="!isVisitCurrentYear">本周</el-radio-button>
                  <el-radio-button value="month" :disabled="!isVisitCurrentYear">本月</el-radio-button>
                  <el-radio-button value="year">全年</el-radio-button>
                </el-radio-group>
              </div>
            </div>
          </template>
          <div ref="chartRef" class="chart-container"></div>
        </el-card>
      </el-col>
            <el-col :xs="24" :lg="8">
        <el-card class="notice-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>待处理</span>
              <el-link type="primary" underline="never" @click="$router.push('/content/pending-audits')">更多</el-link>
            </div>
          </template>
          <div class="notice-list">
            <div v-if="pendingAudits.length === 0" class="notice-empty">暂无待处理事项</div>
            <div
              v-for="item in pendingAudits"
              :key="item.id"
              class="notice-item"
              @click="handleAuditClick(item)"
            >
              <el-tag type="warning" size="small">审核</el-tag>
              <span class="notice-title">文章《{{ item.title }}》待审核</span>
              <span class="notice-time">{{ item.createTime }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="dashboard-bottom">
      <el-col :xs="24" :lg="16">
        <el-card class="chart-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>发布趋势</span>
              <div class="header-controls">
                <el-select v-model="articleYear" size="small" class="year-select">
                  <el-option v-for="y in yearOptions" :key="y" :label="`${y}年`" :value="y" />
                </el-select>
                <el-radio-group v-model="articleChartPeriod" size="small">
                  <el-radio-button value="week" :disabled="!isArticleCurrentYear">本周</el-radio-button>
                  <el-radio-button value="month" :disabled="!isArticleCurrentYear">本月</el-radio-button>
                  <el-radio-button value="year">全年</el-radio-button>
                </el-radio-group>
              </div>
            </div>
          </template>
          <div ref="articleChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="8">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>登录日志</span>
              <el-link type="primary" underline="never" @click="$router.push('/system/login-logs')">更多</el-link>
            </div>
          </template>
          <div class="table-container">
          <el-table :data="loginLogs" class="login-log-table" size="small" :show-header="false">
            <el-table-column prop="time" width="150" />
            <el-table-column prop="username" width="90" />
            <el-table-column prop="ip" width="120" />
            <el-table-column>
              <template #default="{ row }">
                <span>{{ row.browser }} / {{ row.os }}</span>
              </template>
            </el-table-column>
            <el-table-column width="60">
              <template #default="{ row }">
                <el-tag :type="row.status === '成功' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive, watch, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'
import type { ECharts } from 'echarts'
import {
  DocumentChecked,
  View,
  Document,
  EditPen
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getDashboardStats, getLoginLogs, getVisitTrend, getArticleTrend, getMyAuditArticles } from '@/api/dashboard'

const router = useRouter()

const stats = reactive({
  adCount: 0,
  articleCount: 0,
  todayVisit: 0,
  todayStaticCount: 0,
  todayAuditCount: 0,
  myArticleCount: 0
})

const fetchStats = async () => {
  try {
    const res: any = await getDashboardStats()
    if (res && res.data) {
      stats.adCount = res.data.adCount || 0
      stats.articleCount = res.data.articleCount || 0
      stats.todayVisit = res.data.todayVisit || 0
      stats.todayStaticCount = res.data.todayStaticCount || 0
      stats.todayAuditCount = res.data.todayAuditCount || 0
      stats.myArticleCount = res.data.myArticleCount || 0
    }
  } catch (error) {
    // 静默失败，保持默认值
  }
}

const loginLogs = ref<any[]>([])

const fetchLoginLogs = async () => {
  try {
    const res: any = await getLoginLogs()
    if (res && res.data) {
      loginLogs.value = res.data
    }
  } catch (error) {
    // 静默失败
  }
}

const pendingAudits = ref<any[]>([])

const fetchPendingAudits = async () => {
  try {
    const res: any = await getMyAuditArticles({ pageSize: 8 })
    if (res && res.data) {
      pendingAudits.value = res.data.list || []
    }
  } catch (error) {
    // 静默失败
  }
}

const handleAuditClick = (item: any) => {
  router.push({
    path: '/content/article',
    query: { auditArticleId: String(item.id) }
  })
}

onMounted(() => {
  fetchStats()
  fetchLoginLogs()
  fetchVisitTrend()
  fetchPendingAudits()
  initChart()
  initArticleChart()
  fetchArticleTrend()
})

const chartPeriod = ref('week')
const currentYear = new Date().getFullYear()
const yearOptions = computed(() => {
  const years: number[] = []
  for (let y = currentYear; y >= currentYear - 4; y--) {
    years.push(y)
  }
  return years
})

const visitYear = ref(currentYear)
const lastVisitPeriod = ref('week')
const isVisitCurrentYear = computed(() => visitYear.value === currentYear)

const visitData = ref<{ label: string; value: number }[]>([
  { label: '周一', value: 45 },
  { label: '周二', value: 62 },
  { label: '周三', value: 55 },
  { label: '周四', value: 78 },
  { label: '周五', value: 68 },
  { label: '周六', value: 85 },
  { label: '周日', value: 72 }
])

const chartRef = ref<HTMLDivElement | null>(null)
let chartInstance: ECharts | null = null

const updateChart = () => {
  if (!chartInstance) return
  const labels = visitData.value.map(item => item.label)
  const values = visitData.value.map(item => item.value)
  chartInstance.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', top: '10%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: labels },
    yAxis: { type: 'value', name: '访问次数', minInterval: 1 },
    series: [{
      type: 'line',
      data: values,
      smooth: true,
      symbol: 'circle',
      symbolSize: 8,
      itemStyle: { color: '#002fa7' },
      lineStyle: { width: 3, color: '#002fa7' },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(0, 47, 167, 0.3)' },
          { offset: 1, color: 'rgba(0, 47, 167, 0.05)' }
        ])
      }
    }]
  }, true)
}

const initChart = () => {
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)
  updateChart()
  window.addEventListener('resize', () => chartInstance?.resize())
}

const fetchVisitTrend = async () => {
  try {
    const res: any = await getVisitTrend(chartPeriod.value, visitYear.value)
    if (res && Array.isArray(res.data) && res.data.length > 0) {
      visitData.value = res.data.map((item: any) => ({
        label: item.label,
        value: Number(item.value) || 0
      }))
      nextTick(() => updateChart())
    }
  } catch (error) {
    ElMessage.error('获取访问趋势失败')
  }
}

watch(chartPeriod, () => {
  fetchVisitTrend()
})

watch(visitYear, (newYear) => {
  if (newYear !== currentYear) {
    // 查看往年：本周/本月不适用，强制全年
    if (chartPeriod.value !== 'year') {
      lastVisitPeriod.value = chartPeriod.value
      chartPeriod.value = 'year'
    } else {
      fetchVisitTrend()
    }
  } else {
    // 切回今年：恢复之前的周期
    if (chartPeriod.value !== lastVisitPeriod.value) {
      chartPeriod.value = lastVisitPeriod.value
    } else {
      fetchVisitTrend()
    }
  }
})

watch(visitData, () => {
  updateChart()
}, { deep: true })

const articleChartPeriod = ref('week')
const articleYear = ref(currentYear)
const lastArticlePeriod = ref('week')
const isArticleCurrentYear = computed(() => articleYear.value === currentYear)

const articleVisitData = ref<{ label: string; value: number }[]>([
  { label: '周一', value: 0 },
  { label: '周二', value: 0 },
  { label: '周三', value: 0 },
  { label: '周四', value: 0 },
  { label: '周五', value: 0 },
  { label: '周六', value: 0 },
  { label: '周日', value: 0 }
])

const articleChartRef = ref<HTMLDivElement | null>(null)
let articleChartInstance: ECharts | null = null

const updateArticleChart = () => {
  if (!articleChartInstance) return
  const labels = articleVisitData.value.map(item => item.label)
  const values = articleVisitData.value.map(item => item.value)
  articleChartInstance.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', top: '10%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: labels },
    yAxis: { type: 'value', name: '文章数量', minInterval: 1 },
    series: [{
      type: 'line',
      data: values,
      smooth: true,
      symbol: 'circle',
      symbolSize: 8,
      itemStyle: { color: '#002fa7' },
      lineStyle: { width: 3, color: '#002fa7' },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(0, 47, 167, 0.3)' },
          { offset: 1, color: 'rgba(0, 47, 167, 0.05)' }
        ])
      }
    }]
  }, true)
}

const initArticleChart = () => {
  if (!articleChartRef.value) return
  articleChartInstance = echarts.init(articleChartRef.value)
  updateArticleChart()
  window.addEventListener('resize', () => articleChartInstance?.resize())
}

const fetchArticleTrend = async () => {
  try {
    const res: any = await getArticleTrend(articleChartPeriod.value, articleYear.value)
    if (res && Array.isArray(res.data) && res.data.length > 0) {
      articleVisitData.value = res.data.map((item: any) => ({
        label: item.label,
        value: Number(item.value) || 0
      }))
      nextTick(() => updateArticleChart())
    }
  } catch (error) {
    ElMessage.error('获取文章发布趋势失败')
  }
}

watch(articleChartPeriod, () => {
  fetchArticleTrend()
})

watch(articleYear, (newYear) => {
  if (newYear !== currentYear) {
    // 查看往年：本周/本月不适用，强制全年
    if (articleChartPeriod.value !== 'year') {
      lastArticlePeriod.value = articleChartPeriod.value
      articleChartPeriod.value = 'year'
    } else {
      fetchArticleTrend()
    }
  } else {
    // 切回今年：恢复之前的周期
    if (articleChartPeriod.value !== lastArticlePeriod.value) {
      articleChartPeriod.value = lastArticlePeriod.value
    } else {
      fetchArticleTrend()
    }
  }
})

watch(articleVisitData, () => {
  updateArticleChart()
}, { deep: true })

onUnmounted(() => {
  window.removeEventListener('resize', () => chartInstance?.resize())
  chartInstance?.dispose()
  chartInstance = null
  articleChartInstance?.dispose()
  articleChartInstance = null
})



const quickLinks = [
  { name: '用户管理', icon: 'User', path: '/system/users', bg: 'var(--el-color-primary-light-9)', color: 'var(--el-color-primary)' },
  { name: '内容发布', icon: 'EditPen', path: '/content/article', bg: 'var(--el-color-primary-light-9)', color: 'var(--el-color-primary)' },
  { name: '广告管理', icon: 'Promotion', path: '/content/ad', bg: 'var(--el-color-primary-light-9)', color: 'var(--el-color-primary)' },
  { name: '静态化管理', icon: 'DocumentChecked', path: '/content/static', bg: 'var(--el-color-primary-light-9)', color: 'var(--el-color-primary)' },
  { name: '系统设置', icon: 'Setting', path: '/settings', bg: 'var(--el-color-primary-light-9)', color: 'var(--el-color-primary)' }
]


</script>

<style scoped lang="scss">
.dashboard {
  .stat-cards {
    margin-bottom: 20px;
  }

  .stat-card {
    transition: transform 0.3s, box-shadow 0.3s;

    &:hover {
      transform: translateY(-3px);
    }

    :deep(.el-card__body) {
      padding: 20px;
    }
  }

  .stat-body {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  .stat-icon {
    flex-shrink: 0;
    width: 52px;
    height: 52px;
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .stat-value {
    font-size: 26px;
    font-weight: 700;
    color: var(--app-text-primary);
    line-height: 1.2;
  }

  .stat-label {
    font-size: 13px;
    color: var(--app-text-secondary);
    margin-bottom: 6px;
  }

  .table-container {
    height: 360px;
    overflow: auto;
  }

  .dashboard-main {
    margin-bottom: 20px;
  }

  .chart-card,
  .notice-card {
    border-radius: var(--app-card-radius);
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    font-weight: 600;
    color: var(--app-text-primary);
  }

  .header-controls {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .year-select {
    width: 100px;
  }

  .chart-container {
    width: 100%;
    height: 358px;
  }

  .notice-list {
     height: 360px;
    .notice-empty {
      text-align: center;
      padding: 24px 0;
      color: #c0c4cc;
      font-size: 14px;
      height: 360px;
      line-height: 330px;
    }

    .notice-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 12px 0;
      border-bottom: 1px solid #f0f2f7;
      cursor: pointer;
      transition: background 0.2s;

      &:hover {
        background: #f5f7fb;
      }

      &:last-child {
        border-bottom: none;
      }

      .notice-title {
        flex: 1;
        font-size: 14px;
        color: #606266;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .notice-time {
        font-size: 12px;
        color: #c0c4cc;
      }
    }
  }

  .dashboard-bottom {
    .el-card {
      border-radius: var(--app-card-radius);
    }

    .login-log-table {
      font-size: 14px;
    }
  }

  .quick-links {
    display: grid;
    height: 55px;
    grid-template-columns: repeat(5, 1fr);

    .quick-item {
      margin-top: -5px;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 5px;
      border-radius: 12px;
      cursor: pointer;
      transition: all 0.3s;

      &:hover {
        background: #f5f7fb;
        transform: translateY(-2px);
      }

      .quick-icon {
        width: 48px;
        height: 48px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content:center;
      }

      .quick-name {
        font-size: 13px;
        color: #606266;
      }
    }
  }
}
</style>