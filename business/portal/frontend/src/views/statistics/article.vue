<template>
  <div class="page-container">
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="24" :sm="12" :md="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(64, 158, 255, 0.1); color: #409eff;">
              <el-icon size="28"><User /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ authorCount }}</div>
              <div class="stat-label">发文作者数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(103, 194, 58, 0.1); color: #67c23a;">
              <el-icon size="28"><DocumentChecked /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ totalCount }}</div>
              <div class="stat-label">已发布文章数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: rgba(230, 162, 60, 0.1); color: #e6a23c;">
              <el-icon size="28"><Trophy /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ topAuthor || '-' }}</div>
              <div class="stat-label">最高产作者</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover" class="chart-card">
      <template #header>
        <div class="card-header">
          <span>文章作者统计</span>
          <div class="header-extra">
            <span class="total-info">仅统计已发布的文章</span>
            <el-radio-group v-model="period" size="small">
              <el-radio-button value="week">本周</el-radio-button>
              <el-radio-button value="month">本月</el-radio-button>
              <el-radio-button value="year">全年</el-radio-button>
            </el-radio-group>
          </div>
        </div>
      </template>
      <div v-loading="loading" class="chart-wrapper">
        <div ref="chartRef" class="chart-container"></div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed, nextTick } from 'vue'
import * as echarts from 'echarts'
import type { ECharts } from 'echarts'
import { User, DocumentChecked, Trophy } from '@element-plus/icons-vue'
import { getArticleAuthorStats } from '@/api/article'

const loading = ref(false)
const chartRef = ref<HTMLDivElement | null>(null)
let chartInstance: ECharts | null = null

const period = ref('week')
const chartData = ref<any[]>([])

const authorCount = computed(() => chartData.value.length)
const totalCount = computed(() => chartData.value.reduce((sum, item) => sum + (item.count || 0), 0))
const topAuthor = computed(() => {
  if (chartData.value.length === 0) return '-'
  return chartData.value[0].author || '未知'
})

const periodText = computed(() => {
  const map: Record<string, string> = { week: '本周', month: '本月', year: '全年' }
  return map[period.value] || ''
})

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getArticleAuthorStats(period.value)
    if (res && res.data) {
      chartData.value = res.data.list || []
      nextTick(() => updateChart())
    }
  } catch (error) {
    // request interceptor 已处理错误提示
  } finally {
    loading.value = false
  }
}

const updateChart = () => {
  if (!chartInstance) return

  const authors = chartData.value.map(item => item.author || '未知')
  const values = chartData.value.map(item => item.count || 0)

  chartInstance.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: any) => {
        const p = Array.isArray(params) ? params[0] : params
        return `<div style="font-weight:600;margin-bottom:4px">${p.name}</div>
                <div>${periodText.value}已发布：${p.value} 篇</div>`
      },
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#e6f2ff',
      borderWidth: 1,
      textStyle: { color: '#2c3e50' },
      extraCssText: 'box-shadow: 0 4px 12px rgba(0,0,0,0.1); border-radius: 8px;'
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: '12%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: authors,
      axisLine: { lineStyle: { color: '#e4e7ed' } },
      axisLabel: {
        color: '#606266',
        interval: 0,
        rotate: authors.length > 12 ? 30 : 0
      },
      axisTick: { alignWithLabel: true }
    },
    yAxis: {
      type: 'value',
      name: '文章数',
      minInterval: 1,
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: { lineStyle: { color: '#f2f6fc' } },
      axisLabel: { color: '#606266' }
    },
    series: [
      {
        type: 'bar',
        data: values,
        barWidth: '60%',
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#409eff' },
            { offset: 1, color: '#79bbff' }
          ]),
          borderRadius: [6, 6, 0, 0]
        },
        emphasis: {
          itemStyle: {
            color: '#66b1ff'
          }
        },
        label: {
          show: true,
          position: 'top',
          color: '#409eff',
          fontWeight: 600
        }
      }
    ]
  }, true)
}

const initChart = () => {
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)
  updateChart()
  window.addEventListener('resize', () => chartInstance?.resize())
}

watch(period, () => {
  fetchData()
})

onMounted(() => {
  fetchData()
  initChart()
})

onUnmounted(() => {
  window.removeEventListener('resize', () => chartInstance?.resize())
  chartInstance?.dispose()
  chartInstance = null
})
</script>

<style scoped lang="scss">
.page-container {
  .stat-row {
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
    font-size: 22px;
    font-weight: 700;
    color: #2c3e50;
    line-height: 1.2;
  }

  .stat-label {
    font-size: 13px;
    color: #909399;
    margin-top: 4px;
  }

  .chart-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    :deep(.el-card__header) {
      padding: 16px 20px;
    }

    :deep(.el-card__body) {
      padding: 24px;
    }
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-weight: 600;
    color: #2c3e50;
    font-size: 16px;

    .header-extra {
      display: flex;
      align-items: center;
      gap: 16px;

      .total-info {
        font-size: 13px;
        color: #909399;
        font-weight: normal;
      }
    }
  }

  .chart-wrapper {
    min-height: 500px;
  }

  .chart-container {
    width: 100%;
    height: 480px;
  }
}

@media (max-width: 768px) {
  .page-container {
    .card-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;

      .header-extra {
        width: 100%;
        justify-content: space-between;
      }
    }

    .chart-container {
      height: 360px;
    }
  }
}
</style>
