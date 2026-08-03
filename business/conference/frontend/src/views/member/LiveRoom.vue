<template>
  <div style="height:calc(100vh - 140px);display:flex;flex-direction:column">
    <el-card class="player-card">
      <div class="player-area">
        <video controls autoplay style="width:100%;max-height:500px;background:#000" :src="playUrl" v-if="playUrl"></video>
        <el-empty v-else description="直播未开始或已结束" />
      </div>
    </el-card>
    <el-card style="flex:1;margin-top:12px;overflow-y:auto">
      <h4>留言互动</h4>
      <div v-for="msg in messages" :key="msg.id" style="padding:4px 0"><strong>{{ msg.user?.realName || '用户' }}</strong>: {{ msg.content }}</div>
      <el-input v-model="newMsg" placeholder="输入留言..." style="margin-top:8px" @keyup.enter="sendMsg">
        <template #append><el-button @click="sendMsg">发送</el-button></template>
      </el-input>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { memberApi } from '@/api/member'

const route = useRoute()
const playUrl = ref('')
const messages = ref<any[]>([])
const newMsg = ref('')

onMounted(async () => {
  const id = Number(route.params.meetingId)
  try {
    const res = await memberApi.getLiveUrl(id)
    playUrl.value = res.data?.playUrl || ''
    await memberApi.startViewing(id, 'live')
    const msgs = await memberApi.getLiveMessages(id)
    messages.value = msgs.data || []
  } catch (e) {}
})

async function sendMsg() {
  if (!newMsg.value.trim()) return
  try {
    await memberApi.sendLiveMessage(Number(route.params.meetingId), { content: newMsg.value })
    newMsg.value = ''
  } catch (e) {}
}
</script>

<style scoped>
.player-card { padding: 0; }
.player-area { background: #000; border-radius: 4px; overflow: hidden; }
</style>
