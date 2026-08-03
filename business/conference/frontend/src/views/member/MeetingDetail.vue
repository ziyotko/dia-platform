<template>
  <div>
    <el-card>
      <div v-if="meeting.id">
        <el-tag :type="meeting.type==='online'?'success':meeting.type==='offline'?'':'warning'">{{ meeting.type==='online'?'线上会议':meeting.type==='offline'?'线下会议':'混合会议' }}</el-tag>
        <h1 style="margin:12px 0">{{ meeting.title }}</h1>
        <p><strong>时间：</strong>{{ formatTime(meeting.startTime) }} ~ {{ formatTime(meeting.endTime) }}</p>
        <p v-if="meeting.location"><strong>地点：</strong>{{ meeting.location }}</p>
        <p><strong>名额：</strong>{{ meeting.capacity || '不限' }}人 | <strong>费用：</strong>{{ meeting.fee > 0 ? `¥${meeting.fee}` : '免费' }}</p>

        <div style="margin-top:20px;display:flex;gap:12px;flex-wrap:wrap">
          <el-button type="primary" @click="handleRegister" v-if="!registered && meeting.status==='open'">立即报名</el-button>
          <el-button v-if="registered" :type="regStatus==='approved'?'success':regStatus==='pending'?'warning':'info'" disabled>
            {{ regStatus==='approved'?'已通过':regStatus==='pending'?'待审核':regStatus==='waitlist'?'候补中':regStatus==='rejected'?'已驳回':'已取消' }}
          </el-button>
          <el-button @click="cancelReg" v-if="registered && (regStatus==='pending'||regStatus==='waitlist')">取消报名</el-button>
          <el-button type="success" @click="$router.push(`/member/live/${meeting.id}`)" v-if="registered && regStatus==='approved'">进入直播</el-button>
        </div>
      </div>
    </el-card>

    <el-card style="margin-top:16px"><h3>会议议程</h3>
      <el-timeline v-if="meeting.agendas?.length">
        <el-timeline-item v-for="agenda in meeting.agendas" :key="agenda.id" :timestamp="formatTime(agenda.startTime)">
          <strong>{{ agenda.title }}</strong><br/><span>{{ agenda.speaker }}</span>
        </el-timeline-item>
      </el-timeline>
      <el-empty v-else description="暂无议程" />
    </el-card>

    <el-card style="margin-top:16px"><h3>嘉宾</h3>
      <el-row :gutter="16"><el-col :span="6" v-for="guest in meeting.guests" :key="guest.id" style="margin-bottom:12px;text-align:center">
        <el-avatar :size="60">{{ guest.name?.charAt(0) }}</el-avatar>
        <p><strong>{{ guest.name }}</strong></p>
        <p style="color:#666;font-size:13px">{{ guest.title }}</p>
      </el-col></el-row>
      <el-empty v-if="!meeting.guests?.length" description="暂无嘉宾信息" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { memberApi } from '@/api/member'

const route = useRoute()
const meeting = ref<any>({})
const registered = ref(false)
const regStatus = ref('')
const regId = ref(0)

function formatTime(t: string) { return (t || '').replace('T', ' ').slice(0, 16) }

onMounted(async () => {
  const id = Number(route.params.id)
  try {
    const res = await memberApi.getMeetingDetail(id)
    meeting.value = res.data
    const regRes = await memberApi.getRegistrationStatus(id)
    if (regRes.data?.status) {
      registered.value = true
      regStatus.value = regRes.data.status
      regId.value = regRes.data.id
    }
  } catch (e) {}
})

async function handleRegister() {
  try {
    await memberApi.register(meeting.value.id)
    ElMessage.success('报名成功')
    registered.value = true
  } catch (e: any) {}
}

async function cancelReg() {
  try {
    await memberApi.cancelRegistration(regId.value)
    ElMessage.success('已取消')
    registered.value = false
  } catch (e: any) {}
}
</script>
