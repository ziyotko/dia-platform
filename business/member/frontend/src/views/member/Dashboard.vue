<template>
  <div class="dashboard" v-loading="loading">
    <!-- Status Card -->
    <el-card class="status-card" v-if="dash">
      <div class="status-header">
        <div>
          <h2>欢迎，{{ dash.member.company_name || dash.member.username }}</h2>
          <p class="status-label">
            会员状态：
            <el-tag :type="statusType" size="large">{{ statusText }}</el-tag>
          </p>
        </div>
        <el-button v-if="dash.member.status === 'registering'" type="primary" @click="$router.push({ name: 'Applications' })">继续完善资料</el-button>
        <el-button v-if="dash.member.status === 'pending_payment'" type="warning" @click="$router.push({ name: 'Fees' })">立即缴费</el-button>
      </div>
    </el-card>

    <!-- Stats Grid -->
    <el-row :gutter="20" class="stats-row" v-if="dash">
      <el-col :span="6" v-for="stat in stats" :key="stat.label">
        <el-card class="stat-card">
          <div class="stat-icon" :style="{ background: stat.bg }">
            <el-icon :size="24" :color="stat.color"><component :is="stat.icon" /></el-icon>
          </div>
          <div class="stat-info">
            <p class="stat-value">{{ stat.value }}</p>
            <p class="stat-label">{{ stat.label }}</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Two columns -->
    <el-row :gutter="20" v-if="dash">
      <el-col :span="14">
        <el-card>
          <template #header><span>最新公告</span></template>
          <div class="announce-mini" v-for="a in dash.latest_announcements" :key="a.id" @click="$router.push(`/announcements/${a.id}`)">
            <span class="title">{{ a.title }}</span>
            <span class="time">{{ formatDate(a.published_at) }}</span>
          </div>
          <el-empty v-if="!dash.latest_announcements?.length" description="暂无公告" />
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card>
          <template #header><span>会费概览</span></template>
          <div class="fee-summary" v-if="dash.fee_summary">
            <div class="fee-item">
              <span class="label">已缴年数</span>
              <span class="value">{{ dash.fee_summary.paid_count }} 年</span>
            </div>
            <div class="fee-item">
              <span class="label">待缴年数</span>
              <span class="value warn">{{ dash.fee_summary.unpaid_count }} 年</span>
            </div>
            <div class="fee-item">
              <span class="label">累计缴费</span>
              <span class="value">¥{{ dash.fee_summary.total_paid.toFixed(2) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { dashboardApi } from '@/api/index'

const dash = ref<any>(null)
const loading = ref(true)

const statusMap: Record<string, { text: string; type: string }> = {
  registering: { text: '注册中', type: 'info' },
  pending_review: { text: '待审核', type: 'warning' },
  pending_cert: { text: '待证书生成', type: 'warning' },
  pending_payment: { text: '待缴费', type: 'danger' },
  active: { text: '正式会员', type: 'success' },
  rejected: { text: '已拒绝', type: 'danger' },
  expired: { text: '已过期', type: 'info' }
}

const statusText = computed(() => statusMap[dash.value?.member?.status]?.text || dash.value?.member?.status)
const statusType = computed(() => statusMap[dash.value?.member?.status]?.type || 'info')

const stats = computed(() => [
  { icon: 'DocumentChecked', label: '我的申请', value: dash.value?.application?.label || '暂无', color: '#1a6fb5', bg: '#e8f4fd' },
  { icon: 'ChatDotRound', label: '未读消息', value: dash.value?.unread_messages || 0, color: '#f59e0b', bg: '#fef3c7' },
  { icon: 'EditPen', label: '我的文章', value: dash.value?.article_count || 0, color: '#22c55e', bg: '#dcfce7' },
  { icon: 'Money', label: '已缴年数', value: (dash.value?.fee_summary?.paid_count || 0) + ' 年', color: '#8b5cf6', bg: '#ede9fe' }
])

onMounted(async () => {
  try {
    const res = await dashboardApi.getMemberDashboard()
    dash.value = res.data
  } catch {} finally { loading.value = false }
})

function formatDate(d: string) { return d ? d.slice(0, 10) : '' }
</script>

<style scoped lang="scss">
.dashboard {
 width: 100%;
}
.status-card {
  margin-bottom: 20px;
  .status-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    h2 { font-size: 20px; margin-bottom: 8px; }
    .status-label { color: #6b7280; }
  }
}
.stats-row { margin-bottom: 20px; }
.stat-card {
  :deep(.el-card__body) {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px;
  }
  .stat-icon {
    width: 48px; height: 48px; border-radius: 12px;
    display: flex; align-items: center; justify-content: center;
  }
  .stat-info {
    .stat-value { font-size: 24px; font-weight: 700; color: #1f2937; }
    .stat-label { font-size: 13px; color: #9ca3af; margin-top: 4px; }
  }
}
.announce-mini {
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f3f4f6;
  cursor: pointer;
  &:hover { color: #1a6fb5; }
  .title { font-size: 14px; }
  .time { font-size: 12px; color: #9ca3af; white-space: nowrap; }
}
.fee-summary {
  .fee-item {
    display: flex; justify-content: space-between;
    padding: 14px 0;
    border-bottom: 1px solid #f3f4f6;
    .label { color: #6b7280; }
    .value { font-weight: 600; font-size: 16px; }
    .warn { color: #ef4444; }
  }
}
</style>
