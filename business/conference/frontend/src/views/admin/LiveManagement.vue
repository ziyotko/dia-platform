<template>
  <div>
    <el-card>
      <el-form :inline="true"><el-form-item><el-select v-model="meetingId" placeholder="选择会议" @change="load"><el-option v-for="m in meetings" :key="m.id" :label="m.title" :value="m.id" /></el-select></el-form-item></el-form>
      <div v-if="config.id">
        <p>推流地址：{{ config.pushUrl }}</p>
        <p>播放地址：{{ config.playUrl }}</p>
        <p>状态：<el-tag :type="config.status==='live'?'success':'info'">{{ config.status==='live'?'直播中':'已结束' }}</el-tag></p>
        <div style="margin-top:12px"><el-button type="success" @click="start" v-if="config.status!=='live'">开启直播</el-button><el-button type="danger" @click="stop" v-if="config.status==='live'">结束直播</el-button></div>
      </div>
    </el-card>

    <el-card style="margin-top:16px"><h4>上传录播</h4><el-form :inline="true"><el-input v-model="vodUrl" placeholder="录播地址" /><el-button @click="uploadVod">上传</el-button></el-form></el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'

const meetings = ref<any[]>([]); const meetingId = ref(0); const config = ref<any>({}); const vodUrl = ref('')

onMounted(async () => {
  try { const res = await adminApi.getMeetings({ pageSize: 1000 }); meetings.value = res.data?.list || [] } catch (e) {}
})

async function load() {
  try { const res = await adminApi.getLiveConfig(meetingId.value); config.value = res.data || {} } catch (e) { config.value = {} }
}

async function start() { try { await adminApi.startLive(meetingId.value); ElMessage.success('直播已开启'); load() } catch (e: any) {} }
async function stop() { try { await adminApi.stopLive(meetingId.value); ElMessage.success('直播已结束'); load() } catch (e: any) {} }
async function uploadVod() { try { await adminApi.uploadVod(meetingId.value, { videoUrl: vodUrl.value }); ElMessage.success('上传成功') } catch (e: any) {} }
</script>
