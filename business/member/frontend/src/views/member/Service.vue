<template>
  <div class="service-page" v-loading="loading">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="公告动态" name="announcements">
        <el-card>
          <div class="announce-item" v-for="a in announcements" :key="a.id" @click="$router.push(`/announcements/${a.id}`)">
            <div>
              <el-tag size="small" :type="a.is_pinned ? 'danger' : ''">{{ a.is_pinned ? '置顶' : a.type }}</el-tag>
              <span class="title">{{ a.title }}</span>
            </div>
            <span class="time">{{ formatDate(a.published_at) }}</span>
          </div>
          <el-empty v-if="announcements.length === 0" description="暂无公告" />
        </el-card>
      </el-tab-pane>
      <el-tab-pane label="会员文章" name="articles">
        <el-card>
          <div class="announce-item" v-for="a in pubArticles" :key="a.id">
            <div>
              <el-tag size="small">{{ a.category?.name }}</el-tag>
              <span class="title">{{ a.title }}</span>
            </div>
            <span class="time">{{ a.member?.company_name || a.member?.username }}</span>
          </div>
          <el-empty v-if="pubArticles.length === 0" description="暂无文章" />
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { announcementApi, articleApi } from '@/api/index'

const activeTab = ref('announcements')
const announcements = ref<any[]>([])
const pubArticles = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const [annRes, artRes] = await Promise.all([
      announcementApi.getPublished({ page: 1, size: 50 }),
      articleApi.listPublished({ page: 1, size: 50 })
    ])
    announcements.value = annRes.data?.list || []
    pubArticles.value = artRes.data?.list || []
  } catch {} finally { loading.value = false }
})

function formatDate(d: string) { return d ? d.slice(0, 10) : '' }
</script>

<style scoped lang="scss">
.service-page { max-width: 900px; margin: 0 auto; }
.announce-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: 14px 0; border-bottom: 1px solid #f3f4f6; cursor: pointer;
  &:hover { background: #f9fafb; }
  .title { margin-left: 10px; font-size: 14px; }
  .time { color: #9ca3af; font-size: 13px; white-space: nowrap; }
}
</style>
