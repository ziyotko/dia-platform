<template>
  <div class="service-page" v-loading="loading">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="公告动态" name="announcements">
        <el-card>
          <div class="announce-item" v-for="a in announcements" :key="a.id" @click="$router.push(`/announcements/${a.id}`)">
            <div>
              <el-tag v-if="a.is_pinned" size="small" type="danger">置顶</el-tag>
              <el-tag v-else-if="typeLabel(a.type)" size="small" :type="typeTag(a.type)">
                {{ typeLabel(a.type) }}
              </el-tag>
              <span class="title">{{ a.title }}</span>
            </div>
            <span class="time">{{ formatDate(a.published_at) }}</span>
          </div>
          <el-empty v-if="announcements.length === 0" description="暂无公告" />
          <div class="pagination" v-if="annTotal > 0">
            <el-pagination background layout="total, prev, pager, next" :total="annTotal" :page-size="annSize" v-model:current-page="annPage" @change="fetchAnnouncements" />
          </div>
        </el-card>
      </el-tab-pane>
      <el-tab-pane label="会员文章" name="articles">
        <el-card>
          <div class="announce-item" v-for="a in pubArticles" :key="a.id" @click="openArticle(a)">
            <div>
              <el-tag size="small">{{ a.category?.name }}</el-tag>
              <span class="title">{{ a.title }}</span>
            </div>
            <span class="time">{{ a.member?.company_name || a.member?.username }}</span>
          </div>
          <el-empty v-if="pubArticles.length === 0" description="暂无文章" />
          <div class="pagination" v-if="artTotal > 0">
            <el-pagination background layout="total, prev, pager, next" :total="artTotal" :page-size="artSize" v-model:current-page="artPage" @change="fetchArticles" />
          </div>
        </el-card>
      </el-tab-pane>
      <el-tab-pane label="协会章程" name="charter">
        <el-card v-loading="charterLoading">
          <div v-if="charterContent" class="rich-text-content" v-html="sanitizeHtml(charterContent)"></div>
          <el-empty v-else-if="!charterLoading" description="章程暂未维护" />
          <div class="charter-actions" v-if="charterFile || charterContent">
            <el-button v-if="charterFile" type="primary" plain size="small" @click="downloadCharterPdf">
              <el-icon><Download /></el-icon> 下载章程 PDF
            </el-button>
            <el-button size="small" @click="$router.push('/charter')">在新页面打开</el-button>
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 文章详情 -->
    <el-dialog v-model="showArticle" :title="viewedArticle?.title || '文章详情'" width="720px" top="6vh">
      <div class="article-meta" v-if="viewedArticle">
        <el-tag size="small">{{ viewedArticle.category?.name }}</el-tag>
        <span>作者：{{ viewedArticle.member?.company_name || viewedArticle.member?.name || viewedArticle.member?.username }}</span>
        <span>时间：{{ formatDate(viewedArticle.published_at) }}</span>
      </div>
      <div class="article-content" v-html="sanitizeHtml(viewedArticle?.content) || '暂无内容'"></div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { announcementApi, articleApi, charterApi } from '@/api/index'
import { sanitizeHtml } from '@/utils/sanitizeHtml'

const activeTab = ref('announcements')
const announcements = ref<any[]>([])
const pubArticles = ref<any[]>([])
const loading = ref(true)

// 文章详情
const showArticle = ref(false)
const viewedArticle = ref<any>(null)

// 公告分页
const annPage = ref(1)
const annSize = ref(10)
const annTotal = ref(0)

// 文章分页
const artPage = ref(1)
const artSize = ref(10)
const artTotal = ref(0)

onMounted(() => {
  fetchAnnouncements()
  fetchArticles()
})

// 切 tab 时刷新对应数据
watch(activeTab, () => {
  if (activeTab.value === 'announcements' && announcements.value.length === 0) fetchAnnouncements()
  if (activeTab.value === 'articles' && pubArticles.value.length === 0) fetchArticles()
  if (activeTab.value === 'charter' && !charterLoaded.value) fetchCharter()
})

/* ---- 协会章程（后台用富文本编辑器维护） ---- */
const charterContent = ref('')
const charterFile = ref<any>(null)
const charterLoading = ref(false)
const charterLoaded = ref(false)

async function fetchCharter() {
  charterLoading.value = true
  try {
    const res = await charterApi.getContent()
    charterContent.value = res.data?.content || ''
    charterFile.value = res.data?.file || null
    charterLoaded.value = true
  } catch {} finally { charterLoading.value = false }
}

function downloadCharterPdf() {
  const base = import.meta.env.VITE_API_BASE_URL || '/business_member/api'
  window.open(`${base}/charter`, '_blank')
}

async function fetchAnnouncements() {
  loading.value = true
  try {
    const res = await announcementApi.getPublished({ page: annPage.value, size: annSize.value })
    announcements.value = res.data?.list || []
    annTotal.value = res.data?.total || 0
  } catch {} finally { loading.value = false }
}

async function fetchArticles() {
  loading.value = true
  try {
    const res = await articleApi.listPublished({ page: artPage.value, size: artSize.value })
    pubArticles.value = res.data?.list || []
    artTotal.value = res.data?.total || 0
  } catch {} finally { loading.value = false }
}

async function openArticle(row: any) {
  try {
    const res = await articleApi.getPublishedArticle(row.id)
    viewedArticle.value = res.data || row
    showArticle.value = true
  } catch {}
}

function formatDate(d: string) { return d ? d.slice(0, 10) : '' }

const typeMap: Record<string, string> = { notice: '公告', article: '文章', policy: '政策' }
// el-tag 的 type 不接受空串（会触发 prop 校验警告），未知类型统一回退 info
const typeTagMap: Record<string, 'primary' | 'success' | 'warning' | 'info'> = {
  notice: 'primary',
  article: 'success',
  policy: 'warning'
}
function typeLabel(t?: string) { return t ? typeMap[t] || t : '' }
function typeTag(t?: string) { return t ? typeTagMap[t] || 'info' : 'info' }
</script>

<style scoped lang="scss">
.service-page { width: 100%;}
.pagination { display: flex; justify-content: center; margin-top: 20px; }
.announce-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: 14px 0; border-bottom: 1px solid #f3f4f6; cursor: pointer;
  &:hover { background: #f9fafb; }
  .title { margin-left: 10px; font-size: 14px; }
  .time { color: #9ca3af; font-size: 13px; white-space: nowrap; }
}
.article-meta { display: flex; gap: 16px; align-items: center; margin-bottom: 16px; font-size: 13px; color: #6b7280; }
.article-content { line-height: 1.8; }
/* 协会章程标签页：正文用全局 .rich-text-content，这里只加操作区 */
.charter-actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--app-border);
}
</style>
