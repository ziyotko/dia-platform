<template>
  <div class="detail-page" v-loading="loading">
    <div class="detail-card" v-if="announcement">
      <h1>{{ announcement.title }}</h1>
      <div class="meta">
        <span>{{ formatDate(announcement.published_at) }}</span>
        <span>浏览 {{ announcement.view_count }}</span>
      </div>
      <div class="content" v-html="announcement.content.replace(/\n/g, '<br>')"></div>
    </div>
    <el-empty v-else-if="!loading" description="公告不存在" />
    <div class="back">
      <el-button @click="$router.back()">返回列表</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { announcementApi } from '@/api/index'

const route = useRoute()
const announcement = ref<any>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    const id = Number(route.params.id)
    const res = await announcementApi.getDetail(id)
    announcement.value = res.data
  } catch {} finally { loading.value = false }
})

function formatDate(d: string) { return d ? d.slice(0, 10) : '' }
</script>

<style scoped lang="scss">
.detail-page {
  max-width: 860px;
  margin: 40px auto;
  padding: 0 24px;
  .detail-card {
    background: #fff;
    border-radius: 12px;
    padding: 40px;
    box-shadow: 0 1px 3px rgba(0,0,0,0.06);
    h1 { font-size: 24px; margin-bottom: 16px; }
    .meta { color: #9ca3af; font-size: 13px; margin-bottom: 32px; display: flex; gap: 16px; }
    .content { line-height: 2; font-size: 15px; color: #374151; }
  }
  .back { margin-top: 24px; }
}
</style>
