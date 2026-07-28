<template>
  <div class="admin-dash" v-loading="loading">
    <el-row :gutter="20">
      <el-col :span="6" v-for="s in stats" :key="s.label">
        <el-card class="stat-card">
          <div class="stat-icon" :style="{ background: s.bg }">
            <el-icon :size="28" :color="s.color"><component :is="s.icon" /></el-icon>
          </div>
          <div class="stat-info">
            <p class="value">{{ s.value }}</p>
            <p class="label">{{ s.label }}</p>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'

const loading = ref(true)
const stats = ref([
  { icon: 'UserFilled', label: '会员总数', value: 0, color: '#3b82f6', bg: '#dbeafe' },
  { icon: 'User', label: '正式会员', value: 0, color: '#22c55e', bg: '#dcfce7' },
  { icon: 'Clock', label: '待处理', value: 0, color: '#f59e0b', bg: '#fef3c7' },
  { icon: 'CircleClose', label: '已拒绝', value: 0, color: '#ef4444', bg: '#fee2e2' }
])

onMounted(async () => {
  try {
    const res = await adminApi.getMemberStats()
    const d = res.data
    stats.value[0].value = d.total || 0
    stats.value[1].value = d.active || 0
    stats.value[2].value = d.pending || 0
    stats.value[3].value = d.rejected || 0
  } catch {} finally { loading.value = false }
})
</script>

<style scoped lang="scss">
.admin-dash { max-width: 1000px; }
.stat-card {
  :deep(.el-card__body) { display: flex; align-items: center; gap: 16px; padding: 24px; }
  .stat-icon { width: 52px; height: 52px; border-radius: 12px; display: flex; align-items: center; justify-content: center; }
  .stat-info {
    .value { font-size: 28px; font-weight: 700; }
    .label { font-size: 13px; color: #9ca3af; margin-top: 4px; }
  }
}
</style>
