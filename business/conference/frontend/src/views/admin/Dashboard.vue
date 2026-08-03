<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="16">
      <el-col :span="4" v-for="s in statCards" :key="s.label">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" :style="{ background: s.color + '1a', color: s.color }">
            <el-icon :size="22"><component :is="s.icon" /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-num">{{ s.value }}</div>
            <div class="stat-label">{{ s.label }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 第一行图表 -->
    <el-row :gutter="16" style="margin-top:16px">
      <el-col :span="12"><el-card shadow="hover"><h4 class="chart-title">会议概况</h4><div style="height:300px" ref="chart1"></div></el-card></el-col>
      <el-col :span="12"><el-card shadow="hover"><h4 class="chart-title">财务概览</h4><div style="height:300px" ref="chart2"></div></el-card></el-col>
    </el-row>

    <!-- 第二行图表 -->
    <el-row :gutter="16" style="margin-top:16px">
      <el-col :span="12"><el-card shadow="hover"><h4 class="chart-title">报名与签到统计</h4><div style="height:300px" ref="chart3"></div></el-card></el-col>
      <el-col :span="12"><el-card shadow="hover"><h4 class="chart-title">学分发放统计</h4><div style="height:300px" ref="chart4"></div></el-card></el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import * as echarts from 'echarts'
import { adminApi } from '@/api/admin'

const statCards = ref<any[]>([])
const chart1 = ref(); const chart2 = ref(); const chart3 = ref(); const chart4 = ref()
let chartInstances: echarts.ECharts[] = []

onMounted(async () => {
  try {
    const res = await adminApi.getDashboard()
    const d = res.data
    statCards.value = [
      { label: '总会议数', value: d.total_meetings, icon: 'Calendar', color: '#2563eb' },
      { label: '进行中会议', value: d.open_meetings, icon: 'VideoPlay', color: '#10b981' },
      { label: '总报名数', value: d.total_registrations, icon: 'Tickets', color: '#f59e0b' },
      { label: '签到人数', value: d.total_sign_ins, icon: 'Checked', color: '#6366f1' },
      { label: '总收入(元)', value: d.total_income, icon: 'Money', color: '#ec4899' },
      { label: '总学分', value: d.total_credits, icon: 'Star', color: '#14b8a6' },
    ]
    await nextTick()
    initCharts(d)
    window.addEventListener('resize', handleResize)
  } catch (e) {}
})

function initCharts(d: any) {
  // 1. 会议概况 - 环形图
  const c1 = echarts.init(chart1.value)
  c1.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c}场 ({d}%)' },
    legend: { bottom: 0 },
    series: [{
      type: 'pie', radius: ['40%', '70%'], center: ['50%', '45%'],
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { show: false },
      data: [
        { name: '进行中', value: d.open_meetings, itemStyle: { color: '#10b981' } },
        { name: '已关闭', value: d.closed_meetings, itemStyle: { color: '#f59e0b' } },
        { name: '草稿', value: d.total_meetings - d.open_meetings - d.closed_meetings, itemStyle: { color: '#9ca3af' } },
      ]
    }]
  })
  chartInstances.push(c1)

  // 2. 财务概览 - 柱状图
  const c2 = echarts.init(chart2.value)
  c2.setOption({
    tooltip: { trigger: 'axis' },
    legend: { bottom: 0 },
    grid: { left: 50, right: 20, top: 30, bottom: 40 },
    xAxis: { type: 'category', data: ['收入', '退款', '净收入'] },
    yAxis: { type: 'value' },
    series: [{
      type: 'bar', barWidth: 40,
      itemStyle: { borderRadius: [6, 6, 0, 0] },
      data: [
        { value: d.total_income, itemStyle: { color: '#2563eb' } },
        { value: d.total_refunded, itemStyle: { color: '#ef4444' } },
        { value: d.net_income, itemStyle: { color: '#10b981' } },
      ]
    }]
  })
  chartInstances.push(c2)

  // 3. 报名与签到 - 柱状对比
  const c3 = echarts.init(chart3.value)
  c3.setOption({
    tooltip: { trigger: 'axis' },
    legend: { bottom: 0 },
    grid: { left: 50, right: 20, top: 30, bottom: 40 },
    xAxis: { type: 'category', data: ['报名', '已通过', '已签到'] },
    yAxis: { type: 'value' },
    series: [{
      type: 'bar', barWidth: 40,
      itemStyle: { borderRadius: [6, 6, 0, 0] },
      data: [
        { value: d.total_registrations, itemStyle: { color: '#6366f1' } },
        { value: d.approved_registrations, itemStyle: { color: '#2563eb' } },
        { value: d.total_sign_ins, itemStyle: { color: '#10b981' } },
      ]
    }]
  })
  chartInstances.push(c3)

  // 4. 学分发放 - 仪表盘
  const c4 = echarts.init(chart4.value)
  c4.setOption({
    tooltip: { formatter: '{a}: {c}' },
    series: [{
      name: '已发放学分', type: 'gauge', min: 0, max: Math.max(d.total_credits, 100),
      radius: '85%', center: ['50%', '55%'],
      progress: { show: true, width: 18, itemStyle: { color: '#14b8a6' } },
      axisLine: { lineStyle: { width: 18 } },
      axisTick: { show: false }, splitLine: { length: 12, lineStyle: { width: 2 } },
      axisLabel: { distance: 20, fontSize: 12 },
      pointer: { show: true },
      detail: { valueAnimation: true, fontSize: 24, offsetCenter: [0, '70%'], formatter: '{value}' },
      data: [{ value: d.total_credits, name: '总学分' }]
    }]
  })
  chartInstances.push(c4)
}

function handleResize() { chartInstances.forEach(c => c.resize()) }

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chartInstances.forEach(c => c.dispose())
  chartInstances = []
})
</script>

<style scoped>
.stat-card :deep(.el-card__body) { display: flex; align-items: center; gap: 12px; padding: 16px; }
.stat-icon { width: 48px; height: 48px; border-radius: 10px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.stat-num { font-size: 22px; font-weight: 700; color: #1f2937; }
.stat-label { font-size: 13px; color: #6b7280; margin-top: 2px; }
.chart-title { margin-bottom: 8px; color: #374151; }
</style>
