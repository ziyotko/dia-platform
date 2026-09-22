<template>
  <div class="page-container">
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="24" :sm="12" :md="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon size="28"><PriceTag /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.tagCount }}</div>
              <div class="stat-label">标签总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon size="28"><DocumentChecked /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.taggedArticleCount }}</div>
              <div class="stat-label">已打标签文章</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon size="28"><Trophy /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.topTagName || '-' }}</div>
              <div class="stat-label">最热标签</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover" class="chart-card">
      <template #header>
        <div class="card-header">
          <span>标签词云</span>
          <span class="sub-title">字体大小反映文章绑定数量</span>
        </div>
      </template>
      <div v-loading="loading" class="chart-container">
        <div ref="chartRef" class="wordcloud-chart"></div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import * as echarts from 'echarts'
import type { ECharts } from 'echarts'
import 'echarts-wordcloud'
import { PriceTag, DocumentChecked, Trophy } from '@element-plus/icons-vue'
import { getTagArticleStats } from '@/api/tag'
import { BRAND_THEME_COLOR } from '@/stores/app'

const loading = ref(false)
const chartRef = ref<HTMLDivElement | null>(null)
let chartInstance: ECharts | null = null

const chartData = ref<any[]>([])

const stats = computed(() => {
  const tagCount = chartData.value.length
  const taggedArticleCount = chartData.value.reduce((sum, item) => sum + (item.count || 0), 0)
  const topTag = chartData.value.length > 0 ? chartData.value[0] : null
  return {
    tagCount,
    taggedArticleCount,
    topTagName: topTag?.name || '-'
  }
})

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getTagArticleStats()
    if (res && res.data) {
      chartData.value = res.data.list || []
      updateChart()
    }
  } catch (error) {
    // request interceptor 已处理错误提示
  } finally {
    loading.value = false
  }
}

const defaultColors = [
  BRAND_THEME_COLOR, '#67c23a', '#e6a23c', '#f56c6c',
  '#a855f7', '#6366f1', '#14b8a6', '#f472b6',
  '#8b5cf6', '#22c55e', '#0ea5e9', '#ef4444',
  '#3b82f6', '#10b981', '#f59e0b', '#ec4899'
]

const updateChart = () => {
  if (!chartInstance || chartData.value.length === 0) return

  const data = chartData.value.map((item, index) => {
    const color = item.color || defaultColors[index % defaultColors.length]
    return {
      name: item.name,
      value: item.count || 0,
      textStyle: { color },
      original: item
    }
  })

  chartInstance.setOption({
    tooltip: {
      show: true,
      formatter: (params: any) => {
        return `<div style="font-weight:600;margin-bottom:4px">${params.name}</div>
                <div>绑定文章：${Number(params.value || 0).toLocaleString()} 篇</div>`
      },
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: 'var(--app-brand-soft)',
      borderWidth: 1,
      textStyle: { color: 'var(--app-text-heading)' },
      extraCssText: 'box-shadow: 0 4px 12px rgba(0,0,0,0.1); border-radius: 8px;'
    },
    series: [
      {
        type: 'wordCloud' as any,
        shape: 'circle',
        left: 'center',
        top: 'center',
        width: '95%',
        height: '95%',
        sizeRange: [14, 72],
        rotationRange: [-30, 30],
        rotationStep: 15,
        gridSize: 10,
        drawOutOfBound: false,
        layoutAnimation: true,
        textStyle: {
          fontFamily: '"PingFang SC", "Microsoft YaHei", sans-serif',
          fontWeight: 'bold'
        },
        emphasis: {
          focus: 'self',
          textStyle: {
            shadowBlur: 10,
            shadowColor: 'rgba(0,0,0,0.15)'
          }
        },
        data
      }
    ]
  }, true)
}

// 具名 resize 回调：匿名函数无法被 removeEventListener 移除，反复进出页面会不断堆积监听器
const resizeChart = () => chartInstance?.resize()

const initChart = () => {
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)
  updateChart()
  window.addEventListener('resize', resizeChart)
}

onMounted(() => {
  fetchData()
  initChart()
})

onUnmounted(() => {
  window.removeEventListener('resize', resizeChart)
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
    border-radius: var(--app-card-radius);
    border: 1px solid var(--app-brand-soft);

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
    color: var(--app-text-heading);
    line-height: 1.2;
  }

  .stat-label {
    font-size: 13px;
    color: #909399;
    margin-top: 4px;
  }

  .chart-card {
    border-radius: var(--app-card-radius);
    border: 1px solid var(--app-brand-soft);

    :deep(.el-card__header) {
      padding: 16px 20px;
    }

    :deep(.el-card__body) {
      padding: 20px;
    }
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-weight: 600;
    color: var(--app-text-heading);
    font-size: 16px;

    .sub-title {
      font-size: 13px;
      color: #909399;
      font-weight: normal;
    }
  }

  .chart-container {
    width: 100%;
    height: 560px;
  }

  .wordcloud-chart {
    width: 100%;
    height: 100%;
  }
}

@media (max-width: 768px) {
  .page-container {
    .chart-container {
      height: 400px;
    }
  }
}
</style>
