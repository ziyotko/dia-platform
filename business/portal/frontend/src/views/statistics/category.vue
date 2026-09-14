<template>
  <div class="page-container">
    <el-card shadow="hover" class="chart-card">
      <template #header>
        <div class="card-header">
          <span>分类文章统计</span>
          <div class="header-extra">
            <span class="total-info">总文章数：<strong>{{ totalCount }}</strong> 篇</span>
          </div>
        </div>
      </template>
      <div v-loading="loading" class="chart-wrapper">
        <div ref="chartRef" class="chart-container"></div>
        <div class="legend-panel">
          <div
            v-for="(item, index) in chartData"
            :key="item.categoryId"
            class="legend-item"
            @mouseenter="highlightSector(index)"
            @mouseleave="downplaySector(index)"
          >
            <span class="legend-dot" :style="{ background: colors[index % colors.length] }"></span>
            <span class="legend-name">{{ item.categoryName || '未分类' }}</span>
            <span class="legend-value">{{ item.count }} 篇</span>
            <span class="legend-percent">{{ percentOf(item.count) }}%</span>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import type { ECharts } from 'echarts'
import { getCategoryArticleStats } from '@/api/category'

const loading = ref(false)
const chartRef = ref<HTMLDivElement | null>(null)
let chartInstance: ECharts | null = null

const chartData = ref<any[]>([])
const totalCount = ref(0)

/** 百分比：总文章数为 0（暂无已发布文章）时返回 0.0，避免出现 NaN% */
const percentOf = (count: number) =>
  totalCount.value > 0 ? ((count / totalCount.value) * 100).toFixed(1) : '0.0'

const colors = [
  '#002fa7',
  '#67c23a',
  '#e6a23c',
  '#f56c6c',
  '#a855f7',
  '#6366f1',
  '#14b8a6',
  '#f472b6',
  '#8b5cf6',
  '#22c55e',
  '#0ea5e9',
  '#ef4444'
]

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getCategoryArticleStats()
    if (res && res.data) {
      chartData.value = res.data.list || []
      totalCount.value = res.data.total || 0
      nextTick(() => updateChart())
    }
  } catch (error) {
    // request interceptor 已处理错误提示
  } finally {
    loading.value = false
  }
}

const updateChart = () => {
  if (!chartInstance || chartData.value.length === 0) return

  const data = chartData.value.map(item => ({
    value: item.count,
    name: item.categoryName || '未分类'
  }))

  chartInstance.setOption({
    color: colors,
    tooltip: {
      trigger: 'item',
      formatter: (params: any) => {
        return `<div style="font-weight:600;margin-bottom:4px">${params.name}</div>
                <div>文章数：${params.value} 篇</div>
                <div>占比：${params.percent}%</div>`
      },
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#e6f2ff',
      borderWidth: 1,
      textStyle: { color: '#2c3e50' },
      extraCssText: 'box-shadow: 0 4px 12px rgba(0,0,0,0.1); border-radius: 8px;'
    },
    legend: { show: false },
    series: [
      {
        type: 'pie',
        radius: ['45%', '72%'],
        center: ['50%', '50%'],
        avoidLabelOverlap: true,
        itemStyle: {
          borderRadius: 8,
          borderColor: '#fff',
          borderWidth: 3
        },
        label: {
          show: true,
          formatter: '{b}\n{d}% ({c}篇)',
          color: '#606266',
          fontSize: 13,
          lineHeight: 18
        },
        labelLine: {
          show: true,
          length: 16,
          length2: 12,
          smooth: true,
          lineStyle: { color: '#c0c4cc' }
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 14,
            fontWeight: 'bold'
          },
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.2)'
          }
        },
        data
      }
    ],
    graphic: [
      {
        type: 'text',
        left: 'center',
        top: '42%',
        style: {
          text: '总文章数',
          textAlign: 'center',
          fill: '#909399',
          fontSize: 14
        }
      },
      {
        type: 'text',
        left: 'center',
        top: '52%',
        style: {
          text: String(totalCount.value),
          textAlign: 'center',
          fill: '#2c3e50',
          fontSize: 28,
          fontWeight: 'bold'
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

const highlightSector = (index: number) => {
  chartInstance?.dispatchAction({ type: 'highlight', seriesIndex: 0, dataIndex: index })
}

const downplaySector = (index: number) => {
  chartInstance?.dispatchAction({ type: 'downplay', seriesIndex: 0, dataIndex: index })
}

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
      .total-info {
        font-size: 14px;
        color: #606266;
        font-weight: normal;

        strong {
          color: #002fa7;
          font-size: 18px;
        }
      }
    }
  }

  .chart-wrapper {
    display: flex;
    align-items: center;
    gap: 24px;
    min-height: 500px;
  }

  .chart-container {
    flex: 1;
    height: 480px;
    min-width: 0;
  }

  .legend-panel {
    width: 260px;
    max-height: 480px;
    overflow-y: auto;
    padding: 12px;
    background: #f8fafc;
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    &::-webkit-scrollbar {
      width: 4px;
    }
    &::-webkit-scrollbar-thumb {
      background: #c0c4cc;
      border-radius: 2px;
    }

    .legend-item {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 10px 12px;
      border-radius: 8px;
      cursor: pointer;
      transition: background 0.2s;

      &:hover {
        background: #eef5ff;
      }

      .legend-dot {
        width: 10px;
        height: 10px;
        border-radius: 50%;
        flex-shrink: 0;
      }

      .legend-name {
        flex: 1;
        font-size: 13px;
        color: #2c3e50;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .legend-value {
        font-size: 13px;
        color: #606266;
        white-space: nowrap;
      }

      .legend-percent {
        font-size: 13px;
        color: #002fa7;
        font-weight: 600;
        white-space: nowrap;
        width: 52px;
        text-align: right;
      }
    }
  }
}

@media (max-width: 768px) {
  .page-container {
    .chart-wrapper {
      flex-direction: column;
    }

    .legend-panel {
      width: 100%;
      max-height: 280px;
    }
  }
}
</style>
