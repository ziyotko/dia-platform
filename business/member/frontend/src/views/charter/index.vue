<template>
  <div class="charter-page">
    <!-- 阅读进度条 -->
    <div class="read-progress" aria-hidden="true">
      <span :style="{ width: progress + '%' }"></span>
    </div>
    <section class="charter-hero">
      <div class="hero-decor" aria-hidden="true">
        <span class="decor-blob blob-a"></span>
        <span class="decor-blob blob-b"></span>
        <span class="decor-glow"></span>
        <span class="decor-grid"></span>
      </div>
      <div class="hero-inner">
        <h1>协会章程</h1>
        <p class="hero-desc">本章程经协会审议通过，面向社会公众公开查阅，内容由协会管理员维护更新</p>
      </div>
    </section>

    <div class="charter-body">
      <div class="charter-card">
          <!-- 富文本正文（样式见全局 .rich-text-content） -->
          <template v-if="content">
            <article class="rich-text-content charter-article" v-html="sanitizeHtml(content)"></article>
          </template>

          <el-empty v-else description="章程暂未维护，请联系管理员" />
      </div>

      <div class="charter-actions">
        <el-button class="back-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回上一页
        </el-button>
        <el-button v-if="content" @click="printPage">
          <el-icon><Printer /></el-icon>
          打印本章程
        </el-button>
        <el-button v-if="file" type="primary" @click="downloadPdf">
          <el-icon><Download /></el-icon>
          下载章程 PDF
        </el-button>
      </div>
    </div>

    <!-- 回到顶部 -->
    <button class="to-top" type="button" aria-label="回到顶部" v-show="showTop" @click="scrollToTop">
      <el-icon><Top /></el-icon>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { charterApi } from '@/api/index'
import { sanitizeHtml } from '@/utils/sanitizeHtml'

interface CharterFile {
  path?: string
  name?: string
  size?: number
  uploaded_at?: string
}

const router = useRouter()

const content = ref('')
const file = ref<CharterFile | null>(null)
const loading = ref(true)

const progress = ref(0)
const showTop = ref(false)

// 正文字数：剥离 HTML 标签与空白后的字符数
const wordCount = computed(() =>
  content.value.replace(/<[^>]*>/g, '').replace(/&nbsp;/g, ' ').replace(/\s+/g, '').length
)

onMounted(async () => {
  window.addEventListener('scroll', onScroll, { passive: true })
  try {
    const res: any = await charterApi.getContent()
    content.value = res.data?.content || ''
    file.value = res.data?.file || null
  } catch {
    content.value = ''
    file.value = null
  } finally {
    loading.value = false
  }
  updateScrollState()
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll)
})

let ticking = false

function onScroll() {
  if (ticking) return
  ticking = true
  window.requestAnimationFrame(() => {
    ticking = false
    updateScrollState()
  })
}

