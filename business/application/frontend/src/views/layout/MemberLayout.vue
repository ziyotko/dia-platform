<template>
  <el-container style="height:100vh">
    <el-aside :width="collapsed ? '64px' : '220px'" class="sidebar">
      <div class="logo-area" @click="$router.push('/member/dashboard')">
        <span v-if="!collapsed">申报系统</span>
        <span v-else>申</span>
      </div>
      <el-menu :default-active="route.path" router :collapse="collapsed" background-color="#fff" text-color="#374151" active-text-color="#002fa7">
        <el-menu-item index="/member/dashboard"><el-icon><DataAnalysis /></el-icon><span>申报中心</span></el-menu-item>
        <el-menu-item index="/member/batches"><el-icon><Files /></el-icon><span>可申报项目</span></el-menu-item>
        <el-menu-item index="/member/applications"><el-icon><Tickets /></el-icon><span>我的申报</span></el-menu-item>
        <el-menu-item index="/member/announcements"><el-icon><Bell /></el-icon><span>结果公示</span></el-menu-item>
        <el-menu-item index="/member/certificates"><el-icon><Medal /></el-icon><span>我的证书</span></el-menu-item>
        <el-menu-item index="/member/notifications">
          <el-icon><Message /></el-icon><span>通知中心</span>
          <el-badge v-if="unread > 0" :value="unread" :max="99" class="menu-badge" />
        </el-menu-item>
        <el-menu-item index="/member/profile"><el-icon><User /></el-icon><span>个人资料</span></el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="topbar">
        <div class="topbar-left">
          <el-icon class="collapse-btn" @click="toggleCollapse"><Fold /></el-icon>
        </div>
        <div class="topbar-right">
          <el-dropdown>
            <span class="user-info">{{ userStore.userInfo?.realName || '申报人' }} <el-icon><ArrowDown /></el-icon></span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="$router.push('/member/profile')">个人资料</el-dropdown-item>
                <el-dropdown-item divided @click="userStore.logout()">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main-content"><router-view /></el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import { memberApi } from '@/api/member'

const route = useRoute()
const userStore = useUserStore()
const appStore = useAppStore()
const { collapsed, toggleCollapse } = appStore

const unread = ref(0)
let timer: number | undefined

async function refreshUnread() {
  try {
    const res = await memberApi.getUnreadCount()
    unread.value = res.data?.count || 0
  } catch {
    // ignore transient errors
  }
}

// 页面聚焦时立即刷新（例如从后台标签页回来）
function onVisible() {
  if (!document.hidden) refreshUnread()
}

onMounted(() => {
  if (!userStore.userInfo) userStore.fetchUserInfo()
  refreshUnread()
  // 红点不再只靠路由切换刷新：轮询 + 可见性变化，站内信一到就能看到
  timer = window.setInterval(refreshUnread, 60000)
  document.addEventListener('visibilitychange', onVisible)
})

onUnmounted(() => {
  if (timer) window.clearInterval(timer)
  document.removeEventListener('visibilitychange', onVisible)
})

watch(() => route.fullPath, refreshUnread)
</script>

<style scoped>
.sidebar { background: #fff; border-right: 1px solid var(--app-border); overflow: hidden; transition: width .3s; }
.logo-area { height: 60px; display: flex; align-items: center; justify-content: center; color: #002fa7; font-size: 18px; font-weight: bold; cursor: pointer; border-bottom: 1px solid var(--app-border); }
.topbar { background: #fff; box-shadow: 0 1px 4px rgba(0,0,0,.08); display: flex; align-items: center; justify-content: space-between; padding: 0 20px; }
.collapse-btn { font-size: 20px; cursor: pointer; }
.topbar-right { display: flex; align-items: center; gap: 20px; }
.user-info { display: flex; align-items: center; gap: 4px; cursor: pointer; }
.main-content { padding: 20px; background: var(--app-bg); }
.menu-badge { margin-left: 8px; vertical-align: middle; }
.menu-badge :deep(.el-badge__content) { border: none; }
</style>
