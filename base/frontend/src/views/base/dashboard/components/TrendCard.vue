<template>
  <el-card class="trend-card" shadow="never">
    <template #header>
      <div class="trend-header">
        <span>{{ title }}</span>
        <span class="trend-total">近 7 天合计 {{ total }}</span>
      </div>
    </template>
    <div class="trend-bars">
      <div v-for="point in points" :key="point.date" class="trend-item">
        <div class="trend-value">{{ point.count }}</div>
        <div class="trend-bar-track">
          <div class="trend-bar" :style="{ height: barHeight(point.count), backgroundColor: color }" />
        </div>
        <div class="trend-label">{{ shortDate(point.date) }}</div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { DailyPoint } from '@/api/dashboard'

const props = withDefaults(
  defineProps<{
    title: string
    points: DailyPoint[]
    color?: string
  }>(),
  { color: '#409EFF' }
)

const total = computed(() => props.points.reduce((sum, p) => sum + (p.count || 0), 0))

// 以最大值为基准计算柱高（最小 4%，避免全 0 时看不到柱子）
const maxCount = computed(() => Math.max(1, ...props.points.map((p) => p.count || 0)))

const barHeight = (count: number) => {
  const ratio = (count || 0) / maxCount.value
  return `${Math.max(4, Math.round(ratio * 100))}%`
}

const shortDate = (date: string) => (date?.length >= 10 ? date.slice(5).replace('-', '/') : date)
</script>

<style scoped lang="scss">
.trend-card {
  border-radius: 14px;
  border: none;

  :deep(.el-card__header) {
    padding: 16px 20px;
    border-bottom: 1px solid #f1f5f9;
  }

  :deep(.el-card__body) {
    padding: 20px;
  }
}

.trend-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;

  .trend-total {
    font-size: 12px;
    font-weight: 400;
    color: #94a3b8;
  }
}

.trend-bars {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  height: 160px;
}

.trend-item {
  flex: 1;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;

  .trend-value {
    font-size: 12px;
    color: #64748b;
    margin-bottom: 4px;
  }

  .trend-bar-track {
    flex: 1;
    width: 100%;
    display: flex;
    align-items: flex-end;
    background: #f8fafc;
    border-radius: 6px;
    overflow: hidden;
  }

  .trend-bar {
    width: 100%;
    border-radius: 6px 6px 0 0;
    transition: height 0.3s ease;
  }

  .trend-label {
    margin-top: 6px;
    font-size: 12px;
    color: #94a3b8;
  }
}
</style>
