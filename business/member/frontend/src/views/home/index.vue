<template>
  <div class="home-page">
    <!-- ═══════════════ Hero ═══════════════ -->
    <section class="hero">
      <div class="hero-decor" aria-hidden="true">
        <span class="blob blob-1"></span>
        <span class="blob blob-2"></span>
        <span class="blob blob-3"></span>
        <div class="hero-grid"></div>
        <div class="float-chip chip-1"><el-icon><DocumentChecked /></el-icon></div>
        <div class="float-chip chip-2"><el-icon><Coin /></el-icon></div>
        <div class="float-chip chip-3"><el-icon><Medal /></el-icon></div>
        <div class="float-chip chip-4"><el-icon><ChatDotRound /></el-icon></div>
      </div>
      <div class="container hero-content">
        <div class="hero-badge">
          <el-icon><Promotion /></el-icon>
          <span>数字化会员服务平台</span>
        </div>
        <h1>{{ siteStore.site_name }}</h1>
        <p class="hero-subtitle">数字化会员服务 · 数字化管理 · 数字化增值服务</p>
        <p class="hero-desc">
          面向行业单位的专业会员服务门户，提供在线入会、会费管理、证书下载与信息共享等一站式服务，
          助力会员单位数字化发展。
        </p>
        <div class="hero-actions">
          <el-button class="hero-btn-primary" size="large" @click="$router.push('/register')">
            立即申请入会 <el-icon class="el-icon--right"><ArrowRight /></el-icon>
          </el-button>
          <el-button class="hero-btn-ghost" size="large" @click="$router.push('/login')">会员登录</el-button>
        </div>
        <div class="hero-features">
          <div class="hf-item" v-for="f in heroFeatures" :key="f"><el-icon><Check /></el-icon>{{ f }}</div>
        </div>
      </div>
      <div class="hero-wave" aria-hidden="true">
        <svg viewBox="0 0 1440 90" preserveAspectRatio="none">
          <path d="M0,60 C240,95 480,25 720,40 C960,55 1200,90 1440,55 L1440,90 L0,90 Z" fill="#f5f7fa"></path>
        </svg>
      </div>
    </section>

    <!-- ═══════════════ 会员服务 ═══════════════ -->
    <section class="services" id="services">
      <div class="container">
        <div class="section-head">
          <span class="section-eyebrow">SERVICES</span>
          <h2>会员服务</h2>
          <p>为您提供一站式数字化会员服务体验</p>
        </div>
        <div class="service-grid">
          <div class="service-card" v-for="f in features" :key="f.title" @click="goService(f)">
            <div class="service-icon" :style="{ background: f.bg, color: f.color }">
              <el-icon :size="26"><component :is="f.icon" /></el-icon>
            </div>
            <h3>{{ f.title }}</h3>
            <p>{{ f.desc }}</p>
            <span class="service-more">了解更多 <el-icon><ArrowRight /></el-icon></span>
          </div>
        </div>
      </div>
    </section>

    <!-- ═══════════════ 公告 + 会员等级 ═══════════════ -->
    <section class="news-section">
      <div class="container">
        <div class="news-grid">
          <!-- 公告动态 -->
          <div class="news-card announcements-card">
            <div class="card-head">
              <h3><el-icon><Bell /></el-icon>公告动态</h3>
              <el-button text type="primary" size="small" @click="$router.push('/announcements')">
                查看全部 <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
            <div class="type-tabs">
              <span
                v-for="t in tabs" :key="t.value"
                :class="['tab', { active: activeTab === t.value }]"
                @click="activeTab = t.value"
              >{{ t.label }}</span>
            </div>
            <div v-loading="loadingAnn" class="announce-list">
              <div
                class="announce-item"
                v-for="a in filteredAnnouncements" :key="a.id"
                @click="$router.push(`/announcements/${a.id}`)"
              >
                <el-tag v-if="a.is_pinned" size="small" type="danger" effect="dark" class="pin-tag">置顶</el-tag>
                <el-tag v-else size="small" :type="typeTag(a.type)" effect="plain">{{ typeLabel(a.type) }}</el-tag>
                <span class="announce-title">{{ a.title }}</span>
                <span class="announce-meta">
                  <span class="views"><el-icon><View /></el-icon>{{ a.view_count || 0 }}</span>
                  <span class="date">{{ formatDate(a.published_at) }}</span>
                </span>
              </div>
              <el-empty v-if="!loadingAnn && filteredAnnouncements.length === 0" description="暂无公告" :image-size="70" />
            </div>
          </div>
          <!-- 会员等级 -->
          <div class="news-card levels-card">
            <div class="card-head">
              <h3><el-icon><Medal /></el-icon>会员等级</h3>
            </div>
            <div v-loading="loadingLevels" class="level-list">
              <div class="level-item" v-for="(lv, i) in levels" :key="lv.id">
                <div class="level-rank" :style="{ background: rankColors[i % rankColors.length] }">{{ lv.level + 1 }}</div>
                <div class="level-info">
                  <div class="level-name">{{ lv.name }}</div>
                  <div class="level-desc">{{ lv.description || '加入协会即可享受相应等级的会员权益与服务' }}</div>
                </div>
              </div>
              <el-empty v-if="!loadingLevels && levels.length === 0" description="暂无等级配置" :image-size="70" />
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ═══════════════ 入会流程 ═══════════════ -->
    <section class="process-section">
      <div class="container">
        <div class="section-head">
          <span class="section-eyebrow">HOW TO JOIN</span>
          <h2>入会流程</h2>
          <p>四步轻松完成入会，全程在线跟踪办理进度</p>
        </div>
        <div class="process-steps">
          <div class="step" v-for="(s, i) in processSteps" :key="s.title">
            <div class="step-node">
              <div class="step-circle"><el-icon :size="24"><component :is="s.icon" /></el-icon></div>
              <span v-if="i < processSteps.length - 1" class="step-line"></span>
            </div>
            <h4>{{ s.title }}</h4>
            <p>{{ s.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- ═══════════════ 分支机构 ═══════════════ -->
    <section class="org-section" v-if="divisions.length">
      <div class="container">
        <div class="section-head">
          <span class="section-eyebrow">BRANCHES</span>
          <h2>分支机构</h2>
          <p>覆盖行业各细分领域的专业分支机构</p>
        </div>
        <div class="org-grid">
          <div class="org-card" v-for="o in divisions" :key="o.id">
            <div class="org-head">
              <div class="org-icon"><el-icon :size="22"><OfficeBuilding /></el-icon></div>
              <div class="org-name">{{ o.name }}</div>
            </div>
            <p class="org-desc">{{ o.description || '暂无介绍' }}</p>
            <div class="org-foot">
              <span class="org-type">{{ orgTypeLabel(o.type) }}</span>
              <span v-if="o.contact_info" class="org-contact" :title="o.contact_info">
                <el-icon><Phone /></el-icon>{{ o.contact_info }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ═══════════════ CTA ═══════════════ -->
    <section class="cta-section">
      <div class="container">
        <div class="cta-banner">
          <div class="cta-decor" aria-hidden="true">
            <span class="cta-blob cta-blob-1"></span>
            <span class="cta-blob cta-blob-2"></span>
          </div>
          <div class="cta-text">
            <h3>加入我们，共享行业资源与发展机遇</h3>
            <p>如有入会意向或疑问，欢迎随时与我们联系</p>
          </div>
          <div class="cta-actions">
            <el-button class="cta-btn" size="large" @click="$router.push('/register')">立即申请入会</el-button>
            <el-button class="cta-btn-ghost" size="large" @click="$router.push('/login')">已有账号，登录</el-button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { announcementApi, orgApi } from '@/api/index'
import { useSiteStore } from '@/stores/site'

const router = useRouter()
const siteStore = useSiteStore()

// ── Hero ──
const heroFeatures = ['在线入会', '会费管理', '证书下载', '信息共享']

// ── Services ──
const features = [
  { icon: 'DocumentChecked', title: '在线入会', desc: '在线填写资料、提交申请、追踪审核进度', color: '#1a6fb5', bg: '#e6f0fa' },
  { icon: 'Money', title: '会费管理', desc: '查看缴费记录、在线缴纳年度会费', color: '#16a34a', bg: '#e6f7ee' },
  { icon: 'Medal', title: '证书下载', desc: '审核通过后在线下载会员证书', color: '#d97706', bg: '#fdf3e3' },
  { icon: 'Reading', title: '服务中心', desc: '浏览协会公告动态与行业文章', color: '#7c3aed', bg: '#f1ebfe' },
  { icon: 'ChatDotRound', title: '在线留言', desc: '向协会提交建议、诉求与反馈', color: '#dc2626', bg: '#fdeaea' },
  { icon: 'EditPen', title: '文章发布', desc: '发布行业动态、技术交流文章', color: '#0891b2', bg: '#e4f7fa' }
]

const serviceRouteMap: Record<string, string> = {
  在线入会: '/register',
  会费管理: '/member/fees',
  证书下载: '/member/certificates',
  服务中心: '/member/service',
  在线留言: '/member/messages',
  文章发布: '/member/articles'
}

function goService(f: { title: string }) {
  router.push(serviceRouteMap[f.title] || '/register')
}

// ── Announcements ──
const announcements = ref<any[]>([])
const loadingAnn = ref(false)
const activeTab = ref('all')
const tabs = [
  { label: '全部', value: 'all' },
  { label: '通知', value: 'notice' },
  { label: '文章', value: 'article' },
  { label: '政策', value: 'policy' }
]
const typeMap: Record<string, string> = { notice: '通知', article: '文章', policy: '政策' }
const typeTagMap: Record<string, string> = { notice: 'primary', article: 'success', policy: 'warning' }
function typeLabel(t: string) { return typeMap[t] || t }
function typeTag(t: string) { return typeTagMap[t] || 'info' }

const filteredAnnouncements = computed(() =>
  activeTab.value === 'all' ? announcements.value : announcements.value.filter(a => a.type === activeTab.value)
)

// ── Member levels ──
const levels = ref<any[]>([])
const loadingLevels = ref(false)
const rankColors = ['#3b82f6', '#0ea5e9', '#8b5cf6', '#f59e0b', '#22c55e']

// ── Divisions ──
const divisions = ref<any[]>([])

// ── Process ──
const processSteps = [
  { icon: 'EditPen', title: '提交申请', desc: '在线填写入会申请资料并提交' },
  { icon: 'View', title: '协会审核', desc: '协会对申请资料进行审核' },
  { icon: 'Wallet', title: '缴纳会费', desc: '审核通过后缴纳年度会费' },
  { icon: 'Medal', title: '成为会员', desc: '颁发会员证书，享会员权益' }
]

onMounted(async () => {
  await Promise.allSettled([fetchAnnouncements(), fetchLevels(), fetchDivisions()])
})

async function fetchAnnouncements() {
  loadingAnn.value = true
  try {
    const res = await announcementApi.getPublished({ page: 1, size: 8 })
    announcements.value = res.data.list || []
  } catch {} finally { loadingAnn.value = false }
}

async function fetchLevels() {
  loadingLevels.value = true
  try {
    const res = await orgApi.getMemberLevels()
    levels.value = (res.data || []).slice().sort((a: any, b: any) => (a.level ?? 0) - (b.level ?? 0))
  } catch {} finally { loadingLevels.value = false }
}

// 扁平化组织树：展示具体分支机构（含子分支），没有子分支时展示顶级机构
async function fetchDivisions() {
  try {
    const res = await orgApi.getTree()
    const tree: any[] = res.data || []
    const flat: any[] = []
    for (const root of tree) {
      const kids = root.children || []
      if (kids.length) {
        flat.push(...kids)
      } else {
        flat.push(root)
      }
    }
    divisions.value = flat
  } catch {}
}

function orgTypeLabel(t: string) {
  return t === 'branch' ? '分支机构' : t === 'representative' ? '代表机构' : t === 'association' ? '协会' : t || '机构'
}

function formatDate(d: string) { return d ? d.slice(0, 10) : '' }
</script>

<style scoped lang="scss">
$primary: #1a6fb5;
$primary-dark: #0b3d6f;
$primary-deep: #0a335c;
$text: #1f2937;
$text-secondary: #6b7280;
$text-light: #9ca3af;
$bg-soft: #f5f7fa;

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
}

.home-page {
  overflow-x: hidden;
}

// ════════════════════════════════════════════
// Section heading (shared)
// ════════════════════════════════════════════
.section-head {
  text-align: center;
  padding: 64px 0 40px;
  .section-eyebrow {
    display: inline-block;
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 3px;
    color: $primary;
    background: #e6f0fa;
    padding: 4px 14px;
    border-radius: 999px;
    margin-bottom: 14px;
  }
  h2 {
    font-size: 30px;
    font-weight: 700;
    color: $text;
    margin: 0;
  }
  p { color: $text-secondary; margin: 10px 0 0; font-size: 15px; }
}

// ════════════════════════════════════════════
// Hero
// ════════════════════════════════════════════
.hero {
  position: relative;
  background: linear-gradient(135deg, $primary-deep 0%, $primary-dark 40%, #145a9e 70%, $primary 100%);
  color: #fff;
  text-align: center;
  padding: 110px 24px 150px;
  overflow: hidden;

  .hero-decor {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 0;
  }
  .blob {
    position: absolute;
    border-radius: 50%;
    filter: blur(60px);
    opacity: 0.35;
  }
  .blob-1 { width: 420px; height: 420px; background: #3b82f6; top: -140px; left: -100px; }
  .blob-2 { width: 360px; height: 360px; background: #22d3ee; bottom: -120px; right: -80px; }
  .blob-3 { width: 280px; height: 280px; background: #818cf8; top: 10%; right: 22%; }
  .hero-grid {
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(rgba(255,255,255,0.05) 1px, transparent 1px),
      linear-gradient(90deg, rgba(255,255,255,0.05) 1px, transparent 1px);
    background-size: 46px 46px;
    mask-image: radial-gradient(ellipse at center, rgba(0,0,0,0.6) 0%, transparent 75%);
  }
  .float-chip {
    position: absolute;
    width: 52px;
    height: 52px;
    border-radius: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;
    background: rgba(255,255,255,0.14);
    border: 1px solid rgba(255,255,255,0.28);
    backdrop-filter: blur(8px);
    color: #fff;
    box-shadow: 0 8px 24px rgba(0,0,0,0.18);
    animation: float 6s ease-in-out infinite;
    z-index: 1;
  }
  .chip-1 { top: 22%; left: 10%; animation-delay: 0s; }
  .chip-2 { top: 58%; left: 16%; animation-delay: 1.4s; }
  .chip-3 { top: 24%; right: 12%; animation-delay: 0.7s; }
  .chip-4 { top: 62%; right: 8%; animation-delay: 2.1s; }

  .hero-content {
    position: relative;
    z-index: 2;
    animation: fadeUp 0.7s ease both;
  }
  .hero-badge {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    background: rgba(255,255,255,0.14);
    border: 1px solid rgba(255,255,255,0.3);
    padding: 6px 16px;
    border-radius: 999px;
    margin-bottom: 26px;
    backdrop-filter: blur(6px);
    .el-icon { color: #7dd3fc; }
  }
  h1 {
    font-size: 46px;
    font-weight: 800;
    margin: 0 0 14px;
    letter-spacing: 1px;
    background: linear-gradient(90deg, #fff 0%, #bfdbfe 100%);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .hero-subtitle {
    font-size: 20px;
    font-weight: 500;
    color: rgba(255,255,255,0.92);
    margin: 0 0 18px;
  }
  .hero-desc {
    max-width: 680px;
    margin: 0 auto 38px;
    font-size: 15px;
    line-height: 1.9;
    color: rgba(255,255,255,0.72);
  }
  .hero-actions {
    display: flex;
    gap: 16px;
    justify-content: center;
    flex-wrap: wrap;
  }
  .hero-btn-primary.el-button--primary {
    --el-button-bg-color: #ffffff;
    --el-button-border-color: rgba(255,255,255,0.7);
    --el-button-text-color: $primary-dark;
    --el-button-hover-bg-color: #dbeafe;
    --el-button-hover-border-color: #ffffff;
    --el-button-hover-text-color: $primary-deep;
    --el-button-active-bg-color: #bfdbfe;
    border-radius: 10px;
    font-weight: 600;
    padding: 14px 30px;
  }
  .hero-btn-ghost.el-button {
    --el-button-bg-color: rgba(255,255,255,0.1);
    --el-button-border-color: rgba(255,255,255,0.55);
    --el-button-text-color: #fff;
    --el-button-hover-bg-color: rgba(255,255,255,0.2);
    --el-button-hover-border-color: #fff;
    --el-button-hover-text-color: #fff;
    border-radius: 10px;
    padding: 14px 30px;
  }
  .hero-features {
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
    gap: 12px 28px;
    margin-top: 40px;
    .hf-item {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 14px;
      color: rgba(255,255,255,0.85);
      .el-icon { color: #6ee7b7; }
    }
  }
  .hero-wave {
    position: absolute;
    left: 0;
    right: 0;
    bottom: -1px;
    z-index: 2;
    svg { display: block; width: 100%; height: 80px; }
  }
}

// ════════════════════════════════════════════
// Services
// ════════════════════════════════════════════
.services {
  padding-bottom: 20px;
}
.service-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
  padding-bottom: 40px;
}
.service-card {
  background: #fff;
  border: 1px solid #eef1f5;
  border-radius: 16px;
  padding: 30px 26px;
  cursor: pointer;
  transition: transform 0.25s, box-shadow 0.25s, border-color 0.25s;
  &:hover {
    transform: translateY(-6px);
    box-shadow: 0 16px 34px rgba(26, 111, 181, 0.12);
    border-color: #bcd7f0;
    .service-more { color: $primary; }
  }
  .service-icon {
    width: 56px;
    height: 56px;
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 20px;
  }
  h3 { margin: 0 0 10px; font-size: 18px; color: $text; }
  p { margin: 0 0 18px; font-size: 14px; line-height: 1.7; color: $text-secondary; }
  .service-more {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 13px;
    color: $text-light;
    transition: color 0.2s;
    .el-icon { font-size: 14px; }
  }
}

// ════════════════════════════════════════════
// News (announcements + levels)
// ════════════════════════════════════════════
.news-section {
  background: $bg-soft;
  padding: 16px 0 72px;
  margin-top: 24px;
}
.news-grid {
  display: grid;
  grid-template-columns: 1.6fr 1fr;
  gap: 24px;
  align-items: stretch;
}
.news-card {
  background: #fff;
  border: 1px solid #eef1f5;
  border-radius: 16px;
  padding: 24px 26px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.04);
  .card-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    h3 {
      margin: 0;
      font-size: 18px;
      color: $text;
      display: flex;
      align-items: center;
      gap: 8px;
      .el-icon { color: $primary; }
    }
  }
}
.type-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 14px;
  .tab {
    font-size: 13px;
    padding: 5px 14px;
    border-radius: 999px;
    color: $text-secondary;
    background: $bg-soft;
    cursor: pointer;
    transition: all 0.2s;
    &:hover { color: $primary; }
    &.active {
      background: #e6f0fa;
      color: $primary;
      font-weight: 600;
    }
  }
}
.announce-list {
  min-height: 120px;
}
.announce-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 13px 6px;
  border-bottom: 1px solid #f3f5f7;
  cursor: pointer;
  transition: background 0.2s;
  &:hover { background: #f8fafc; .announce-title { color: $primary; } }
  .pin-tag { flex-shrink: 0; }
  .announce-title {
    flex: 1;
    font-size: 14px;
    color: #374151;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    transition: color 0.2s;
  }
  .announce-meta {
    display: flex;
    align-items: center;
    gap: 14px;
    flex-shrink: 0;
    color: $text-light;
    font-size: 12px;
    .views { display: inline-flex; align-items: center; gap: 3px; }
    .date { white-space: nowrap; }
  }
}
.level-list {
  min-height: 120px;
}
.level-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 6px;
  border-bottom: 1px solid #f3f5f7;
  .level-rank {
    width: 38px;
    height: 38px;
    border-radius: 10px;
    color: #fff;
    font-size: 16px;
    font-weight: 700;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .level-name { font-size: 15px; font-weight: 600; color: $text; }
  .level-desc {
    font-size: 13px;
    color: $text-secondary;
    margin-top: 3px;
    line-height: 1.5;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
}

// ════════════════════════════════════════════
// Process
// ════════════════════════════════════════════
.process-section {
  padding-bottom: 40px;
}
.process-steps {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 24px;
  padding: 20px 0 56px;
  .step {
    text-align: center;
    position: relative;
  }
  .step-node { position: relative; margin-bottom: 20px; }
  .step-circle {
    width: 72px;
    height: 72px;
    margin: 0 auto;
    border-radius: 50%;
    background: linear-gradient(135deg, #e6f0fa, #cfe3f6);
    color: $primary;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 3px solid #fff;
    box-shadow: 0 8px 20px rgba(26, 111, 181, 0.18);
    position: relative;
    z-index: 2;
  }
  .step-line {
    position: absolute;
    top: 36px;
    left: calc(50% + 40px);
    width: calc(100% - 76px);
    height: 2px;
    background: repeating-linear-gradient(90deg, #c3d9f0 0 6px, transparent 6px 12px);
  }
  h4 { margin: 0 0 8px; font-size: 16px; color: $text; }
  p { margin: 0; font-size: 13px; color: $text-secondary; line-height: 1.6; }
}

// ════════════════════════════════════════════
// Organizations
// ════════════════════════════════════════════
.org-section {
  background: $bg-soft;
  padding: 16px 0 72px;
}
.org-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}
.org-card {
  background: #fff;
  border: 1px solid #eef1f5;
  border-radius: 16px;
  padding: 24px;
  transition: transform 0.25s, box-shadow 0.25s;
  &:hover { transform: translateY(-4px); box-shadow: 0 12px 28px rgba(0,0,0,0.08); }
  .org-head { display: flex; align-items: center; gap: 12px; margin-bottom: 14px; }
  .org-icon {
    width: 44px;
    height: 44px;
    border-radius: 12px;
    background: #e6f0fa;
    color: $primary;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .org-name { font-size: 16px; font-weight: 600; color: $text; }
  .org-desc {
    font-size: 13px;
    color: $text-secondary;
    line-height: 1.7;
    min-height: 44px;
    margin: 0 0 16px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .org-foot {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
    .org-type {
      font-size: 12px;
      color: $primary;
      background: #e6f0fa;
      padding: 3px 10px;
      border-radius: 999px;
      white-space: nowrap;
    }
    .org-contact {
      display: inline-flex;
      align-items: center;
      gap: 4px;
      font-size: 12px;
      color: $text-light;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
}

// ════════════════════════════════════════════
// CTA
// ════════════════════════════════════════════
.cta-section {
  padding: 56px 0 72px;
}
.cta-banner {
  position: relative;
  overflow: hidden;
  background: linear-gradient(120deg, $primary-deep 0%, $primary-dark 50%, $primary 100%);
  border-radius: 20px;
  padding: 48px 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 28px;
  flex-wrap: wrap;
  color: #fff;
  box-shadow: 0 18px 40px rgba(11, 61, 111, 0.28);
  .cta-decor {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }
  .cta-blob { position: absolute; border-radius: 50%; filter: blur(50px); opacity: 0.3; }
  .cta-blob-1 { width: 240px; height: 240px; background: #22d3ee; top: -90px; right: 8%; }
  .cta-blob-2 { width: 200px; height: 200px; background: #818cf8; bottom: -80px; left: 6%; }
  .cta-text {
    position: relative;
    z-index: 1;
    h3 { margin: 0 0 8px; font-size: 24px; font-weight: 700; }
    p { margin: 0; font-size: 14px; color: rgba(255,255,255,0.75); }
  }
  .cta-actions {
    position: relative;
    z-index: 1;
    display: flex;
    gap: 14px;
    flex-wrap: wrap;
  }
  .cta-btn.el-button--primary {
    --el-button-bg-color: #ffffff;
    --el-button-border-color: #ffffff;
    --el-button-text-color: $primary-dark;
    --el-button-hover-bg-color: #dbeafe;
    --el-button-hover-border-color: #ffffff;
    --el-button-hover-text-color: $primary-deep;
    border-radius: 10px;
    font-weight: 600;
    padding: 14px 30px;
  }
  .cta-btn-ghost.el-button {
    --el-button-bg-color: transparent;
    --el-button-border-color: rgba(255,255,255,0.6);
    --el-button-text-color: #fff;
    --el-button-hover-bg-color: rgba(255,255,255,0.15);
    --el-button-hover-border-color: #fff;
    --el-button-hover-text-color: #fff;
    border-radius: 10px;
    padding: 14px 30px;
  }
}

// ════════════════════════════════════════════
// Animations & Responsive
// ════════════════════════════════════════════
@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-14px); }
}
@keyframes fadeUp {
  from { opacity: 0; transform: translateY(24px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (max-width: 992px) {
  .service-grid, .org-grid { grid-template-columns: repeat(2, 1fr); }
  .news-grid { grid-template-columns: 1fr; }
  .process-steps { grid-template-columns: repeat(2, 1fr); row-gap: 36px; }
  .hero {
    padding: 90px 24px 130px;
    h1 { font-size: 36px; }
    .float-chip { display: none; }
  }
}
@media (max-width: 640px) {
  .service-grid, .org-grid { grid-template-columns: 1fr; }
  .process-steps { grid-template-columns: 1fr; }
  .hero h1 { font-size: 30px; }
  .cta-banner { padding: 36px 26px; }
}
</style>
