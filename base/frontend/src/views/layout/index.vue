<template>
  <el-container class="layout-container">
    <el-aside width="64px" class="sidebar">
      <div class="logo">
        <div class="logo-icon">
          <el-icon :size="26" color="#fff"><Management /></el-icon>
        </div>
      </div>
      <div class="menu-icons">
        <div
          v-for="menu in visibleTopMenus"
          :key="menu.id"
          class="menu-icon-item"
          :class="{ active: isTopActive(menu), 'popup-open': popupMenu?.id === menu.id }"
          @mouseenter="handleIconEnter(menu)"
          @mouseleave="handleIconLeave"
          @click="handleIconClick(menu)"
        >
          <el-tooltip :content="menu.name" placement="right" :disabled="popupMenu?.id === menu.id">
            <el-icon :size="20"><component :is="menu.icon || 'Menu'" /></el-icon>
          </el-tooltip>
        </div>
      </div>
    </el-aside>

    <!-- 弹出式菜单面板 -->
    <transition name="popup-slide">
      <div
        v-if="popupMenu"
        class="menu-popup"
        @mouseenter="handlePopupEnter"
        @mouseleave="handlePopupLeave"
      >
        <div class="popup-header">
          <el-icon :size="18"><component :is="popupMenu.icon || 'Menu'" /></el-icon>
          <span class="popup-title">{{ popupMenu.name }}</span>
        </div>
        <el-scrollbar class="popup-scroll">
          <el-menu
            :default-active="activeMenu"
            router
            background-color="transparent"
            text-color="#475569"
            active-text-color="#fff"
          >
            <template v-if="popupMenu.children?.length">
              <sub-menu v-for="child in popupMenu.children" :key="child.id" :menu="child" />
            </template>
            <el-menu-item v-else :index="popupMenu.path">
              <span>{{ popupMenu.name }}</span>
            </el-menu-item>
          </el-menu>
        </el-scrollbar>
      </div>
    </transition>

    <el-container class="main-wrapper">
      <el-header class="header">
        <div class="header-left">
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
                <el-dropdown-item v-if="userStore.can('base:message:read-all')" command="mark-all" divided>全部标为已读</el-dropdown-item>
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
        <el-alert
          v-if="userStore.menus.length === 0"
          title="当前账号未分配任何菜单权限"
          description="请联系管理员在「系统管理 → 角色管理」中为该账号分配角色与菜单后再使用。"
          type="warning"
          :closable="false"
          show-icon
          class="empty-menu-alert"
        />
        <router-view v-slot="{ Component }">
          <keep-alive :include="cachedViewNames">
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import type { Menu } from '@/api/menu'
import SubMenu from './components/SubMenu.vue'
import Breadcrumb from '@/components/Breadcrumb.vue'
import { Management, ArrowDown, Bell, UserFilled } from '@element-plus/icons-vue'
import { getUnreadCount, markAllMessageRead } from '@/api/message'
import { ElMessage, ElMessageBox } from 'element-plus'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const unreadCount = ref(0)
const popupMenu = ref<Menu | null>(null)
let showTimer: number | null = null
let hideTimer: number | null = null
let timer: number | null = null

const activeMenu = computed(() => route.path)

const activeTopMenu = computed(() => {
  const path = route.path
  for (const menu of userStore.menus) {
    if (isPathInMenu(path, menu)) return menu
  }
  return null
})

function isMenuPathMatch(routePath: string, menuPath: string): boolean {
  if (!menuPath) return false
  if (routePath === menuPath) return true
  return routePath.startsWith(menuPath + '/')
}

function isPathInMenu(path: string, menu: Menu): boolean {
  if (isMenuPathMatch(path, menu.path)) return true
  return menu.children?.some((child) => isPathInMenu(path, child)) || false
}

const isTopActive = (menu: Menu) => activeTopMenu.value?.id === menu.id

const showPopup = (menu: Menu) => {
  popupMenu.value = menu
}

const hidePopup = () => {
  popupMenu.value = null
}

const handleIconEnter = (menu: Menu) => {
  if (hideTimer) {
    clearTimeout(hideTimer)
    hideTimer = null
  }
  showTimer = window.setTimeout(() => {
    showPopup(menu)
  }, 120)
}

const handleIconLeave = () => {
  if (showTimer) {
    clearTimeout(showTimer)
    showTimer = null
  }
  hideTimer = window.setTimeout(() => {
    hidePopup()
  }, 180)
}

const handlePopupEnter = () => {
  if (hideTimer) {
    clearTimeout(hideTimer)
    hideTimer = null
  }
}

const handlePopupLeave = () => {
  hideTimer = window.setTimeout(() => {
    hidePopup()
  }, 180)
}

const handleIconClick = (menu: Menu) => {
  if (!menu.children?.length && menu.path) {
    router.push(menu.path)
    hidePopup()
    return
  }
  showPopup(menu)
}

