<template>
  <div>
    <el-row :gutter="16">
      <el-col :span="4" v-for="s in statCards" :key="s.label"><el-card><div style="text-align:center"><div style="font-size:28px;font-weight:bold;color:#2563eb">{{ s.value }}</div><div style="color:#666;font-size:13px">{{ s.label }}</div></div></el-card></el-col>
    </el-row>
    <el-row :gutter="16" style="margin-top:16px">
      <el-col :span="12"><el-card><h4>会议概况</h4><div style="height:250px" ref="chart1"></div></el-card></el-col>
      <el-col :span="12"><el-card><h4>财务概览</h4><div style="height:250px" ref="chart2"></div></el-card></el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import { adminApi } from '@/api/admin'

const statCards = ref<any[]>([])
const chart1 = ref(); const chart2 = ref()

onMounted(async () => {
  try {
    const res = await adminApi.getDashboard()
    const d = res.data
    statCards.value = [
      { label: '总会议数', value: d.total_meetings },
      { label: '进行中会议', value: d.open_meetings },
      { label: '总报名数', value: d.total_registrations },
      { label: '签到人数', value: d.total_sign_ins },
      { label: '总收入(元)', value: d.total_income },
      { label: '总学分', value: d.total_credits },
    ]
    await nextTick()
    initChart1(d); initChart2(d)
  } catch (e) {}
})

function initChart1(d: any) {
  const c = echarts.init(chart1.value)
  c.setOption({
    tooltip: {}, series: [{ type: 'pie', radius: ['40%','70%'], data: [{ name:'进行中',value:d.open_meetings},{ name:'已关闭',value:d.closed_meetings}] }]
  })
}

function initChart2(d: any) {
  const c = echarts.init(chart2.value)
  c.setOption({
    tooltip: {}, xAxis: { data: ['收入','退款'] }, yAxis: {}, series: [{ type: 'bar', data: [d.total_income || 0, d.total_refunded || 0], itemStyle: { color: '#2563eb' } }]
  })
}
</script>