function updateScrollState() {
  const y = window.scrollY || document.documentElement.scrollTop || 0
  const max = document.documentElement.scrollHeight - window.innerHeight
  progress.value = max > 0 ? Math.min(100, Math.max(0, (y / max) * 100)) : 0
  showTop.value = y > 480
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function goBack() {
  const state = window.history.state as { back?: string } | null
  if (state && state.back) router.back()
  else router.push('/')
}

function printPage() {
  window.print()
}

function downloadPdf() {
  const base = import.meta.env.VITE_API_BASE_URL || '/business_member/api'
  window.open(`${base}/charter`, '_blank')
}

function formatSize(n?: number) {
  if (!n) return ''
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(2)} MB`
}

function formatDate(d?: string) {
  return d ? d.slice(0, 10) : ''
}
</script>

<style scoped lang="scss">
$primary: #002fa7;
$primary-deep: #001b5e;
$primary-light: #2050cf;

.charter-page {
  padding-bottom: 64px;

  // 全局 .el-icon 把 --color 固定为主色，在深色横幅里几乎看不见；这里改为跟随文字颜色
  .el-icon {
    --color: inherit;
  }
}

// ════════════════════════════════════════════
// 顶部阅读进度条
// ════════════════════════════════════════════
.read-progress {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 200;
  height: 3px;
  pointer-events: none;

  span {
    display: block;
    width: 0;
    height: 100%;
    background: linear-gradient(90deg, $primary-deep, $primary 55%, $primary-light);
    transition: width 0.12s linear;
  }
}

// ════════════════════════════════════════════
// 品牌横幅
// ════════════════════════════════════════════
.charter-hero {
  position: relative;
  overflow: hidden;
  padding: 56px 24px 84px;
  text-align: center;
  color: #fff;
  background: linear-gradient(135deg, $primary-deep 0%, $primary 58%, $primary-light 100%);

  .hero-decor {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }
  .decor-blob {
    position: absolute;
    border-radius: 50%;
  }
  .blob-a {
    width: 340px;
    height: 340px;
    top: -180px;
    left: -90px;
    background: rgba(255, 255, 255, 0.07);
  }
  .blob-b {
    width: 260px;
    height: 260px;
    right: -70px;
    bottom: -150px;
    background: rgba(255, 255, 255, 0.06);
  }
  .decor-grid {
    position: absolute;
    inset: 0;
    background-image: linear-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px),
      linear-gradient(90deg, rgba(255, 255, 255, 0.05) 1px, transparent 1px);
    background-size: 48px 48px;
  }
  .decor-glow {
    position: absolute;
    top: -46%;
    left: 50%;
    width: 640px;
    height: 640px;
    transform: translateX(-50%);
    border-radius: 50%;
    background: radial-gradient(circle, rgba(255, 255, 255, 0.2) 0%, rgba(255, 255, 255, 0) 62%);
  }

  .hero-inner {
    position: relative;
    z-index: 1;
    max-width: 860px;
    margin: 0 auto;
  }

  .hero-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 5px 14px;
    margin-bottom: 18px;
    font-size: 13px;
    font-weight: 500;
    letter-spacing: 0.4px;
    border-radius: 999px;
    color: #fff;
    background: rgba(255, 255, 255, 0.14);
    border: 1px solid rgba(255, 255, 255, 0.24);
  }

  h1 {
    margin: 0;
    font-size: 34px;
    font-weight: 700;
    letter-spacing: 2px;
  }

  .hero-desc {
    margin: 14px 0 0;
    font-size: 14px;
    color: rgba(255, 255, 255, 0.82);
  }

  .hero-meta {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 10px 12px;
    margin: 24px 0 0;
    padding: 0;
    list-style: none;

    li {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 6px 14px;
      font-size: 13px;
      line-height: 1.4;
      color: rgba(255, 255, 255, 0.9);
      background: rgba(255, 255, 255, 0.1);
      border: 1px solid rgba(255, 255, 255, 0.18);
      border-radius: 999px;
      backdrop-filter: blur(2px);

      .el-icon {
        font-size: 14px;
        opacity: 0.85;
      }
    }
  }
}

// ════════════════════════════════════════════
// 正文卡片
// ════════════════════════════════════════════
.charter-body {
  position: relative;
  z-index: 2;
  max-width: 860px;
  margin: -44px auto 0;
  padding: 0 24px;
}

.charter-card {
  position: relative;
  min-height: 120px;
  padding: 36px 44px 44px;
  overflow: hidden;
  background: var(--app-card-bg);
  border: 1px solid var(--app-border);
  border-radius: var(--app-card-radius);
  box-shadow: 0 12px 32px rgba(16, 24, 40, 0.1);
  animation: charter-fade-up 0.5s ease 0.05s both;

}

.charter-skeleton {
  padding: 6px 2px;
}

// ── PDF 附件条 ──
.file-bar {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 18px;
  margin-bottom: 32px;
  background: linear-gradient(135deg, #f7f9ff, #eef2fd);
  border: 1px solid var(--app-border);
  border-radius: 12px;
  transition: border-color 0.25s ease, box-shadow 0.25s ease;

  &:hover {
    border-color: var(--el-color-primary-light-7);
    box-shadow: 0 6px 18px rgba(0, 47, 167, 0.08);
  }

  .file-icon {
    flex: none;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--el-color-primary);
    background: #fff;
    border-radius: 10px;
    box-shadow: 0 2px 8px rgba(0, 47, 167, 0.12);
  }

  .file-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .file-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--app-text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .file-meta {
    font-size: 12px;
    color: var(--app-text-muted);
  }
}

// ── 「章程全文」分隔标题 ──
.card-head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 28px;

  .head-line {
    flex: 1;
    height: 1px;
    background: linear-gradient(90deg, transparent, var(--app-border));

    &:last-child {
      background: linear-gradient(90deg, var(--app-border), transparent);
    }
  }

  .head-text {
    flex: none;
    font-size: 15px;
    font-weight: 600;
    letter-spacing: 3px;
    color: var(--el-color-primary);
  }
}

.charter-article {
  // 首个/末个块不需要额外外边距
  :deep(> :first-child) {
    margin-top: 0;
  }
  :deep(> :last-child) {
    margin-bottom: 0;
  }
}

// ── 底部操作区 ──
.charter-actions {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  margin-top: 28px;

  .back-btn {
    padding: 0 22px;
  }
}

@keyframes charter-fade-up {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

// ── 回到顶部 ──
.to-top {
  position: fixed;
  right: 28px;
  bottom: 34px;
  z-index: 90;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  color: #fff;
  font-size: 18px;
  background: linear-gradient(135deg, $primary, $primary-light);
  border: none;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 6px 18px rgba(0, 47, 167, 0.32);
  transition: transform 0.2s ease, box-shadow 0.2s ease;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 10px 24px rgba(0, 47, 167, 0.38);
  }
}

@media (prefers-reduced-motion: reduce) {
  .charter-card {
    animation: none;
  }
  .read-progress span {
    transition: none;
  }
}

// ════════════════════════════════════════════
// 响应式
// ════════════════════════════════════════════
@media (max-width: 992px) {
  .charter-hero {
    padding: 44px 20px 76px;

    h1 {
      font-size: 28px;
    }
  }
  .charter-card {
    padding: 28px 26px 32px;
  }
}

@media (max-width: 640px) {
  .charter-hero {
    padding: 36px 18px 68px;

    h1 {
      font-size: 24px;
      letter-spacing: 1px;
    }
    .hero-desc {
      font-size: 13px;
    }
    .hero-meta {
      gap: 8px;
      margin-top: 18px;

      li {
        padding: 5px 12px;
        font-size: 12px;
      }
    }
  }
  .charter-body {
    margin-top: -32px;
    padding: 0 16px;
  }
  .charter-card {
    padding: 22px 18px 26px;
  }
  .file-bar {
    flex-wrap: wrap;
    gap: 12px;

    .el-button {
      width: 100%;
    }
  }
  .card-head .head-text {
    letter-spacing: 2px;
  }
  .charter-actions {
    flex-direction: column;
    gap: 8px;

    .el-button {
      width: 100%;
      margin-left: 0;
    }
  }
  .to-top {
    right: 16px;
    bottom: 20px;
    width: 38px;
    height: 38px;
  }
}

// ════════════════════════════════════════════
// 打印：去掉横幅与操作区，只保留正文
// ════════════════════════════════════════════
@media print {
  .charter-page .charter-hero,
  .charter-page .charter-actions,
  .charter-page .file-bar,
  .charter-page .card-head,
  .charter-page .to-top,
  .charter-page .read-progress {
    display: none;
  }
  .charter-body {
    margin-top: 0;
    max-width: none;
    padding: 0;
  }
  .charter-card {
    padding: 0;
    border: none;
    box-shadow: none;
    animation: none;

    &::before {
      display: none;
    }
  }
}
</style>
