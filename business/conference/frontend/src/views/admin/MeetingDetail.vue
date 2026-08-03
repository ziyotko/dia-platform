<template>
  <div>
    <el-card v-if="meeting.id">
      <h2>{{ meeting.title }}</h2>
      <p>类型：{{ meeting.type }} | 名额：{{ meeting.capacity || '不限' }} | 费用：¥{{ meeting.fee }}</p>
      <p>时间：{{ meeting.startTime?.slice(0,16) }} ~ {{ meeting.endTime?.slice(0,16) }}</p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { adminApi } from '@/api/admin'

const route = useRoute()
const meeting = ref<any>({})

onMounted(async () => {
  try { const res = await adminApi.getMeeting(Number(route.params.id)); meeting.value = res.data } catch (e) {}
})
</script>
