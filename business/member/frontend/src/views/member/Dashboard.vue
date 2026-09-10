<template>
  <div class="dashboard" v-loading="loading">
    <!-- ═══ Welcome Hero Banner ═══ -->
    <el-card class="hero-card" v-if="dash" shadow="never">
      <div class="hero-inner">
        <div class="hero-avatar">
          <div class="avatar-circle" :style="{ background: avatarBg }">
            <span class="avatar-text">{{ avatarLetter }}</span>
          </div>
        </div>
        <div class="hero-info">
          <div class="hero-top">
            <h2 class="hero-greeting">{{ greeting }}，{{ displayName }}</h2>
            <!-- 机构徽章统一显示：机构名称 + 级别标签，避免重复展示 -->
            <template v-for="org in dash?.organizations" :key="org.id">
              <div class="org-badge" :class="org.is_fee_based ? 'org-badge-fee' : 'org-badge-join'">
                <span class="org-badge-name">{{ org.org_name }}</span>
                <el-tag v-if="org.level_name" size="small" effect="plain" class="org-badge-level">{{ org.level_name }}</el-tag>
              </div>
            </template>
          </div>
          <div class="hero-meta">
            <el-tag :type="statusType" size="small" effect="light" class="status-tag-mini">
              <el-icon style="margin-right:3px;vertical-align:-1px"><VideoPause /></el-icon>
              {{ statusText }}
            </el-tag>
            <span class="meta-item">
              <el-icon><User /></el-icon>
              {{ memberTypeLabel }}
            </span>
          </div>
          <div class="hero-actions">
            <el-button
              v-if="dash.member.status === 'registering'"
              type="primary" size="small"
              @click="$router.push({ name: 'Applications' })"
            >
              <el-icon><Edit /></el-icon> 申请入会
            </el-button>
            <el-button
              v-if="dash.member.status === 'pending_payment'"
              type="warning" size="small"
              @click="$router.push({ name: 'Fees' })"
            >
              <el-icon><Coin /></el-icon> 立即缴费
            </el-button>
            <el-button size="small" plain @click="$router.push({ name: 'Profile' })">
              <el-icon><User /></el-icon> 编辑资料
            </el-button>
          </div>
        </div>
      </div>
    </el-card>

    <!-- ═══ Quick Actions ═══ -->
    <el-row :gutter="16" class="quick-actions-row" v-if="dash">
      <el-col :xs="8" :sm="8" :md="6" :lg="4" v-for="action in quickActions" :key="action.name">
        <el-card class="quick-action-card" shadow="hover" @click="$router.push({ name: action.name })">
          <div class="qa-icon" :style="{ background: action.bg, color: action.color }">
            <el-icon :size="22"><component :is="action.icon" /></el-icon>
          </div>
          <span class="qa-label">{{ action.label }}</span>
        </el-card>
      </el-col>
    </el-row>

    <!-- ═══ Stats Grid ═══ -->
    <el-row :gutter="20" class="stats-row" v-if="dash">
      <el-col :xs="12" :sm="12" :md="6" v-for="stat in stats" :key="stat.label">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-body">
            <div class="stat-icon" :style="{ background: stat.bg }">
              <el-icon :size="24" :color="stat.color"><component :is="stat.icon" /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-value">{{ stat.value }}</p>
              <p class="stat-label">{{ stat.label }}</p>
            </div>
          </div>
          <div class="stat-trend" :style="{ color: stat.color }">
            <el-icon><Top /></el-icon>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- ═══ Two Columns: Announcements + Fee & Profile ═══ -->
    <el-row :gutter="20" class="main-row" v-if="dash">
      <!-- Left: Announcements -->
      <el-col :xs="24" :md="14">
        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="section-header">
              <span><el-icon style="vertical-align:-3px;margin-right:6px"><Bell /></el-icon>最新公告</span>
              <el-button text type="primary" size="small" @click="$router.push({ name: 'Announcements' })">
                查看全部 <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <div class="announce-list">
            <div class="announce-item" v-for="a in dash.latest_announcements" :key="a.id" @click="$router.push(`/announcements/${a.id}`)">
              <div class="announce-left">
                <el-tag v-if="a.is_pinned" size="small" type="danger" effect="dark" class="pin-tag">置顶</el-tag>
                <el-tag v-else :type="announceTypeTag(a.type)" size="small" effect="plain" class="type-tag">
                  {{ announceTypeLabel(a.type) }}
                </el-tag>
                <span class="announce-title">{{ a.title }}</span>
              </div>
              <div class="announce-right">
                <span class="announce-views">
                  <el-icon style="vertical-align:-2px;margin-right:2px"><View /></el-icon>
                  {{ a.view_count || 0 }}
                </span>
                <span class="announce-time">{{ formatDate(a.published_at) }}</span>
              </div>
            </div>
            <el-empty v-if="!dash.latest_announcements?.length" description="暂无公告" :image-size="80" />
          </div>
        </el-card>
      </el-col>

      <!-- Right: Member Profile + Fee Summary -->
      <el-col :xs="24" :md="10">
        <!-- Member Profile Mini -->
        <el-card class="section-card profile-mini-card" shadow="never">
          <template #header>
            <span><el-icon style="vertical-align:-3px;margin-right:6px"><InfoFilled /></el-icon>基本信息</span>
          </template>
          <div class="profile-mini">
            <div class="profile-row" v-if="dash.member.company_name">
              <span class="profile-label">单位名称</span>
              <span class="profile-value">{{ dash.member.company_name }}</span>
            </div>
            <div class="profile-row">
              <span class="profile-label">联系人姓名</span>
              <span class="profile-value">{{ dash.member.contact_person || '未设置' }}</span>
            </div>
            <div class="profile-row">
              <span class="profile-label">手机号码</span>
              <span class="profile-value">{{ dash.member.mobile || '未设置' }}</span>
            </div>
            <div class="profile-row">
              <span class="profile-label">电子邮箱</span>
              <span class="profile-value">{{ dash.member.email || '未设置' }}</span>
            </div>
            <div class="profile-row">
              <span class="profile-label">会员类型</span>
              <el-tag :type="dash.member.member_type === 'unit' ? 'primary' : 'success'" size="small">
                {{ memberTypeLabel }}
              </el-tag>
            </div>
          </div>
        </el-card>

        <!-- Fee Summary -->
        <el-card class="section-card fee-card" shadow="never" style="margin-top:16px">
          <template #header>
            <span><el-icon style="vertical-align:-3px;margin-right:6px"><Coin /></el-icon>会费概览</span>
          </template>
          <div class="fee-summary" v-if="dash.fee_summary">
            <div class="fee-progress-section">
              <div class="fee-progress-header">
                <span class="fee-progress-label">缴费进度</span>
                <span class="fee-progress-value">
                  {{ dash.fee_summary.paid_count }} / {{ dash.fee_summary.paid_count + dash.fee_summary.unpaid_count }} 年
                </span>
              </div>
              <el-progress
                :percentage="feeProgress"
                :color="feeProgress === 100 ? '#22c55e' : feeProgress > 50 ? '#4d6dc1' : '#f59e0b'"
                :stroke-width="10"
                striped
                striped-flow
              />
            </div>
            <div class="fee-items">
              <div class="fee-item">
                <div class="fee-item-left">
                  <el-icon color="#22c55e"><CircleCheckFilled /></el-icon>
                  <span>已缴年数</span>
                </div>
                <span class="fee-item-value success">{{ dash.fee_summary.paid_count }} 年</span>
              </div>
              <div class="fee-item">
                <div class="fee-item-left">
                  <el-icon color="#f59e0b"><WarningFilled /></el-icon>
                  <span>待缴年数</span>
                </div>
                <span class="fee-item-value warn">{{ dash.fee_summary.unpaid_count }} 年</span>
              </div>
              <div class="fee-item fee-item-total">
                <div class="fee-item-left">
                  <el-icon color="#8b5cf6"><Money /></el-icon>
                  <span>累计缴费</span>
                </div>
                <span class="fee-item-value total">¥{{ dash.fee_summary.total_paid.toFixed(2) }}</span>
              </div>
            </div>
            <el-button
              v-if="dash.fee_summary.unpaid_count > 0"
              type="warning" class="fee-pay-btn" size="default" plain
              @click="$router.push({ name: 'Fees' })"
            >
              <el-icon><Coin /></el-icon> 去缴费
            </el-button>
          </div>
          <el-empty v-else description="暂无缴费记录" :image-size="60" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { dashboardApi } from '@/api/index'

