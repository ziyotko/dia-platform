<template>
  <div class="home-page">
    <section class="hero">
      <div class="hero-content">
        <h1>{{ siteStore.site_name }}</h1>
        <p class="subtitle">会员服务系统</p>
        <p class="desc">数字化会员服务 · 数字化管理 · 数字化增值服务</p>
        <div class="hero-actions">
          <el-button type="primary" size="large" @click="$router.push('/register')">入会申请</el-button>
          <el-button size="large" @click="$router.push('/login')">会员登录</el-button>
        </div>
      </div>
    </section>

    <section class="features">
      <div class="section-title">
        <h2>会员服务</h2>
        <p>为您提供一站式数字化会员服务体验</p>
      </div>
      <div class="feature-grid">
        <div class="feature-card" v-for="f in features" :key="f.title">
          <el-icon :size="40" :color="f.color"><component :is="f.icon" /></el-icon>
          <h3>{{ f.title }}</h3>
          <p>{{ f.desc }}</p>
        </div>
      </div>
    </section>

    <section class="announcements" v-if="announcements.length">
      <div class="section-title">
        <h2>最新公告</h2>
      </div>
      <div class="announce-list">
        <div class="announce-item" v-for="a in announcements" :key="a.id" @click="$router.push(`/announcements/${a.id}`)">
          <el-tag size="small" :type="a.is_pinned ? 'danger' : 'info'">{{ a.is_pinned ? '置顶' : '公告' }}</el-tag>
          <span class="title">{{ a.title }}</span>
          <span class="time">{{ formatDate(a.published_at) }}</span>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { announcementApi } from '@/api/index'
import { useSiteStore } from '@/stores/site'

const siteStore = useSiteStore()

const features = [
  { icon: 'DocumentChecked', title: '在线入会', desc: '在线填写资料、提交申请、追踪审核进度', color: '#1a6fb5' },
  { icon: 'Money', title: '会费管理', desc: '查看缴费记录、在线缴纳年度会费', color: '#22c55e' },
  { icon: 'Medal', title: '证书下载', desc: '审核通过后在线下载会员证书', color: '#f59e0b' },
  { icon: 'Reading', title: '服务中心', desc: '浏览协会公告动态与行业文章', color: '#8b5cf6' },
  { icon: 'ChatDotRound', title: '在线留言', desc: '向协会提交建议、诉求与反馈', color: '#ef4444' },
  { icon: 'EditPen', title: '文章发布', desc: '发布行业动态、技术交流文章', color: '#06b6d4' }
]

const announcements = ref<any[]>([])

onMounted(async () => {
  try {
    const res = await announcementApi.getPublished({ page: 1, size: 5 })
    announcements.value = res.data.list || []
  } catch {}
})

function formatDate(d: string) {
  if (!d) return ''
  return d.slice(0, 10)
}
</script>

<style scoped lang="scss">
.home-page {
  .hero {
    background: linear-gradient(135deg, #0d4f85 0%, #1a6fb5 50%, #4a9fd5 100%);
    color: #fff;
    text-align: center;
    padding: 100px 24px;
    h1 { font-size: 42px; font-weight: 700; margin-bottom: 8px; }
    .subtitle { font-size: 24px; opacity: 0.9; margin-bottom: 16px; }
    .desc { font-size: 16px; opacity: 0.75; margin-bottom: 40px; }
    .hero-actions { display: flex; gap: 16px; justify-content: center; }
  }
  .section-title {
    text-align: center;
    padding: 60px 0 40px;
    h2 { font-size: 28px; font-weight: 700; color: #1f2937; }
    p { color: #6b7280; margin-top: 8px; }
  }
  .feature-grid {
    max-width: 1100px;
    margin: 0 auto;
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 24px;
    padding: 0 24px 60px;
  }
  .feature-card {
    background: #fff;
    border-radius: 12px;
    padding: 32px 24px;
    text-align: center;
    box-shadow: 0 1px 3px rgba(0,0,0,0.06);
    transition: transform 0.2s, box-shadow 0.2s;
    &:hover { transform: translateY(-4px); box-shadow: 0 8px 25px rgba(0,0,0,0.1); }
    h3 { margin: 16px 0 8px; font-size: 18px; }
    p { color: #6b7280; font-size: 14px; line-height: 1.6; }
  }
  .announcements {
    background: #fff;
    padding: 0 24px 60px;
  }
  .announce-list {
    max-width: 900px;
    margin: 0 auto;
  }
  .announce-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px;
    border-bottom: 1px solid #f3f4f6;
    cursor: pointer;
    &:hover { background: #f9fafb; }
    .title { flex: 1; font-size: 15px; color: #374151; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .time { color: #9ca3af; font-size: 13px; white-space: nowrap; }
  }
}
</style>
