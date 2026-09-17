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
          <span>全国性行业社会团体 · 会员服务平台</span>
        </div>
        <h1>{{ siteStore.site_name }}</h1>
        <p class="hero-subtitle">服务政府 · 服务行业 · 服务企业</p>
        <p class="hero-desc">
          中国环保机械行业协会（CAMIE）成立于 1994 年，是经民政部批准注册登记的全国性社会团体独立法人单位，
          以环保机械和资源综合利用装备制造厂商为主干。协会围绕行业调研、团体标准制定、科技成果评议与国际
          交流合作开展工作，本平台为会员单位提供信息共享等一站式服务。
        </p>
        <div class="hero-features">
          <div class="hf-item" v-for="f in heroFeatures" :key="f"><el-icon><Check /></el-icon>{{ f }}</div>
        </div>
      </div>
      <div class="hero-wave" aria-hidden="true">
        <svg viewBox="0 0 1440 90" preserveAspectRatio="none">
          <path d="M0,60 C240,95 480,25 720,40 C960,55 1200,90 1440,55 L1440,90 L0,90 Z" fill="#f5f7fb"></path>
        </svg>
      </div>
    </section>
    <!-- ═══════════════ 入会流程 ═══════════════ -->
    <section class="process-section">
      <div class="container">
        <div class="section-head">
          <h2>入会流程</h2>
          <p>先注册会员，再提交申请，五步完成入会，全程在线跟踪办理进度</p>
        </div>
        <div class="process-steps">
          <div class="step" v-for="(s, i) in processSteps" :key="s.title">
            <div class="step-node">
              <div class="step-circle"><el-icon :size="24"><component :is="s.icon" /></el-icon></div>
              <span v-if="i < processSteps.length - 1" class="step-line"></span>
            </div>
            <h4>{{ s.title }}</h4>
            <p>{{ s.desc }}</p>
            <el-button v-if="s.to" link type="primary" class="step-action" @click="goStep(s)">
              {{ s.action }}<el-icon class="el-icon--right"><ArrowRight /></el-icon>
            </el-button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useSiteStore } from '@/stores/site'

const router = useRouter()
const siteStore = useSiteStore()

// ── Hero ──
const heroFeatures = ['1994 年成立', '经民政部注册登记', '4A 级全国社会组织', '会员单位近 500 家']

// ── Process ──
// 入会分两个阶段：先在注册页申请成为「注册会员」，
// 再登录会员中心在「我的申请」提交入会申请成为「会员单位」
interface ProcessStep { icon: string; title: string; desc: string; action?: string; to?: string }
const processSteps: ProcessStep[] = [
  {
    icon: 'UserFilled',
    title: '注册账号',
    desc: '在线注册成为注册会员',
    action: '立即注册',
    to: '/register'
  },
  {
    icon: 'DocumentAdd',
    title: '入会申请',
    desc: '提交入会申请表，签字盖章后上传',
    action: '发起入会申请',
    to: '/member/applications'
  },
  { icon: 'View', title: '协会审核', desc: '协会对申请资料进行审核' },
  { icon: 'Wallet', title: '缴纳会费', desc: '审核通过后按会员等级缴纳年度会费' },
  { icon: 'Medal', title: '成为会员', desc: '颁发会员证书，享会员权益与服务' }
]

function goStep(s: ProcessStep) {
  if (s.to) router.push(s.to)
}
</script>

<style scoped lang="scss">
$primary: #002fa7;
$text: #1f2937;
$text-secondary: #6b7280;

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
  background: linear-gradient(135deg, #001b5e 0%, #002fa7 55%, #2050cf 100%);
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
  .blob-1 { width: 420px; height: 420px; background: #002fa7; top: -140px; left: -100px; }
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
    background: linear-gradient(90deg, #fff 0%, #b3c1e5 100%);
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
// Process
// ════════════════════════════════════════════
.process-section {
  padding-bottom: 40px;
}
.process-steps {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 24px;
  padding: 20px 0 40px;
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
    background: linear-gradient(135deg, #e6eaf6, #cfe3f6);
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
  .step-action {
    margin-top: 8px;
    padding: 0;
    height: auto;
    font-size: 13px;
    font-weight: 600;
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
  // 换行后连接线不再连续，直接隐藏避免出现悬空虚线
  .process-steps { grid-template-columns: repeat(2, 1fr); row-gap: 36px; .step-line { display: none; } }
  .hero {
    padding: 90px 24px 130px;
    h1 { font-size: 36px; }
    .float-chip { display: none; }
  }
}
@media (max-width: 640px) {
  .process-steps { grid-template-columns: 1fr; }
  .hero h1 { font-size: 30px; }
}
</style>
