<template>
  <div class="public-layout">
    <header class="public-header" :class="{ 'is-scrolled': scrolled }">
      <div class="header-inner">
        <div class="brand" @click="goHome">
          <span class="brand-mark">{{ brandInitial }}</span>
          <span class="brand-text">
            <span class="brand-name">{{ siteStore.site_name }}</span>
            <span class="brand-sub">会员服务平台</span>
          </span>
        </div>

        <!-- 桌面端导航 -->
        <nav class="nav-links">
          <router-link v-for="item in navItems" :key="item.to" :to="item.to" class="nav-item">
            {{ item.label }}
          </router-link>
          <span class="nav-divider" aria-hidden="true"></span>
          <router-link to="/login" class="nav-item">
            <el-icon><User /></el-icon>
            会员登录
          </router-link>
          <router-link to="/register" class="nav-register">
            <el-button type="primary" round>注册账号</el-button>
          </router-link>
        </nav>

        <!-- 移动端菜单开关 -->
        <button
          class="nav-toggle"
          type="button"
          aria-label="打开导航菜单"
          :aria-expanded="menuOpen ? 'true' : 'false'"
          @click="menuOpen = !menuOpen"
        >
          <el-icon><Close v-if="menuOpen" /><Menu v-else /></el-icon>
        </button>
      </div>

      <!-- 移动端下拉导航 -->
      <transition name="nav-drop">
        <nav class="nav-mobile" v-show="menuOpen">
          <router-link
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="nav-mobile-item"
            @click="menuOpen = false"
          >
            {{ item.label }}
          </router-link>
          <router-link to="/login" class="nav-mobile-item" @click="menuOpen = false">会员登录</router-link>
          <router-link
            to="/register"
            class="nav-mobile-item nav-mobile-register"
            @click="menuOpen = false"
          >
            <el-button type="primary" round>注册账号</el-button>
          </router-link>
        </nav>
      </transition>
    </header>
    <main class="public-main">
      <router-view />
    </main>
    <footer class="public-footer">
      <p v-if="siteStore.copyright_name">版权所有：{{ siteStore.copyright_name }}</p>
      <p v-if="siteStore.icp_no || siteStore.beian_no">
        <span v-if="siteStore.icp_no">ICP备案号：{{ siteStore.icp_no }}</span>
        <span v-if="siteStore.icp_no && siteStore.beian_no" class="footer-sep">　</span>
        <span v-if="siteStore.beian_no">网安备案号：{{ siteStore.beian_no }}</span>
      </p>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useSiteStore } from '@/stores/site'

const router = useRouter()
const siteStore = useSiteStore()

// 公共导航项（登录/注册单独排布）
const navItems = [
  { to: '/announcements', label: '公告动态' },
  { to: '/charter', label: '协会章程' }
]

const menuOpen = ref(false)
const scrolled = ref(false)

// 站点名首字作为品牌标识
const brandInitial = computed(() => (siteStore.site_name || '会').trim().charAt(0))

function goHome() {
  router.push('/')
}

// 滚动后给吸顶导航加分隔与投影
function onScroll() {
  scrolled.value = (window.scrollY || 0) > 8
}

onMounted(() => {
  window.addEventListener('scroll', onScroll, { passive: true })
  onScroll()
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll)
})
</script>

<style scoped lang="scss">
.public-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

// ════════════════════════════════════════════
// 顶部导航（吸顶，滚动后加深分隔与投影）
// ════════════════════════════════════════════
.public-header {
  position: sticky;
  top: 0;
  z-index: 100;
  background: rgba(255, 255, 255, 0.86);
  backdrop-filter: saturate(180%) blur(12px);
  border-bottom: 1px solid transparent;
  transition: background 0.25s ease, border-color 0.25s ease, box-shadow 0.25s ease;

  &.is-scrolled {
    background: rgba(255, 255, 255, 0.96);
    border-bottom-color: var(--app-border);
    box-shadow: 0 4px 18px rgba(16, 24, 40, 0.06);
  }

  .header-inner {
    max-width: 1200px;
    margin: 0 auto;
    height: 64px;
    padding: 0 24px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }
}