watch(
  () => route.path,
  () => {
    hidePopup()
  }
)

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
    if (unreadCount.value === 0) {
      ElMessage.info('没有未读消息')
      return
    }
    await ElMessageBox.confirm('确认将全部未读消息标为已读？', '提示', { type: 'warning' })
    const res: any = await markAllMessageRead()
    ElMessage.success(`已标记 ${res.data?.count ?? 0} 条为已读`)
    fetchUnread()
  }
}

// 侧边栏顶部图标不展示隐藏菜单与按钮类型菜单
const visibleTopMenus = computed(() => userStore.menus.filter((m) => !m.hidden && m.type !== 'button'))

// 需要缓存的页面：菜单上勾选了「页面缓存」的菜单项（组件名 = 菜单名，见 router/index.ts）
const cachedViewNames = computed(() => {
  const names: string[] = []
  const walk = (list: Menu[]) => {
    list.forEach((m) => {
      if (m.type === 'menu' && m.keepAlive && m.target !== 'iframe') names.push(m.name)
      if (m.children?.length) walk(m.children)
    })
  }
  walk(userStore.menus)
  return names
})

onMounted(() => {
  fetchUnread()
  timer = window.setInterval(fetchUnread, 30000)
  window.addEventListener('base:unread-changed', fetchUnread)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  if (showTimer) clearTimeout(showTimer)
  if (hideTimer) clearTimeout(hideTimer)
  window.removeEventListener('base:unread-changed', fetchUnread)
})
</script>

<style scoped lang="scss">
.empty-menu-alert {
  margin-bottom: 12px;
}

.layout-container {
  height: 100vh;
  background: #f1f5f9;
}

.sidebar {
  position: relative;
  z-index: 100;
  background: linear-gradient(180deg, #f0f9ff 0%, #e0f2fe 100%);
  overflow: hidden;
  box-shadow: 4px 0 24px rgba(0, 0, 0, 0.06);

  &::after {
    content: '';
    position: absolute;
    top: 0;
    right: 0;
    width: 1px;
    height: 100%;
    background: linear-gradient(180deg, transparent, rgba(0, 0, 0, 0.06), transparent);
  }
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #1e293b;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);

  .logo-icon {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 6px 16px rgba(37, 99, 235, 0.3);
    flex-shrink: 0;
  }
}

.menu-icons {
  padding: 12px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.menu-icon-item {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
  cursor: pointer;
  transition: all 0.25s ease;
  position: relative;

  &:hover,
  &.popup-open {
    background: rgba(59, 130, 246, 0.08);
    color: #3b82f6;
  }

  &.active {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: #fff;
    box-shadow: 0 6px 16px rgba(37, 99, 235, 0.35);
  }

  &.active::before {
    content: '';
    position: absolute;
    left: -10px;
    top: 50%;
    transform: translateY(-50%);
    width: 3px;
    height: 18px;
    border-radius: 0 3px 3px 0;
    background: #3b82f6;
  }
}

.menu-popup {
  position: fixed;
  top: 0;
  left: 64px;
  width: 220px;
  height: 100vh;
  z-index: 99;
  background: #fff;
  border-right: 1px solid #e2e8f0;
  box-shadow: 8px 0 32px rgba(0, 0, 0, 0.06);
  display: flex;
  flex-direction: column;
}

.popup-header {
  height: 64px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 18px;
  color: #1e293b;
  border-bottom: 1px solid #e2e8f0;
  flex-shrink: 0;

  .popup-title {
    font-size: 16px;
    font-weight: 600;
  }
}

.popup-scroll {
  flex: 1;
  overflow-y: auto;

  &::-webkit-scrollbar {
    width: 4px;
  }

  &::-webkit-scrollbar-thumb {
    background: #cbd5e1;
    border-radius: 2px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }
}

.popup-scroll :deep(.el-menu) {
  border-right: none;
  padding: 12px 10px;
}

.popup-scroll :deep(.el-menu-item),
.popup-scroll :deep(.el-sub-menu__title) {
  color: #475569 !important;
  border-radius: 8px;
  height: 46px;
  line-height: 46px;
  margin-bottom: 4px;
  transition: all 0.25s ease;

  &:hover {
    background: rgba(59, 130, 246, 0.06) !important;
    color: #2563eb !important;
  }
}

.popup-scroll :deep(.el-menu-item.is-active) {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%) !important;
  color: #fff !important;
  box-shadow: 0 6px 16px rgba(37, 99, 235, 0.35);
}

.popup-scroll :deep(.el-sub-menu.is-active > .el-sub-menu__title) {
  color: #2563eb !important;
}

.popup-slide-enter-active,
.popup-slide-leave-active {
  transition: all 0.25s ease;
}

.popup-slide-enter-from,
.popup-slide-leave-to {
  opacity: 0;
  transform: translateX(-12px);
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
