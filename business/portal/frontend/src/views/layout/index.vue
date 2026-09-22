<template>
  <el-container class="layout-container">
    <el-aside
      :width="appStore.sidebarCollapsed ? '64px' : '220px'"
      class="sidebar"
    >
      <div class="logo">
        <div class="logo-icon-wrap">
          <el-icon size="24" :color="appStore.themeColor"><Connection /></el-icon>
        </div>
        <span v-if="!appStore.sidebarCollapsed" class="logo-text">统一管理后台</span>
      </div>
      <el-scrollbar class="menu-scrollbar">
        <el-menu
          :default-active="route.path"
          :collapse="appStore.sidebarCollapsed"
          :collapse-transition="false"
          :unique-opened="true"
          class="sidebar-menu"
          background-color="transparent"
          text-color="var(--app-text-heading)"
          :active-text-color="appStore.themeColor"
          @select="handleMenuSelect"
        >
          <sidebar-menu-item v-for="menu in menuList" :key="menu.id" :menu="menu" />
        </el-menu>
      </el-scrollbar>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-tooltip :content="appStore.sidebarCollapsed ? '展开菜单' : '折叠菜单'" placement="bottom">
            <el-icon
              class="collapse-btn"
              size="36"
              aria-label="展开或折叠侧边菜单"
              @click="appStore.toggleSidebar"
            >
              <Fold v-if="!appStore.sidebarCollapsed" />
              <Expand v-else />
            </el-icon>
          </el-tooltip>
          <breadcrumb />
        </div>
        <div class="header-right">
          <el-tooltip content="全屏" placement="bottom">
            <el-icon class="header-icon" size="36" @click="toggleFullScreen">
              <FullScreen />
            </el-icon>
          </el-tooltip>
          <el-dropdown @command="handleCommand">
            <div class="user-info">
              <el-avatar :size="32" :src="userStore.userInfo?.avatar || defaultAvatar" />
              <span class="username">{{ userStore.userInfo?.username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>个人中心
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <tags-view />

      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" :key="route.path + '_' + appStore.refreshKey" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { h, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, ElSubMenu, ElMenuItem, ElIcon } from 'element-plus'
import * as Icons from '@element-plus/icons-vue'
import {
  Connection,
  Fold,
  Expand,
  FullScreen,
  ArrowDown,
  SwitchButton,
  User,
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import type { MenuItem } from '@/api/menus'
import { getPublicSiteInfo, getMinPasswordLengthSettings } from '@/api/settings'
import { logout as logoutApi } from '@/api/auth'
import type { Settings } from '@/api/settings'
import { resolveAssetUrl as resolveLogoUrl } from '@/utils/asset'
import Breadcrumb from '@/components/Breadcrumb.vue'
import TagsView from '@/components/TagsView.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()
const menuList = computed(() => userStore.menuList)
const defaultAvatar = new URL('../../assets/avatar-default.svg', import.meta.url).href

const SidebarMenuItem = {
  name: 'SidebarMenuItem',
  props: { menu: { type: Object, required: true } },
  setup(props: { menu: MenuItem }) {
    return () => {
      const menu = props.menu
      const isExternal = /^https?:\/\//.test(menu.path)
      const indexPath = isExternal ? menu.path : (menu.path.startsWith('/') ? menu.path : '/' + menu.path)
      const iconComp = menu.icon ? (Icons as Record<string, any>)[menu.icon] : null
      if (menu.type === 'directory' && menu.children && menu.children.length > 0) {
        return h(ElSubMenu, { index: indexPath }, {
          title: () => [
            iconComp ? h(ElIcon, null, () => h(iconComp)) : null,
            h('span', null, menu.name)
          ],
          default: () => menu.children!.map((child) => h(SidebarMenuItem, { menu: child }))
        })
      }
      return h(ElMenuItem, { index: indexPath }, {
        default: () => iconComp ? h(ElIcon, null, () => h(iconComp)) : null,
        title: () => menu.name
      })
    }
  }
}

const loadSiteInfo = async () => {
  try {
    const res: any = await getPublicSiteInfo()
    if (res.data) {
      document.title = res.data.siteName || '门户网站管理后台'
      const favicon = document.querySelector('link[rel="icon"]') as HTMLLinkElement | null
      if (favicon && res.data.logo) {
        favicon.href = resolveLogoUrl(res.data.logo)
      }
    } 
  } catch {
    // 使用默认值
  }
}

const loadSettings = async () => {
  try {
    const res: any = await getMinPasswordLengthSettings()
    const data = res.data as Settings
    if (data) {
      appStore.setSecuritySettings({
        minPasswordLength: data.minPasswordLength ?? 8
      })
    }
  } catch {
    // 使用本地缓存或默认值
  }
}

onMounted(() => {
  loadSiteInfo()
  loadSettings()
})

const handleMenuSelect = (index: string) => {
  if (/^https?:\/\//.test(index)) {
    window.open(index, '_blank')
    return
  }
  if (index === route.path) {
    appStore.triggerRefresh()
    return
  }
  router.push(index)
}

const toggleFullScreen = () => {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen()
  } else {
    document.exitFullscreen()
  }
}

const handleCommand = (command: string) => {
  switch (command) {
    case 'profile':
      router.push('/profile')
      break
    case 'logout':
      ElMessageBox.confirm('确定要退出登录吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        // 先通知后端将当前 Token 加入黑名单，避免登出后旧 Token 仍可使用；失败也继续清理本地会话
        try {
          await logoutApi()
        } catch {
          // 忽略：网络异常或 Token 已失效时仍需完成本地登出
        }
        userStore.logout()
        ElMessage.success('已退出登录')
        router.push('/login')
      })
      break
  }
}
</script>

<style scoped lang="scss">
.layout-container {
  min-height: 100vh;
  background: var(--app-bg);
}

.sidebar {
  background: #fff;
  border-right: 1px solid var(--app-border);
  box-shadow: 2px 0 12px rgba(16, 24, 40, 0.04);
  transition: width 0.3s;
  display: flex;
  flex-direction: column;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  border-bottom: 1px solid var(--app-border);
  padding: 0 16px;
  transition: border-color 0.3s;

  .logo-icon-wrap {
    flex-shrink: 0;
    width: 38px;
    height: 38px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 12px;
    background: linear-gradient(135deg, var(--el-color-primary-light-8) 0%, var(--el-color-primary-light-9) 100%);
    box-shadow: 0 4px 10px rgba(0, 47, 167, 0.2);
    transition: transform 0.3s;

    :deep(.el-icon) {
      color: var(--el-color-primary);
    }

    &:hover {
      transform: scale(1.06);
    }
  }

  .logo-text {
    font-size: 18px;
    font-weight: 700;
    color: var(--app-text-primary);
    white-space: nowrap;
    transition: color 0.3s, opacity 0.2s;
    letter-spacing: 0.5px;
    background: linear-gradient(90deg, #1d2739 0%, var(--el-color-primary) 100%);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
}

.menu-scrollbar {
  flex: 1;
}

.sidebar-menu {
  border-right: none;
  background: transparent !important;
  padding: 8px 0;

  :deep(.el-menu-item) {
    position: relative;
    margin: 4px 12px;
    height: 44px;
    line-height: 44px;
    border-radius: 10px;
    transition: background 0.3s, color 0.3s;

    &:hover {
      background: rgba(0, 47, 167, 0.06) !important;
    }

    &.is-active {
      background: linear-gradient(90deg, rgba(0, 47, 167, 0.14) 0%, rgba(0, 47, 167, 0.06) 100%) !important;
      color: var(--el-color-primary);
      font-weight: 600;

      &::before {
        content: '';
        position: absolute;
        left: 0;
        top: 50%;
        transform: translateY(-50%);
        width: 3px;
        height: 20px;
        border-radius: 2px;
        background: var(--el-color-primary);
      }
    }
  }

  :deep(.el-sub-menu__title) {
    margin: 4px 12px;
    height: 44px;
    line-height: 44px;
    border-radius: 10px;
    transition: background 0.3s, color 0.3s;

    &:hover {
      background: rgba(0, 47, 167, 0.06) !important;
    }
  }

  :deep(.el-menu--inline .el-menu-item) {
    margin-left: 24px;
    margin-right: 16px;
    min-width: 0;
  }

  &.el-menu--collapse {
    :deep(.el-menu-item),
    :deep(.el-sub-menu__title) {
      margin-left: 0;
      margin-right: 0;
    }
  }
}

.header {
  height: 64px;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid var(--app-border);
  box-shadow: 0 1px 4px rgba(16, 24, 40, 0.04);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  cursor: pointer;
  color: var(--app-text-secondary);
  padding: 6px;
  border-radius: 8px;
  transition: all 0.3s;

  &:hover {
    background: #f5f7fb;
    color: var(--el-color-primary);
  }
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  position: relative;
  cursor: pointer;
  color: var(--app-text-secondary);
  padding: 8px;
  border-radius: 10px;
  transition: all 0.3s;

  &:hover {
    background: #f5f7fb;
    color: var(--el-color-primary);
  }

  &.has-dot::after {
    content: '';
    position: absolute;
    top: 8px;
    right: 8px;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #f56c6c;
    border: 2px solid #fff;
  }
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  padding: 5px 10px;
  border-radius: 10px;
  transition: all 0.3s;

  &:hover {
    background: #f5f7fb;
  }

  .username {
    font-size: 14px;
    font-weight: 500;
    color: var(--app-text-primary);
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.main-content {
  padding: 20px;
  background: var(--app-bg);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>