// ── 品牌区 ──
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  cursor: pointer;
  transition: opacity 0.2s ease;

  &:hover {
    opacity: 0.86;
  }

  .brand-mark {
    flex: none;
    width: 34px;
    height: 34px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    font-weight: 700;
    color: #fff;
    border-radius: 10px;
    background: linear-gradient(135deg, #001b5e, #002fa7 55%, #2050cf);
    box-shadow: 0 2px 8px rgba(0, 47, 167, 0.28);
  }

  .brand-text {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .brand-name {
    font-size: 17px;
    font-weight: 700;
    line-height: 1.25;
    letter-spacing: 0.2px;
    color: var(--app-text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .brand-sub {
    font-size: 11px;
    line-height: 1.3;
    letter-spacing: 1.6px;
    color: var(--app-text-muted);
  }
}

// ── 桌面端导航 ──
.nav-links {
  display: flex;
  align-items: center;
  gap: 4px;

  .nav-item {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 8px 14px;
    font-size: 15px;
    line-height: 1;
    color: var(--app-text-secondary);
    text-decoration: none;
    border-radius: 8px;
    transition: color 0.2s ease, background 0.2s ease;

    .el-icon {
      --color: inherit;
      font-size: 15px;
    }

    &:hover {
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
    }

    // 当前所在页（含公告详情等子路由）
    &.router-link-active {
      color: var(--el-color-primary);
      font-weight: 600;
      background: var(--el-color-primary-light-9);
    }
  }

  .nav-divider {
    width: 1px;
    height: 18px;
    margin: 0 10px;
    background: var(--app-border);
  }

  .nav-register {
    display: inline-flex;
    margin-left: 6px;
  }
}

// ── 移动端菜单开关 ──
.nav-toggle {
  display: none;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  font-size: 19px;
  color: var(--app-text-primary);
  background: #fff;
  border: 1px solid var(--app-border);
  border-radius: 10px;
  cursor: pointer;
  transition: color 0.2s ease, border-color 0.2s ease, background 0.2s ease;

  .el-icon {
    --color: inherit;
  }

  &:hover {
    color: var(--el-color-primary);
    border-color: var(--el-color-primary-light-7);
    background: var(--el-color-primary-light-9);
  }
}

.nav-mobile {
  display: none;
  flex-direction: column;
  gap: 4px;
  padding: 10px 16px 18px;
  background: #fff;
  border-top: 1px solid var(--app-border);
  box-shadow: 0 8px 20px rgba(16, 24, 40, 0.08);

  .nav-mobile-item {
    padding: 11px 12px;
    font-size: 15px;
    color: var(--app-text-secondary);
    text-decoration: none;
    border-radius: 8px;
    transition: color 0.2s ease, background 0.2s ease;

    &:hover,
    &.router-link-active {
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
    }
  }

  .nav-mobile-register {
    padding: 0;

    &:hover {
      background: none;
    }

    .el-button {
      width: 100%;
      margin-top: 6px;
    }
  }
}

.nav-drop-enter-active,
.nav-drop-leave-active {
  transition: opacity 0.22s ease, transform 0.22s ease;
}

.nav-drop-enter-from,
.nav-drop-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.public-main {
  flex: 1;
}

.public-footer {
  text-align: center;
  padding: 32px 24px;
  background: #1f2937;
  color: #9ca3af;
  font-size: 13px;
  line-height: 1.8;
  .footer-sep {
    margin: 0 8px;
  }
}

// ════════════════════════════════════════════
// 响应式：窄屏收起为下拉菜单
// ════════════════════════════════════════════
@media (max-width: 768px) {
  .public-header .header-inner {
    padding: 0 16px;
  }
  .nav-links {
    display: none;
  }
  .nav-toggle {
    display: inline-flex;
  }
  .nav-mobile {
    display: flex;
  }
}
</style>