const router = useRouter()
const dash = ref<any>(null)
const loading = ref(true)

// ── Status ──
const statusMap: Record<string, { text: string; type: string }> = {
  registering: { text: '注册会员', type: 'info' },
  pending_review: { text: '入会待审核', type: 'warning' },
  pending_payment: { text: '待缴费', type: 'danger' },
  active: { text: '正式会员', type: 'success' },
  rejected: { text: '已拒绝', type: 'danger' },
  expired: { text: '已过期', type: 'info' }
}

const statusText = computed(() => statusMap[dash.value?.member?.status]?.text || dash.value?.member?.status)
const statusType = computed(() => statusMap[dash.value?.member?.status]?.type || 'info')

// ── Greeting ──
const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '夜深了'
  if (h < 9) return '早上好'
  if (h < 12) return '上午好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const displayName = computed(() => dash.value?.member?.company_name || dash.value?.member?.username || '用户')

// ── Avatar ──
const avatarLetter = computed(() => {
  const name = displayName.value
  return name ? name.charAt(0).toUpperCase() : 'M'
})

const avatarColors = ['#002fa7', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6']
const avatarBg = computed(() => {
  const name = displayName.value
  if (!name) return avatarColors[0]
  let hash = 0
  for (let i = 0; i < name.length; i++) { hash = name.charCodeAt(i) + ((hash << 5) - hash) }
  return avatarColors[Math.abs(hash) % avatarColors.length]
})

// ── Member Type ──
const memberTypeLabel = computed(() => {
  const t = dash.value?.member?.member_type
  return t === 'unit' ? '单位会员' : t === 'personal' ? '个人会员' : t || '未知'
})

// ── Quick Actions ──
const quickActions = [
  { name: 'Applications', icon: 'DocumentChecked', label: '我的申请', bg: '#e6eaf6', color: '#002fa7' },
  { name: 'Fees', icon: 'Coin', label: '会费管理', bg: '#fef3c7', color: '#f59e0b' },
  { name: 'Certificates', icon: 'Medal', label: '我的证书', bg: '#dcfce7', color: '#22c55e' },
  { name: 'Articles', icon: 'EditPen', label: '我的文章', bg: '#ede9fe', color: '#8b5cf6' },
  { name: 'Messages', icon: 'ChatDotRound', label: '会员留言', bg: '#fce7f3', color: '#ec4899' },
  { name: 'Organizations', icon: 'OfficeBuilding', label: '加入信息', bg: '#e0f2fe', color: '#0ea5e9' },
]

// ── Stats ──
const stats = computed(() => [
  { icon: 'DocumentChecked', label: '我的申请', value: dash.value?.application?.label || '暂无', color: '#002fa7', bg: '#ccd5ed' },
  { icon: 'ChatDotRound', label: '未读消息', value: dash.value?.unread_messages || 0, color: '#f59e0b', bg: '#fef3c7' },
  { icon: 'EditPen', label: '我的文章', value: dash.value?.article_count || 0, color: '#22c55e', bg: '#dcfce7' },
  { icon: 'Coin', label: '已缴年数', value: (dash.value?.fee_summary?.paid_count || 0) + ' 年', color: '#8b5cf6', bg: '#ede9fe' }
])

// ── Fee ──
const feeProgress = computed(() => {
  const paid = dash.value?.fee_summary?.paid_count || 0
  const unpaid = dash.value?.fee_summary?.unpaid_count || 0
  const total = paid + unpaid
  return total === 0 ? 0 : Math.round((paid / total) * 100)
})

// ── Announcement helpers ──
const announceTypeMap: Record<string, { label: string; tag: string }> = {
  notice: { label: '通知', tag: 'primary' },
  article: { label: '文章', tag: 'success' },
  policy: { label: '政策', tag: 'warning' }
}

function announceTypeLabel(type: string) { return announceTypeMap[type]?.label || type || '公告' }
function announceTypeTag(type: string) { return announceTypeMap[type]?.tag || '' }

// ── Lifecycle ──
onMounted(async () => {
  try {
    const res = await dashboardApi.getMemberDashboard()
    dash.value = res.data
  } catch {} finally { loading.value = false }
})

// ── Utilities ──
function formatDate(d: string) { return d ? d.slice(0, 10) : '' }
</script>

<style scoped lang="scss">
// ── Variables ──
$primary: #002fa7;
$success: #22c55e;
$warning: #f59e0b;
$danger: #ef4444;
$purple: #8b5cf6;
$text: #1f2937;
$text-secondary: #6b7280;
$text-light: #9ca3af;
$border: #e5e7eb;
$bg-card: #ffffff;
$radius-sm: 8px;
$radius: 12px;
$radius-lg: 16px;
$shadow: 0 1px 3px rgba(0,0,0,0.06);
$shadow-hover: 0 4px 16px rgba(0,0,0,0.08);

.dashboard {
  width: 100%;
  padding-bottom: 24px;
}

// ════════════════════════════════════════
// Hero Banner — light pastel theme
// ════════════════════════════════════════
.hero-card {
  margin-bottom: 20px;
  border: 1px solid $border !important;
  border-radius: $radius-lg !important;
  background: linear-gradient(135deg, #f0f4ff 0%, #e8f0fe 100%) !important;
  overflow: hidden;
  position: relative;

  &::before {
    content: '';
    position: absolute;
    top: -40%;
    right: -10%;
    width: 300px;
    height: 300px;
    background: radial-gradient(circle, rgba(59,130,246,0.06) 0%, transparent 70%);
    border-radius: 50%;
  }

  :deep(.el-card__body) { padding: 24px 28px; }
}

.hero-inner {
  display: flex;
  align-items: center;
  gap: 20px;
  position: relative;
  z-index: 1;
}

.hero-avatar {
  flex-shrink: 0;
}

.avatar-circle {
  width: 68px;
  height: 68px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 3px solid rgba(255,255,255,0.9);
  box-shadow: 0 4px 14px rgba(59,130,246,0.15);
}

.avatar-text {
  font-size: 28px;
  font-weight: 700;
  color: #fff;
}

.hero-info {
  flex: 1;
  min-width: 0;
}

.hero-top {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.hero-greeting {
  font-size: 20px;
  font-weight: 700;
  color: $text;
  white-space: nowrap;
}

.hero-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.status-tag-mini {
  font-weight: 500;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: $text-secondary;
  .el-icon { font-size: 14px; color: $primary; }
}

.meta-divider {
  width: 1px;
  height: 14px;
  background: $border;
}

.hero-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;

  .el-button {
    transition: all 0.25s;
    &:hover {
      transform: translateY(-1px);
      box-shadow: $shadow-hover;
    }
  }

  .el-button--primary {
    background: linear-gradient(135deg, #ccd5ed, #b3c1e5) !important;
    border: 1px solid #8097d3 !important;
    color: #1e40af !important;
    &:hover {
      background: linear-gradient(135deg, #b3c1e5, #8097d3) !important;
    }
  }

  .el-button--warning {
    background: linear-gradient(135deg, #fef3c7, #fde68a) !important;
    border: 1px solid #fcd34d !important;
    color: #92400e !important;
    &:hover {
      background: linear-gradient(135deg, #fde68a, #fcd34d) !important;
    }
  }

  .el-button.is-plain {
    background: #fff !important;
    border: 1px solid $border !important;
    color: $text-secondary !important;
    &:hover {
      color: $primary !important;
      border-color: $primary !important;
    }
  }
}

// ════════════════════════════════════════
// Quick Actions
// ════════════════════════════════════════
.quick-actions-row {
  margin-bottom: 20px;
}

.quick-action-card {
  cursor: pointer;
  border-radius: $radius !important;
  transition: all 0.25s !important;
  margin-bottom: 8px;

  &:hover {
    transform: translateY(-4px);
    box-shadow: $shadow-hover !important;
  }

  :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    padding: 18px 8px;
    text-align: center;
  }
}

.qa-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.25s;

  .quick-action-card:hover & {
    transform: scale(1.1);
  }
}

.qa-label {
  font-size: 13px;
  font-weight: 600;
  color: $text;
  white-space: nowrap;
}

// ════════════════════════════════════════
// Stats Row
// ════════════════════════════════════════
.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  border-radius: $radius !important;
  transition: all 0.25s !important;
  cursor: default;
  margin-bottom: 8px;

  &:hover {
    transform: translateY(-3px);
    box-shadow: $shadow-hover !important;
  }

  :deep(.el-card__body) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px !important;
  }
}

.stat-body {
  display: flex;
  align-items: center;
  gap: 14px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-info {
  .stat-value {
    font-size: 24px;
    font-weight: 700;
    color: $text;
    line-height: 1.2;
  }
  .stat-label {
    font-size: 13px;
    color: $text-light;
    margin-top: 4px;
  }
}

.stat-trend {
  opacity: 0.3;
  font-size: 18px;
}

// ════════════════════════════════════════
// Section Cards — equal height columns
// ════════════════════════════════════════
.main-row {
  margin-bottom: 0;
  display: flex;
  flex-wrap: wrap;

  > .el-col {
    display: flex;
    flex-direction: column;
  }
}

.section-card {
  border-radius: $radius !important;
  margin-bottom: 8px;

  :deep(.el-card__header) {
    padding: 16px 20px;
    border-bottom: 1px solid $border;
    font-weight: 600;
    font-size: 15px;
    flex-shrink: 0;
  }

  :deep(.el-card__body) {
    padding: 0 20px 20px;
  }
}

// Left-column card stretches to match right column
.main-row > .el-col:first-child .section-card {
  flex: 1;
  display: flex;
  flex-direction: column;

  :deep(.el-card__body) {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .announce-list {
    flex: 1;
    display: flex;
    flex-direction: column;

    .el-empty {
      flex: 1;
      display: flex;
      flex-direction: column;
      justify-content: center;
    }
  }
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

// ════════════════════════════════════════
// Announcements
// ════════════════════════════════════════
.announce-list {
  padding-top: 4px;
}

.announce-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 0;
  border-bottom: 1px solid #f3f4f6;
  cursor: pointer;
  transition: all 0.2s;

  &:last-child { border-bottom: none; }

  &:hover {
    padding-left: 6px;
    .announce-title { color: $primary; }
  }
}

.announce-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.pin-tag {
  flex-shrink: 0;
  font-size: 11px;
  padding: 0 6px;
  height: 20px;
  line-height: 20px;
}

.type-tag {
  flex-shrink: 0;
  font-size: 11px;
  padding: 0 6px;
  height: 20px;
  line-height: 20px;
}

.announce-title {
  font-size: 14px;
  color: $text;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: color 0.2s;
}

.announce-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
  margin-left: 12px;
}

.announce-views {
  font-size: 12px;
  color: $text-light;
  white-space: nowrap;
}

.announce-time {
  font-size: 12px;
  color: $text-light;
  white-space: nowrap;
}

// ════════════════════════════════════════
// Profile Mini
// ════════════════════════════════════════
.profile-mini-card {
  :deep(.el-card__body) { padding: 16px 20px !important; }
}

.profile-mini {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.profile-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 11px 0;
  border-bottom: 1px solid #f8f9fa;

  &:last-child { border-bottom: none; }
}

.profile-label {
  font-size: 13px;
  color: $text-secondary;
  flex-shrink: 0;
}

.profile-value {
  font-size: 13px;
  font-weight: 500;
  color: $text;
  text-align: right;
  max-width: 60%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

// ════════════════════════════════════════
// Fee Summary
// ════════════════════════════════════════
.fee-card {
  :deep(.el-card__body) { padding: 16px 20px 20px !important; }
}

.fee-summary {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.fee-progress-section {
  padding: 12px 0 16px;
  border-bottom: 1px solid #f3f4f6;
}

.fee-progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;

  .fee-progress-label {
    font-size: 13px;
    color: $text-secondary;
  }

  .fee-progress-value {
    font-size: 13px;
    font-weight: 600;
    color: $text;
  }
}

.fee-items {
  padding: 4px 0;
}

.fee-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f8f9fa;

  &:last-child { border-bottom: none; }
}

.fee-item-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: $text-secondary;
}

.fee-item-value {
  font-weight: 600;
  font-size: 15px;

  &.success { color: $success; }
  &.warn { color: $warning; }
  &.total { color: $purple; font-size: 17px; }
}

.fee-item-total {
  padding: 14px 0 8px;
}

.fee-pay-btn {
  margin-top: 12px;
  width: 100%;
  border-radius: $radius-sm !important;
  font-weight: 600;
  transition: all 0.25s;
  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(245,158,11,0.3);
  }
}

// ════════════════════════════════════════
// Responsive
// ════════════════════════════════════════
// ════════════════════════════════════════
// Organization Badges (in hero-top)
// ════════════════════════════════════════
.org-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}
.org-badge-fee {
  background: #dcfce7;
  border: 1px solid #86efac;
  color: #166534;
}
.org-badge-join {
  background: #e0f2fe;
  border: 1px solid #7dd3fc;
  color: #1e40af;
}
.org-badge-name {
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
}
.org-badge-level {
  font-size: 11px !important;
  height: 18px !important;
  line-height: 18px !important;
  padding: 0 5px !important;
}

@media (max-width: 768px) {
  .hero-card :deep(.el-card__body) { padding: 20px 16px; }
  .hero-inner { flex-direction: column; text-align: center; }
  .hero-top { justify-content: center; }
  .hero-meta { justify-content: center; }
  .hero-actions { justify-content: center; }
  .avatar-circle { width: 56px; height: 56px; }
  .avatar-text { font-size: 22px; }
  .hero-greeting { font-size: 17px; }
  .announce-item { flex-direction: column; align-items: flex-start; gap: 6px; }
  .announce-right { margin-left: 0; width: 100%; justify-content: flex-end; }

  .org-badge-name {
    max-width: 70px;
  }
}
</style>
