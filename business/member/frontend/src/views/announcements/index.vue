<template>
  <div class="announce-page">
    <div class="page-header">
      <h2>公告动态</h2>
      <div class="search-bar">
        <el-input v-model="keyword" placeholder="搜索公告" clearable style="width:240px" @clear="search" @keyup.enter="search" />
        <el-button type="primary" @click="search">搜索</el-button>
      </div>
    </div>
    <el-card>
      <div v-loading="loading">
        <div class="announce-item" v-for="a in list" :key="a.id" @click="$router.push(`/announcements/${a.id}`)">
          <div class="item-left">
            <el-tag v-if="a.is_pinned" size="small" type="danger">置顶</el-tag>
            <el-tag v-else-if="typeLabel(a.type)" size="small" :type="typeTag(a.type)">
              {{ typeLabel(a.type) }}
            </el-tag>
            <span class="title">{{ a.title }}</span>
          </div>
          <span class="time">{{ formatDate(a.published_at) }}</span>
        </div>
        <el-empty v-if="!loading && list.length === 0" description="暂无公告" />
      </div>
      <div class="pagination" v-if="total > size">
        <el-pagination background layout="prev, pager, next" :total="total" :page-size="size" v-model:current-page="page" @change="fetchData" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { announcementApi } from '@/api/index'

const list = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const size = ref(10)
const total = ref(0)
const keyword = ref('')

onMounted(() => fetchData())

async function fetchData() {
  loading.value = true
  try {
    const res = await announcementApi.getPublished({ page: page.value, size: size.value, keyword: keyword.value })
    list.value = res.data.list || []
    total.value = res.data.total || 0
  } catch {} finally { loading.value = false }
}

function search() { page.value = 1; fetchData() }
function formatDate(d: string) { return d ? d.slice(0, 10) : '' }

const typeMap: Record<string, string> = { notice: '通知', article: '文章', policy: '政策' }
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
.announce-page {
  max-width: 900px;
  margin: 0 auto;
  padding: 40px 24px;
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
    h2 { font-size: 24px; }
    .search-bar { display: flex; gap: 12px; }
  }
  .announce-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid #f3f4f6;
    cursor: pointer;
    &:hover { background: #f9fafb; }
    .item-left { display: flex; align-items: center; gap: 12px; }
    .title { font-size: 15px; color: #374151; }
    .time { color: #9ca3af; font-size: 13px; white-space: nowrap; }
  }
  .pagination { display: flex; justify-content: center; margin-top: 24px; }
}
</style>
