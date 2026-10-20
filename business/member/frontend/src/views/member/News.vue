<template>
  <div class="news-page">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="公告动态" name="announcements">
        <el-card v-loading="annLoading">
          <div class="news-item" v-for="item in announcements" :key="item.id" @click="openAnnouncement(item)">
            <div>
              <el-tag v-if="item.is_pinned" size="small" type="danger">置顶</el-tag>
              <el-tag v-else size="small" type="info">公告</el-tag>
              <span class="title">{{ item.title }}</span>
            </div>
            <span class="time">{{ formatDate(item.published_at) }}</span>
          </div>
          <el-empty v-if="!annLoading && announcements.length === 0" description="暂无公告" />
          <div class="pagination" v-if="annTotal > 0">
            <el-pagination background layout="total, prev, pager, next" :total="annTotal" :page-size="pageSize" v-model:current-page="annPage" @change="fetchAnnouncements" />
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="会员文章" name="articles">
        <el-card v-loading="articleLoading">
          <div class="news-item" v-for="item in articles" :key="item.id" @click="openArticle(item)">
            <div>
              <el-tag size="small">{{ item.category?.name || '会员文章' }}</el-tag>
              <span class="title">{{ item.title }}</span>
            </div>
            <span class="time">{{ item.member?.company_name || item.member?.name || item.member?.username }}</span>
          </div>
          <el-empty v-if="!articleLoading && articles.length === 0" description="暂无文章" />
          <div class="pagination" v-if="articleTotal > 0">
            <el-pagination background layout="total, prev, pager, next" :total="articleTotal" :page-size="pageSize" v-model:current-page="articlePage" @change="fetchArticles" />
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="detailVisible" :title="detail?.title || '资讯详情'" width="720px" top="6vh">
      <div class="detail-meta" v-if="detail">
        <span v-if="detail.category?.name">{{ detail.category.name }}</span>
        <span v-if="detail.member">作者：{{ detail.member.company_name || detail.member.name || detail.member.username }}</span>
        <span>时间：{{ formatDate(detail.published_at) }}</span>
      </div>
      <div class="rich-text-content" v-html="sanitizeHtml(detail?.content) || '暂无内容'"></div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { announcementApi, articleApi } from '@/api'
import { sanitizeHtml } from '@/utils/sanitizeHtml'

const activeTab = ref('announcements')
const announcements = ref<any[]>([])
const articles = ref<any[]>([])
const annLoading = ref(false)
const articleLoading = ref(false)
const annPage = ref(1)
const articlePage = ref(1)
const pageSize = 10
const annTotal = ref(0)
const articleTotal = ref(0)
const detailVisible = ref(false)
const detail = ref<any>(null)

async function fetchAnnouncements() {
  annLoading.value = true
  try {
    const res = await announcementApi.getPublished({ page: annPage.value, pageSize })
    announcements.value = res.data?.list || []
    annTotal.value = res.data?.total || 0
  } finally { annLoading.value = false }
}

async function fetchArticles() {
  articleLoading.value = true
  try {
    const res = await articleApi.listPublished({ page: articlePage.value, pageSize })
    articles.value = res.data?.list || []
    articleTotal.value = res.data?.total || 0
  } finally { articleLoading.value = false }
}

async function openAnnouncement(item: any) {
  const res = await announcementApi.getDetail(item.id)
  detail.value = res.data || item
  detailVisible.value = true
}

async function openArticle(item: any) {
  const res = await articleApi.getPublishedArticle(item.id)
  detail.value = res.data || item
  detailVisible.value = true
}

function formatDate(value?: string) { return value ? value.slice(0, 10) : '' }
onMounted(() => { fetchAnnouncements(); fetchArticles() })
</script>

<style scoped lang="scss">
.news-page { width: 100%; }
.news-item {
  display: flex; justify-content: space-between; align-items: center; gap: 16px;
  padding: 14px 4px; border-bottom: 1px solid #f1f5f9; cursor: pointer;
  &:hover { background: #f8fafc; }
  .title { margin-left: 10px; }
  .time { color: #94a3b8; font-size: 13px; white-space: nowrap; }
}
.pagination { display: flex; justify-content: center; margin-top: 20px; }
.detail-meta { display: flex; gap: 16px; margin-bottom: 16px; color: #64748b; font-size: 13px; }
</style>
