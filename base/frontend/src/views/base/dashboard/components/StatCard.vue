<template>
  <div class="stat-card">
    <el-icon class="stat-icon" :style="{ color: color, backgroundColor: color + '1a' }">
      <component :is="icon" />
    </el-icon>
    <div class="stat-info">
      <div class="stat-value">{{ value }}</div>
      <div class="stat-title">{{ title }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  value: number | string
  icon: string
  color: string
}>()
</script>

<style scoped lang="scss">
.stat-card {
  display: flex;
  align-items: center;
}
.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 8px;
  font-size: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat-info {
  margin-left: 16px;
}
.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}
.stat-title {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}
</style>
