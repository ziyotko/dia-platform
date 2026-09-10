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

    <div class="page-card" style="margin-top:16px">
      <h3 style="margin-bottom:16px">快捷入口</h3>
      <div style="display:flex;gap:12px;flex-wrap:wrap">
        <el-button type="primary" @click="$router.push('/member/batches')">浏览可申报项目</el-button>
        <el-button @click="$router.push('/member/applications')">我的申报</el-button>
        <el-button @click="$router.push('/member/announcements')">结果公示</el-button>
        <el-button @click="$router.push('/member/certificates')">我的证书</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { memberApi } from '@/api/member'

const stats = ref<any>({})

const cards = computed(() => [
  { label: '我的申报', value: stats.value.totalApplications ?? 0, color: '#002fa7' },
  { label: '已提交', value: stats.value.submittedApplications ?? 0, color: '#4d6dc1' },
  { label: '通过立项', value: stats.value.passedApplications ?? 0, color: '#10b981' },
  { label: '获得证书', value: stats.value.certificates ?? 0, color: '#f59e0b' },
])

onMounted(async () => {
  const res = await memberApi.getDashboard()
  stats.value = res.data
})
</script>
