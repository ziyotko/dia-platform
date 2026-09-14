<template>
  <div>
    <el-row :gutter="16">
      <el-col :span="6" v-for="card in cards" :key="card.label">
        <div class="stat-card" :style="{ background: card.color }">
          <div class="stat-value">{{ card.value }}</div>
          <div class="stat-label">{{ card.label }}</div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { adminApi } from '@/api/admin'

const stats = ref<any>({})

const cards = computed(() => {
  if (stats.value.pendingReviews !== undefined) {
    return [
      { label: '待评审任务', value: stats.value.pendingReviews ?? 0, color: '#002fa7' },
      { label: '已完成评审', value: stats.value.scoredReviews ?? 0, color: '#10b981' },
    ]
  }
  return [
    { label: '申报批次', value: stats.value.totalBatches ?? 0, color: '#002fa7' },
    { label: '申报总数', value: stats.value.totalApplications ?? 0, color: '#4d6dc1' },
    { label: '待初审', value: stats.value.pendingPreliminary ?? 0, color: '#f59e0b' },
    { label: '评审中', value: stats.value.underReview ?? 0, color: '#8097d3' },
    { label: '通过立项', value: stats.value.passed ?? 0, color: '#10b981' },
    { label: '申报人', value: stats.value.totalUsers ?? 0, color: '#6366f1' },
    { label: '颁发证书', value: stats.value.certificates ?? 0, color: '#f59e0b' },
    { label: '开放批次', value: stats.value.openBatches ?? 0, color: '#10b981' },
  ]
})

onMounted(async () => {
  const res = await adminApi.getDashboard()
  stats.value = res.data
})
</script>
