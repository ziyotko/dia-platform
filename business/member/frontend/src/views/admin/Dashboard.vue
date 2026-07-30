<template>
  <div class="admin-dash" v-loading="loading">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h3>控制台</h3>
        <p class="header-desc">欢迎回来，这是会员管理系统的数据概览</p>
      </div>
      <div class="header-right">
        <el-button text @click="refresh">
          <el-icon><Refresh /></el-icon> 刷新数据
        </el-button>
        <span class="update-time">上次更新：{{ lastUpdate }}</span>
      </div>
    </div>

    <!-- 统计卡片行 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6" :lg="4" v-for="s in stats" :key="s.label">
        <el-card class="stat-card" shadow="never" :style="{ '--card-color': s.color, '--card-color-light': s.bg }">
          <div class="stat-icon" :style="{ background: s.bg }">
            <el-icon :size="24" :color="s.color"><component :is="s.icon" /></el-icon>
          </div>
          <div class="stat-info">
            <count-up :num="s.value" class="value" />
            <p class="label">{{ s.label }}</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 第二行：会员概览 + 快捷操作 -->
    <el-row :gutter="20" class="row-section">
      <!-- 会员类型分布 -->
      <el-col :xs="24" :sm="24" :md="8">
        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="section-header">
              <span><el-icon :size="16" color="#3b82f6"><PieChart /></el-icon> 会员类型分布</span>
            </div>
          </template>
          <div class="dist-chart">
            <div class="dist-ring">
              <svg width="140" height="140" viewBox="0 0 140 140">
                <circle cx="70" cy="70" r="56" fill="none" stroke="#f1f5f9" stroke-width="14" />
                <circle cx="70" cy="70" r="56" fill="none" stroke="#60a5fa" stroke-width="14"
                  :stroke-dasharray="typeDist.unitPercent * 3.5186" stroke-dashoffset="0"
                  transform="rotate(-90, 70, 70)" stroke-linecap="round" />
                <circle cx="70" cy="70" r="56" fill="none" stroke="#fbbf24" stroke-width="14"
                  v-if="typeDist.personalPercent > 0"
                  :stroke-dasharray="typeDist.personalPercent * 3.5186"
                  :stroke-dashoffset="-(typeDist.unitPercent * 3.5186)"
                  transform="rotate(-90, 70, 70)" stroke-linecap="round" />
              </svg>
              <div class="ring-center">
                <p class="ring-total">{{ typeDist.total }}</p>
                <p class="ring-label">总会员</p>
              </div>
            </div>
            <div class="dist-legend">
              <div class="legend-item">
                <span class="dot" style="background:#60a5fa"></span>
                <span class="legend-label">单位会员</span>
                <span class="legend-value">{{ typeDist.unit }}</span>
                <span class="legend-pct">({{ typeDist.unitPercent }}%)</span>
              </div>
              <div class="legend-item">
                <span class="dot" style="background:#f59e0b"></span>
                <span class="legend-label">个人会员</span>
                <span class="legend-value">{{ typeDist.personal }}</span>
                <span class="legend-pct">({{ typeDist.personalPercent }}%)</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 状态分布 -->
      <el-col :xs="24" :sm="24" :md="8">
        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="section-header">
              <span><el-icon :size="16" color="#22c55e"><DataAnalysis /></el-icon> 状态分布</span>
            </div>
          </template>
          <div class="status-dist">
            <div class="status-bar-item" v-for="item in statusDist" :key="item.label">
              <div class="status-bar-header">
                <span class="status-label">{{ item.label }}</span>
                <span class="status-value">{{ item.value }}</span>
              </div>
              <el-progress
                :percentage="item.percent"
                :color="item.color"
                :stroke-width="10"
                :show-text="false"
                class="status-progress"
              />
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 快捷操作 -->
      <el-col :xs="24" :sm="24" :md="8">
        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="section-header">
              <span><el-icon :size="16" color="#8b5cf6"><Lightning /></el-icon> 快捷操作</span>
            </div>
          </template>
          <div class="quick-actions">
            <div class="action-item" v-for="act in quickActions" :key="act.label" @click="$router.push(act.path)">
              <div class="action-icon" :style="{ background: act.bg }">
                <el-icon :size="22" :color="act.color"><component :is="act.icon" /></el-icon>
              </div>
              <div class="action-info">
                <p class="action-title">{{ act.label }}</p>
                <p class="action-desc">{{ act.desc }}</p>
              </div>
              <el-icon class="action-arrow" color="#cbd5e1"><ArrowRight /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 第三行：最近注册 + 待办事项 -->
    <el-row :gutter="20" class="row-section">
      <el-col :xs="24" :md="16">
        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="section-header">
              <span><el-icon :size="16" color="#3b82f6"><UserFilled /></el-icon> 最近注册</span>
              <el-button text size="small" type="primary" @click="$router.push('/admin/members')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="recentMembers" stripe size="small" v-if="recentMembers.length">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="username" label="用户名" min-width="100" />
            <el-table-column prop="company_name" label="公司名称" min-width="140" show-overflow-tooltip />
            <el-table-column prop="member_type" label="类型" width="80">
              <template #default="{row}">
                <el-tag :type="row.member_type === 'unit' ? 'primary' : 'warning'" size="small" effect="plain">
                  {{ row.member_type === 'unit' ? '单位' : '个人' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="90">
              <template #default="{row}">
                <el-tag :type="statusTag(row.status)" size="small" effect="plain">{{ statusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="注册时间" width="160">
              <template #default="{row}">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>
          <el-empty v-else description="暂无数据" :image-size="80" />
        </el-card>
      </el-col>

      <el-col :xs="24" :md="8">
        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="section-header">
              <span><el-icon :size="16" color="#f59e0b"><Bell /></el-icon> 待办事项</span>
            </div>
          </template>
          <div class="todo-list">
            <div class="todo-item" v-for="(todo, i) in todos" :key="i" @click="todo.path && $router.push(todo.path)">
              <div class="todo-dot" :style="{ background: todo.color }"></div>
              <div class="todo-content">
                <p class="todo-text">{{ todo.text }}</p>
                <p class="todo-count" v-if="todo.count !== undefined">
                  <el-tag :type="todo.tag" size="small" effect="dark">{{ todo.count }} 项</el-tag>
                </p>
              </div>
              <el-icon v-if="todo.path" color="#cbd5e1" size="14"><ArrowRight /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import CountUp from '@/components/CountUp.vue'
import {
  UserFilled, User, Clock, CircleClose, Refresh,
  PieChart, DataAnalysis, Lightning, Bell,
  ArrowRight, DocumentChecked, Money, Medal, ChatDotRound,
  TrendCharts
} from '@element-plus/icons-vue'

const loading = ref(true)
const lastUpdate = ref('')
const recentMembers = ref<any[]>([])

// ===== 统计卡片 =====
const stats = ref([
  { icon: 'UserFilled', label: '会员总数', value: 0, color: '#3b82f6', bg: '#dbeafe' },
  { icon: 'User', label: '正式会员', value: 0, color: '#22c55e', bg: '#dcfce7' },
  { icon: 'TrendCharts', label: '今日新增', value: 0, color: '#8b5cf6', bg: '#ede9fe' },
  { icon: 'Clock', label: '待审核', value: 0, color: '#f59e0b', bg: '#fef3c7' },
  { icon: 'Money', label: '待缴费', value: 0, color: '#f97316', bg: '#fff7ed' },
  { icon: 'CircleClose', label: '已拒绝', value: 0, color: '#ef4444', bg: '#fee2e2' }
])

// ===== 类型分布 =====
const typeDist = reactive({
  unit: 0, personal: 0, total: 0,
  unitPercent: 0, personalPercent: 0
})

// ===== 状态分布 =====
const statusDist = ref([
  { label: '正式会员', value: 0, percent: 0, color: '#22c55e' },
  { label: '待审核', value: 0, percent: 0, color: '#f59e0b' },
  { label: '待缴费', value: 0, percent: 0, color: '#f97316' },
  { label: '已拒绝', value: 0, percent: 0, color: '#ef4444' }
])

// ===== 快捷操作 =====
const quickActions = [
  { icon: 'DocumentChecked', label: '入会审核', desc: '审核新会员申请', path: '/admin/applications', color: '#3b82f6', bg: '#dbeafe' },
  { icon: 'Money', label: '会费管理', desc: '查看会费缴纳记录', path: '/admin/fees', color: '#22c55e', bg: '#dcfce7' },
  { icon: 'Medal', label: '证书管理', desc: '管理会员证书发放', path: '/admin/certificates', color: '#8b5cf6', bg: '#ede9fe' },
  { icon: 'UserFilled', label: '会员管理', desc: '查看和管理所有会员', path: '/admin/members', color: '#f59e0b', bg: '#fef3c7' },
  { icon: 'ChatDotRound', label: '会员留言', desc: '查看和回复留言', path: '/admin/messages', color: '#ec4899', bg: '#fce7f3' },
  { icon: 'Bell', label: '公告管理', desc: '发布系统公告', path: '/admin/announcements', color: '#06b6d4', bg: '#cffafe' }
]

// ===== 待办事项 =====
const todos = ref([
  { text: '待审核入会申请', tag: 'warning' as const, color: '#f59e0b', count: 0, path: '/admin/applications' },
  { text: '待确认缴费记录', tag: 'danger' as const, color: '#f97316', count: 0, path: '/admin/fees' },
  { text: '待回复会员留言', tag: 'primary' as const, color: '#3b82f6', count: 0, path: '/admin/messages' },
  { text: '待处理证书申请', tag: 'primary' as const, color: '#8b5cf6', count: 0, path: '/admin/certificates' }
])

// ===== 工具函数 =====
function formatTime(t?: string) {
  if (!t) return '-'
  return t.slice(0, 16).replace('T', ' ')
}

const statusMap: Record<string, { l: string; t: string }> = {
  registering: { l: '注册中', t: 'info' },
  pending_review: { l: '待审核', t: 'warning' },
  pending_cert: { l: '待证书', t: 'warning' },
  pending_payment: { l: '待缴费', t: 'danger' },
  active: { l: '正式会员', t: 'success' },
  rejected: { l: '已拒绝', t: 'danger' }
}

function statusLabel(s: string) { return statusMap[s]?.l || s }
function statusTag(s: string) { return statusMap[s]?.t || 'info' }

function now() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')} ${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
}

async function fetchData() {
  loading.value = true
  try {
    // 1. 获取基础统计数据
    const statsRes = await adminApi.getMemberStats()
    const d = statsRes.data
    const total = d.total || 0
    const active = d.active || 0
    const pending = d.pending || 0
    const rejected = d.rejected || 0

    stats.value[0].value = total
    stats.value[1].value = active
    stats.value[2].value = d.today_new || 0
    stats.value[3].value = pending
    stats.value[4].value = d.pending_payment || 0
    stats.value[5].value = rejected

    // 2. 获取类型分布（并行查询）
    const [unitRes, personalRes] = await Promise.all([
      adminApi.getMembers({ member_type: 'unit', page: 1, size: 1 }).catch(() => ({ data: { total: 0 } })),
      adminApi.getMembers({ member_type: 'personal', page: 1, size: 1 }).catch(() => ({ data: { total: 0 } }))
    ])
    const unit = unitRes.data?.total || 0
    const personal = personalRes.data?.total || 0
    const distTotal = Math.max(unit + personal, 1)
    typeDist.unit = unit
    typeDist.personal = personal
    typeDist.total = unit + personal
    typeDist.unitPercent = Math.round(unit / distTotal * 100)
    typeDist.personalPercent = Math.round(personal / distTotal * 100)

    // 3. 状态分布
    const pendingPayment = d.pending_payment || 0
    const statuses = [
      { label: '正式会员', value: active, color: '#22c55e' },
      { label: '待审核', value: pending - pendingPayment > 0 ? pending - pendingPayment : pending, color: '#f59e0b' },
      { label: '待缴费', value: pendingPayment, color: '#f97316' },
      { label: '已拒绝', value: rejected, color: '#ef4444' }
    ]
    const maxVal = Math.max(...statuses.map(s => s.value), 1)
    statusDist.value = statuses.map(s => ({ ...s, percent: Math.round(s.value / maxVal * 100) }))

    // 4. 待办事项
    todos.value[0].count = pending - pendingPayment > 0 ? pending - pendingPayment : pending
    todos.value[1].count = pendingPayment
    todos.value[2].count = 0
    todos.value[3].count = 0

    // 5. 最近注册会员
    try {
      const memberRes = await adminApi.getMembers({ page: 1, size: 5 })
      recentMembers.value = memberRes.data?.list || []
    } catch {}

    lastUpdate.value = now()
  } catch (err) {
    console.error('Dashboard data fetch error:', err)
  } finally {
    loading.value = false
  }
}

function refresh() {
  fetchData()
}

onMounted(fetchData)
</script>

<style scoped lang="scss">
.admin-dash {
  max-width: 1400px;
  margin: 0 auto;
}

// ===== Page Header =====
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 24px;

  .header-left {
    h3 {
      font-size: 22px;
      font-weight: 700;
      color: #1f2937;
      margin: 0;
    }
    .header-desc {
      font-size: 13px;
      color: #94a3b8;
      margin-top: 4px;
    }
  }
  .header-right {
    display: flex;
    align-items: center;
    gap: 12px;
    .update-time {
      font-size: 12px;
      color: #94a3b8;
    }
  }
}

// ===== Stats Row =====
.stat-row {
  margin-bottom: 24px;
}

.stat-card {
  border-radius: 14px;
  border: 1px solid #f0f4f8;
  background: linear-gradient(135deg, #ffffff 0%, #fafcff 100%);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 3px;
    background: linear-gradient(90deg, var(--card-color, #3b82f6), var(--card-color-light, #93c5fd));
    opacity: 0;
    transition: opacity 0.3s;
  }

  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 12px 24px -8px rgba(0, 0, 0, 0.1);
    border-color: #e2e8f0;

    &::before { opacity: 1; }
  }

  :deep(.el-card__body) {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 18px 20px;
  }

  .stat-icon {
    width: 48px;
    height: 48px;
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    transition: transform 0.3s;
  }

  &:hover .stat-icon {
    transform: scale(1.08);
  }

  .stat-info {
    flex: 1;
    min-width: 0;

    .value {
      font-size: 24px;
      font-weight: 800;
      color: #0f172a;
      line-height: 1.2;
      font-variant-numeric: tabular-nums;
    }
    .label {
      font-size: 13px;
      color: #94a3b8;
      margin-top: 2px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
}

// ===== Section Card =====
.row-section {
  margin-bottom: 24px;
}

.section-card {
  border-radius: 14px;
  border: 1px solid #f0f4f8;
  margin-bottom: 20px;

  :deep(.el-card__header) {
    padding: 14px 20px;
    border-bottom: 1px solid #f5f5f5;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 14px;
    font-weight: 600;
    color: #1f2937;

    span {
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }

  :deep(.el-card__body) {
    padding: 20px;
  }
}

// ===== Distribution Chart (Donut) =====
.dist-chart {
  display: flex;
  align-items: center;
  gap: 24px;

  .dist-ring {
    position: relative;
    width: 140px;
    height: 140px;
    flex-shrink: 0;

    svg {
      width: 100%;
      height: 100%;
    }

    .ring-center {
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      text-align: center;

      .ring-total {
        font-size: 26px;
        font-weight: 800;
        color: #0f172a;
        line-height: 1;
      }
      .ring-label {
        font-size: 11px;
        color: #94a3b8;
        margin-top: 4px;
      }
    }
  }

  .dist-legend {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 12px;

    .legend-item {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 10px;
      border-radius: 8px;
      background: #f8fafc;
      transition: background 0.2s;

      &:hover { background: #f1f5f9; }

      .dot {
        width: 10px;
        height: 10px;
        border-radius: 50%;
        flex-shrink: 0;
      }
      .legend-label {
        flex: 1;
        font-size: 13px;
        color: #475569;
      }
      .legend-value {
        font-size: 15px;
        font-weight: 700;
        color: #0f172a;
      }
      .legend-pct {
        font-size: 12px;
        color: #94a3b8;
        min-width: 40px;
        text-align: right;
      }
    }
  }
}

// ===== Status Distribution =====
.status-dist {
  display: flex;
  flex-direction: column;
  gap: 16px;

  .status-bar-item {
    .status-bar-header {
      display: flex;
      justify-content: space-between;
      margin-bottom: 6px;

      .status-label {
        font-size: 13px;
        color: #475569;
      }
      .status-value {
        font-size: 14px;
        font-weight: 700;
        color: #0f172a;
      }
    }

    .status-progress {
      :deep(.el-progress-bar__outer) {
        background: #f1f5f9;
        border-radius: 8px;
      }
      :deep(.el-progress-bar__inner) {
        border-radius: 8px;
        transition: width 1s ease-in-out;
      }
    }
  }
}

// ===== Quick Actions =====
.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 4px;

  .action-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    border-radius: 10px;
    cursor: pointer;
    transition: all 0.2s;
    position: relative;

    &:hover {
      background: #f8fafc;
      transform: translateX(2px);

      .action-arrow {
        opacity: 1;
        transform: translateX(2px);
      }
    }

    .action-icon {
      width: 40px;
      height: 40px;
      border-radius: 10px;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
    }

    .action-info {
      flex: 1;
      min-width: 0;

      .action-title {
        font-size: 13px;
        font-weight: 600;
        color: #1f2937;
      }
      .action-desc {
        font-size: 11px;
        color: #94a3b8;
        margin-top: 1px;
      }
    }

    .action-arrow {
      opacity: 0;
      transition: all 0.2s;
    }
  }
}

// ===== Todo List =====
.todo-list {
  display: flex;
  flex-direction: column;
  gap: 4px;

  .todo-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    border-radius: 10px;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      background: #f8fafc;
      transform: translateX(2px);
    }

    .todo-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      flex-shrink: 0;
      animation: pulse 2s infinite;
    }

    .todo-content {
      flex: 1;
      display: flex;
      align-items: center;
      justify-content: space-between;

      .todo-text {
        font-size: 13px;
        color: #475569;
      }
    }
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

// ===== Responsive =====
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  .dist-chart {
    flex-direction: column;
    align-items: center;
  }
}
</style>
