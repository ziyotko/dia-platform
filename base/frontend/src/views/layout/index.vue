<template>
  <el-container class="layout-container">
    <el-aside :width="appStore.collapsed ? '64px' : '230px'" class="sidebar">
      <div class="logo">
        <div class="logo-icon">
          <el-icon :size="26" color="#fff"><Management /></el-icon>
        </div>
        <span v-show="!appStore.collapsed" class="logo-text">Base 平台</span>
      </div>
      <el-scrollbar class="menu-scroll">
        <el-menu
          :default-active="activeMenu"
          :collapse="appStore.collapsed"
          :collapse-transition="false"
          router
          background-color="transparent"
          text-color="#94a3b8"
          active-text-color="#fff"
        >
          <sub-menu v-for="menu in userStore.menus" :key="menu.id" :menu="menu" />
        </el-menu>
      </el-scrollbar>
    </el-aside>
    <el-container class="main-wrapper">
      <el-header class="header">
        <div class="header-left">
          <div class="collapse-btn" @click="appStore.toggleCollapse">
            <el-icon :size="20">
              <Fold v-if="!appStore.collapsed" />
              <Expand v-else />
            </el-icon>
          </div>
          <breadcrumb />
        </div>
        <div class="header-right">
          <el-dropdown class="message-dropdown" @command="handleMessageCommand" trigger="click">
            <div class="message-trigger">
              <el-badge :value="unreadCount" :hidden="unreadCount === 0" class="message-badge">
                <el-icon :size="20"><Bell /></el-icon>
              </el-badge>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="inbox">消息中心</el-dropdown-item>
                <el-dropdown-item command="mark-all" divided>全部标为已读</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-dropdown @command="handleCommand">
            <div class="user-info">
              <el-avatar :size="32" class="user-avatar">
                <el-icon :size="18"><UserFilled /></el-icon>
              </el-avatar>
              <span class="user-name">{{ userStore.userInfo?.realName || userStore.userInfo?.username }}</span>
              <el-icon class="user-arrow"><arrow-down /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item command="password">修改密码</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import SubMenu from './components/SubMenu.vue'
import Breadcrumb from '@/components/Breadcrumb.vue'
import { Management, Fold, Expand, ArrowDown, Bell, UserFilled } from '@element-plus/icons-vue'
import { getUnreadCount } from '@/api/message'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()
const unreadCount = ref(0)

const activeMenu = computed(() => route.path)

const fetchUnread = async () => {
  const res: any = await getUnreadCount()
  unreadCount.value = res.data || 0
}

const handleCommand = (command: string) => {
  if (command === 'logout') {
    userStore.logout()
  } else if (command === 'profile') {
    router.push('/profile')
  } else if (command === 'password') {
    router.push('/profile')
  }
}

const handleMessageCommand = async (command: string) => {
  if (command === 'inbox') {
    router.push('/message/list')
  } else if (command === 'mark-all') {
    const res: any = await getUnreadCount()
    if (res.data > 0) {
      fetchUnread()
    }
  }
}

onMounted(() => {
  fetchUnread()
  const timer = setInterval(fetchUnread, 30000)
  onUnmounted(() => clearInterval(timer))
})
</script>

<style scoped lang="scss">
.layout-container {
  height: 100vh;
  background: #f1f5f9;
}

.sidebar {
  position: relative;
  background: linear-gradient(180deg, #0f172a 0%, #1e293b 100%);
  transition: width 0.3s ease;
  overflow: hidden;
  box-shadow: 4px 0 24px rgba(15, 23, 42, 0.2);

  &::after {
    content: '';
    position: absolute;
    top: 0;
    right: 0;
    width: 1px;
    height: 100%;
    background: linear-gradient(180deg, transparent, rgba(255, 255, 255, 0.08), transparent);
  }
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #fff;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);

  .logo-icon {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    background: linear-gradient(135deg, #2563eb 0%, #4f46e5 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 6px 16px rgba(37, 99, 235, 0.35);
    flex-shrink: 0;
  }

  .logo-text {
    font-size: 18px;
    font-weight: 700;
    letter-spacing: 0.5px;
    white-space: nowrap;
  }
}

.menu-scroll {
  height: calc(100vh - 64px);
  overflow-y: auto;

  &::-webkit-scrollbar {
    width: 4px;
  }

  &::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.15);
    border-radius: 2px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }
}

.menu-scroll :deep(.el-menu) {
  border-right: none;
  padding: 12px 10px;
}

.main-wrapper {
  position: relative;
  z-index: 1;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 20px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid #e2e8f0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #475569;
  transition: all 0.25s ease;

  &:hover {
    background: #f1f5f9;
    color: #2563eb;
  }
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.message-trigger {
  width: 38px;
  height: 38px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #475569;
  transition: all 0.25s ease;

  &:hover {
    background: #f1f5f9;
    color: #2563eb;
  }
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 10px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.25s ease;

  &:hover {
    background: #f1f5f9;
  }

  .user-avatar {
    background: linear-gradient(135deg, #2563eb 0%, #4f46e5 100%);
    color: #fff;
    flex-shrink: 0;
  }

  .user-name {
    font-size: 14px;
    color: #1e293b;
    font-weight: 500;
  }

  .user-arrow {
    font-size: 12px;
    color: #94a3b8;
  }
}

.main {
  padding: 20px;
  overflow: auto;
}
</style>
