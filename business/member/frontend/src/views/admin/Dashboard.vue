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
              <span><el-icon :size="16" color="#002fa7"><PieChart /></el-icon> 会员类型分布</span>
            </div>
          </template>
          <div ref="typeChartRef" class="chart-container"></div>
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
          <div ref="statusChartRef" class="chart-container"></div>
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
              <span><el-icon :size="16" color="#002fa7"><UserFilled /></el-icon> 最近注册</span>
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
import { ref, reactive, onMounted, onBeforeUnmount, nextTick } from 'vue'
import * as echarts from 'echarts'
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
  { icon: 'UserFilled', label: '会员总数', value: 0, color: '#002fa7', bg: '#ccd5ed' },
  { icon: 'User', label: '正式会员', value: 0, color: '#22c55e', bg: '#dcfce7' },
  { icon: 'TrendCharts', label: '今日新增', value: 0, color: '#8b5cf6', bg: '#ede9fe' },
  { icon: 'Clock', label: '待审核', value: 0, color: '#f59e0b', bg: '#fef3c7' },
  { icon: 'Money', label: '待缴费', value: 0, color: '#f97316', bg: '#fff7ed' },
  { icon: 'CircleClose', label: '已拒绝', value: 0, color: '#ef4444', bg: '#fee2e2' }
])

// ===== ECharts =====
const typeChartRef = ref<HTMLElement>()
const statusChartRef = ref<HTMLElement>()
let typeChart: echarts.ECharts | null = null
let statusChart: echarts.ECharts | null = null

const typeData = reactive({ unit: 0, personal: 0 })

function initCharts() {
  if (typeChartRef.value) {
    typeChart = echarts.init(typeChartRef.value)
    typeChart.setOption(typeChartOption())
  }
  if (statusChartRef.value) {
    statusChart = echarts.init(statusChartRef.value)
    statusChart.setOption(statusChartOption([]))
  }
}

function typeChartOption() {
  const total = typeData.unit + typeData.personal
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: {
      bottom: 0,
      left: 'center',
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { fontSize: 12, color: '#64748b' }
    },
    series: [{
      type: 'pie',
      radius: ['50%', '75%'],
      avoidLabelOverlap: false,
      padAngle: 2,
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { show: false },
      emphasis: {
        label: { show: true, fontSize: 14, fontWeight: 'bold' },
        itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0,0,0,0.2)' }
      },
      data: [
        { value: typeData.unit, name: '单位会员', itemStyle: { color: '#4d6dc1' } },
        { value: typeData.personal, name: '个人会员', itemStyle: { color: '#fbbf24' } }
      ]
    }],
    graphic: total > 0 ? [{
      type: 'text',
      left: 'center',
      top: '38%',
      style: {
        text: String(total),
        textAlign: 'center',
        fill: '#0f172a',
        fontSize: 26,
        fontWeight: 800
      }
    }, {
      type: 'text',
      left: 'center',
      top: '52%',
      style: {
        text: '总会员',
        textAlign: 'center',
        fill: '#94a3b8',
        fontSize: 12
      }
    }] : []
  }
}

function statusChartOption(data: { label: string; value: number; color: string }[]) {
  const labels = data.map(d => d.label)
  const values = data.map(d => d.value)
  const colors = data.map(d => d.color)
  return {
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: 10, right: 30, top: 10, bottom: 10, containLabel: true },
    xAxis: { type: 'value', show: false },
    yAxis: {
      type: 'category',
      data: labels,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { fontSize: 12, color: '#475569' }
    },
    series: [{
      type: 'bar',
      data: values.map((v, i) => ({
        value: v,
        itemStyle: {
          color: colors[i],
          borderRadius: [0, 6, 6, 0]
        }
      })),
      barWidth: 14,
      label: {
        show: true,
        position: 'right',
        fontSize: 13,
        fontWeight: 700,
        color: '#0f172a',
        formatter: (p: any) => p.value > 0 ? p.value : ''
      }
    }]
  }
}

function updateCharts() {
  nextTick(() => {
    typeChart?.setOption(typeChartOption(), true)
    statusChart?.setOption(statusChartOption([
      { label: '正式会员', value: stats.value[1].value, color: '#22c55e' },
      { label: '待审核', value: stats.value[3].value - stats.value[4].value > 0 ? stats.value[3].value - stats.value[4].value : stats.value[3].value, color: '#f59e0b' },
      { label: '待缴费', value: stats.value[4].value, color: '#f97316' },
      { label: '已拒绝', value: stats.value[5].value, color: '#ef4444' }
    ]), true)
  })
}

function handleResize() {
  typeChart?.resize()
  statusChart?.resize()
}

// ===== 快捷操作 =====
const quickActions = [
  { icon: 'DocumentChecked', label: '入会审核', desc: '审核新会员申请', path: '/admin/applications', color: '#002fa7', bg: '#ccd5ed' },
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
  { text: '待回复会员留言', tag: 'primary' as const, color: '#002fa7', count: 0, path: '/admin/messages' },
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
    typeData.unit = unitRes.data?.total || 0
    typeData.personal = personalRes.data?.total || 0

    // 3. 待办事项
    const pendingPayment = d.pending_payment || 0
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
    updateCharts()
  } catch (err) {
    console.error('Dashboard data fetch error:', err)
  } finally {
    loading.value = false
  }
}

function refresh() {
  fetchData()
}

onMounted(() => {
  nextTick(initCharts)
  fetchData()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  typeChart?.dispose()
  statusChart?.dispose()
})
</script>

<style scoped lang="scss">
.admin-dash {
  width: 100%;
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
  border: 1px solid #eef2f8;
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
    background: linear-gradient(90deg, var(--card-color, #002fa7), var(--card-color-light, #8097d3));
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
  display: flex;
  flex-wrap: wrap;

  > .el-col {
    display: flex;

    > .el-card {
      flex: 1;
    }
  }
}

.section-card {
  border-radius: 14px;
  border: 1px solid #eef2f8;
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;

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
    flex: 1;
  }

  :deep(.el-card__header) {
    flex-shrink: 0;
  }
}

// ===== ECharts Container =====
.chart-container {
  width: 100%;
  height: 220px;
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
  .chart-container {
    height: 180px;
  }
}
</style>
