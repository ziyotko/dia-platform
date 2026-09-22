<template>
  <div class="page-container">
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="24" :sm="12" :md="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon size="28"><Star /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ totalLike.toLocaleString() }}</div>
              <div class="stat-label">点赞量</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon size="28"><Share /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ totalShare.toLocaleString() }}</div>
              <div class="stat-label">分享量</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: var(--el-color-primary-light-9); color: var(--el-color-primary);">
              <el-icon size="28"><View /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ totalVisit.toLocaleString() }}</div>
              <div class="stat-label">浏览量</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover" class="chart-card">
      <template #header>
        <div class="card-header">
          <span>文章数据趋势</span>
          <div class="header-extra">
            <span class="total-info">点赞 / 分享 / 浏览 {{ periodText }}数据</span>
            <el-select v-if="period === 'year'" v-model="year" size="small" class="year-select">
              <el-option v-for="y in yearOptions" :key="y" :label="`${y}年`" :value="y" />
            </el-select>
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
import { Star, Share, View } from '@element-plus/icons-vue'
import { getArticleAnalyticsTrend } from '@/api/analytics'
import { BRAND_THEME_COLOR } from '@/stores/app'

const loading = ref(false)
const chartRef = ref<HTMLDivElement | null>(null)
let chartInstance: ECharts | null = null

const period = ref('week')
// 年份仅在 period=year 时生效（与仪表盘趋势一致）
const currentYear = new Date().getFullYear()
const year = ref(currentYear)
const yearOptions = computed(() => {
  const years: number[] = []
  for (let y = currentYear; y >= currentYear - 4; y--) {
    years.push(y)
  }
  return years
})
const labels = ref<string[]>([])
const likeData = ref<number[]>([])
const shareData = ref<number[]>([])
const visitData = ref<number[]>([])

const totalLike = computed(() => likeData.value.reduce((sum, v) => sum + (v || 0), 0))
const totalShare = computed(() => shareData.value.reduce((sum, v) => sum + (v || 0), 0))
const totalVisit = computed(() => visitData.value.reduce((sum, v) => sum + (v || 0), 0))

const periodText = computed(() => {
  const map: Record<string, string> = { week: '本周', month: '本月', year: '全年' }
  return map[period.value] || ''
})

const seriesMeta = [
  { name: '点赞量', color: '#f56c6c', key: 'like' },
  { name: '分享量', color: '#e6a23c', key: 'share' },
  { name: '浏览量', color: BRAND_THEME_COLOR, key: 'visit' }
]

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getArticleAnalyticsTrend(period.value, period.value === 'year' ? year.value : undefined)
    if (res && res.data) {
      labels.value = res.data.labels || []
      likeData.value = (res.data.like || []).map(Number)
      shareData.value = (res.data.share || []).map(Number)
      visitData.value = (res.data.visit || []).map(Number)
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

  const series = seriesMeta.map(meta => ({
    name: meta.name,
    type: 'line' as const,
    data: meta.key === 'like' ? likeData.value : meta.key === 'share' ? shareData.value : visitData.value,
    smooth: true,
    symbol: 'circle',
    symbolSize: 8,
    itemStyle: { color: meta.color },
    lineStyle: { width: 3, color: meta.color },
    areaStyle: {
      color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color: `${meta.color}4d` },
        { offset: 1, color: `${meta.color}0d` }
      ])
    }
  }))

  chartInstance.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'line' },
      formatter: (params: any) => {
        const list = Array.isArray(params) ? params : [params]
        const first = list[0]
        let html = `<div style="font-weight:600;margin-bottom:6px">${first.axisValueLabel || first.name}</div>`
        list.forEach((p: any) => {
          html += `<div style="display:flex;align-items:center;gap:6px;line-height:22px">
                    <span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:${p.color}"></span>
                    <span style="color:#606266">${p.seriesName}</span>
                    <span style="font-weight:600;color:var(--app-text-heading)">${Number(p.value).toLocaleString()}</span>
                  </div>`
        })
        return html
      },
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: 'var(--app-brand-soft)',
      borderWidth: 1,
      textStyle: { color: 'var(--app-text-heading)' },
      extraCssText: 'box-shadow: 0 4px 12px rgba(0,0,0,0.1); border-radius: 8px;'
    },
    legend: {
      top: 0,
      right: 0,
      icon: 'roundRect',
      itemWidth: 14,
      itemHeight: 8,
      textStyle: { color: '#606266' },
      data: seriesMeta.map(m => m.name)
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: '15%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: labels.value,
      axisLine: { lineStyle: { color: '#e4e7ed' } },
      axisLabel: {
        color: '#606266',
        interval: 0,
        rotate: labels.value.length > 12 ? 30 : 0
      },
      axisTick: { alignWithLabel: true }
    },
    yAxis: {
      type: 'value',
      name: '次数',
      minInterval: 1,
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: { lineStyle: { color: '#f2f6fc' } },
      axisLabel: { color: '#606266' }
    },
    series
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

watch(period, () => {
  fetchData()
})

watch(year, () => {
  if (period.value === 'year') {
    fetchData()
  }
})

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

    .header-extra {
      display: flex;
      align-items: center;
      gap: 16px;

      .year-select {
        width: 100px;
      }

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
