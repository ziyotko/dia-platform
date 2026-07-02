<template>
  <div class="dashboard">
    <el-row :gutter="20" class="stat-cards">
      <el-col :xs="24" :sm="12" :md="4">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(64, 158, 255, 0.1); color: #409eff;">
              <el-icon size="28"><Picture /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.adCount.toLocaleString() }}</div>
              <div class="stat-label">广告总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="4">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(103, 194, 58, 0.1); color: #67c23a;">
              <el-icon size="28"><DocumentChecked /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.articleCount.toLocaleString() }}</div>
              <div class="stat-label">文章总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="4">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(230, 162, 60, 0.1); color: #e6a23c;">
              <el-icon size="28"><View /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.todayVisit.toLocaleString() }}</div>
              <div class="stat-label">今日访问量</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="4">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(245, 108, 108, 0.1); color: #f56c6c;">
              <el-icon size="28"><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.todayStaticCount.toLocaleString() }}</div>
              <div class="stat-label">我的草稿文章</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="4">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(168, 85, 247, 0.1); color: #a855f7;">
              <el-icon size="28"><Clock /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.todayAuditCount.toLocaleString() }}</div>
              <div class="stat-label">我的待审核文章</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="4">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(99, 102, 241, 0.1); color: #6366f1;">
              <el-icon size="28"><EditPen /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.myArticleCount.toLocaleString() }}</div>
              <div class="stat-label">我的已发布文章</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="dashboard-main">
            <el-col :xs="24" :lg="8">
        <el-card class="notice-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>待处理</span>
              <el-link type="primary" underline="never" @click="$router.push('/pending-audits')">更多</el-link>
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
      <el-col :xs="24" :lg="16">
        <el-card class="chart-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>访问趋势</span>
              <el-radio-group v-model="chartPeriod" size="small">
                <el-radio-button value="week">本周</el-radio-button>
                <el-radio-button value="month">本月</el-radio-button>
                <el-radio-button value="year">全年</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div ref="chartRef" class="chart-container"></div>
        </el-card>
      </el-col>

    </el-row>

    <el-row :gutter="20" class="dashboard-bottom">
      <el-col :xs="24" :lg="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>快捷入口</span>
            </div>
          </template>
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
      <el-col :xs="24" :lg="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>登录日志</span>
              <el-link type="primary" underline="never" @click="$router.push('/login-logs')">更多</el-link>
            </div>
          </template>
          <el-table :data="loginLogs" size="small" :show-header="false">
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
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, watch, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'
import type { ECharts } from 'echarts'
import {
  Picture,
  DocumentChecked,
  View,
  Document,
  Clock,
  EditPen
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getDashboardStats, getLoginLogs, getVisitTrend, getMyAuditArticles } from '@/api/dashboard'

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
})

const chartPeriod = ref('week')
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
      itemStyle: { color: '#409eff' },
      lineStyle: { width: 3, color: '#409eff' },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
          { offset: 1, color: 'rgba(64, 158, 255, 0.05)' }
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
    const res: any = await getVisitTrend(chartPeriod.value)
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

watch(visitData, () => {
  updateChart()
}, { deep: true })

onUnmounted(() => {
  window.removeEventListener('resize', () => chartInstance?.resize())
  chartInstance?.dispose()
  chartInstance = null
})



const quickLinks = [
  { name: '用户管理', icon: 'User', path: '/system/users', bg: 'rgba(64, 158, 255, 0.1)', color: '#409eff' },
  { name: '内容发布', icon: 'EditPen', path: '/content/article', bg: 'rgba(103, 194, 58, 0.1)', color: '#67c23a' },
  { name: '广告管理', icon: 'Promotion', path: '/content/ad', bg: 'rgba(230, 162, 60, 0.1)', color: '#e6a23c' },
  { name: '静态化管理', icon: 'DocumentChecked', path: '/content/static', bg: 'rgba(168, 85, 247, 0.1)', color: '#a855f7' },
  { name: '系统设置', icon: 'Setting', path: '/settings', bg: 'rgba(245, 108, 108, 0.1)', color: '#f56c6c' }
]


</script>

<style scoped lang="scss">
.dashboard {
  .stat-cards {
    margin-bottom: 20px;
  }

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

  .dashboard-main {
    margin-bottom: 20px;
  }

  .chart-card,
  .notice-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-weight: 600;
    color: #2c3e50;
  }

  .chart-container {
    width: 100%;
    height: 358px;
  }

  .notice-list {
    .notice-empty {
      text-align: center;
      padding: 24px 0;
      color: #c0c4cc;
      font-size: 14px;
    }

    .notice-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 12px 0;
      border-bottom: 1px solid #f0f7ff;
      cursor: pointer;
      transition: background 0.2s;

      &:hover {
        background: #f5f9ff;
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
      border-radius: 12px;
      border: 1px solid #e6f2ff;
    }
  }

  .quick-links {
    display: grid;
    height: 160px;
    grid-template-columns: repeat(5, 1fr);
    gap: 16px;

    .quick-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 10px;
      padding: 20px;
      border-radius: 12px;
      cursor: pointer;
      transition: all 0.3s;

      &:hover {
        background: #f5f9ff;
        transform: translateY(-2px);
      }

      .quick-icon {
        width: 48px;
        height: 48px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
      }

      .quick-name {
        font-size: 13px;
        color: #606266;
      }
    }
  }
}
</style>