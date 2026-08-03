<template>
  <div class="dashboard">
    <el-row :gutter="16">
      <el-col :span="6"><el-card><div class="stat"><div class="stat-num">{{ stats.upcoming_meetings || 0 }}</div><div class="stat-label">可报名会议</div></div></el-card></el-col>
      <el-col :span="6"><el-card><div class="stat"><div class="stat-num">{{ stats.my_registrations || 0 }}</div><div class="stat-label">我的报名</div></div></el-card></el-col>
      <el-col :span="6"><el-card><div class="stat"><div class="stat-num">{{ stats.total_credits || 0 }}</div><div class="stat-label">总学分</div></div></el-card></el-col>
      <el-col :span="6"><el-card><div class="stat"><div class="stat-num">{{ stats.unread_notifications || 0 }}</div><div class="stat-label">未读通知</div></div></el-card></el-col>
    </el-row>
    <el-row :gutter="16" style="margin-top:20px">
      <el-col :span="24"><el-card><h3>快捷操作</h3><div class="quick-actions"><el-button type="primary" @click="$router.push('/member/meetings')">查看会议</el-button><el-button @click="$router.push('/member/registrations')">我的报名</el-button><el-button @click="$router.push('/member/credits')">学分中心</el-button><el-button @click="$router.push('/member/notifications')">通知中心</el-button></div></el-card></el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { memberApi } from '@/api/member'

const stats = ref<any>({})

onMounted(async () => {
  try { const res = await memberApi.getDashboard(); stats.value = res.data } catch (e) {}
})
</script>

<style scoped>
.stat { text-align: center; padding: 10px 0; }
.stat-num { font-size: 32px; font-weight: bold; color: #2563eb; }
.stat-label { font-size: 14px; color: #666; margin-top: 4px; }
.quick-actions { display: flex; gap: 12px; flex-wrap: wrap; margin-top: 12px; }
</style>
