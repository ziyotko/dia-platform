<template>
  <div class="stat-card">
    <div class="stat-icon" :style="iconStyle">
      <el-icon :size="28"><component :is="icon" /></el-icon>
    </div>
    <div class="stat-info">
      <div class="stat-value">{{ value }}</div>
      <div class="stat-title">{{ title }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  title: string
  value: number | string
  icon: string
  color: string
}>()

const iconStyle = computed(() => ({
  color: props.color,
  backgroundColor: `${props.color}1a`,
  boxShadow: `0 8px 20px ${props.color}26`
}))
</script>

<style scoped lang="scss">
.stat-card {
  display: flex;
  align-items: center;
  padding: 8px 4px;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: transform 0.3s ease;
}

.stat-info {
  margin-left: 16px;
  min-width: 0;
}

.stat-value {
  font-size: 26px;
  font-weight: 700;
  color: #1e293b;
  line-height: 1.2;
}

.stat-title {
  font-size: 13px;
  color: #94a3b8;
  margin-top: 6px;
  font-weight: 500;
}
</style>
