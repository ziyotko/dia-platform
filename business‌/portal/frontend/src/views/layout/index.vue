<template>
  <el-container class="layout-container">
    <el-aside
      :width="appStore.sidebarCollapsed ? '64px' : '220px'"
      class="sidebar"
      :class="{ dark: appStore.sidebarStyle === 'dark' }"
    >
      <div class="logo">
        <div class="logo-icon-wrap">
          <el-icon size="24" :color="appStore.themeColor"><Platform /></el-icon>
        </div>
        <span class="logo-text">门户管理后台</span>
      </div>
      <el-scrollbar class="menu-scrollbar">
        <el-menu
          :default-active="route.path"
          :collapse="appStore.sidebarCollapsed"
          :collapse-transition="false"
          :unique-opened="true"
          class="sidebar-menu"
          background-color="transparent"
          :text-color="appStore.sidebarStyle === 'dark' ? '#bfcbd9' : '#2c3e50'"
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
          <el-icon
            class="collapse-btn"
            size="36"
            @click="appStore.toggleSidebar"
          >
            <Fold v-if="!appStore.sidebarCollapsed" />
            <Expand v-else />
          </el-icon>
          <breadcrumb v-if="appStore.breadcrumb" />
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
              <span class="username">{{ userStore.userInfo?.nickname || userStore.userInfo?.username }}</span>
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

      <tags-view v-if="appStore.tagsView" />

      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
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
  Platform,
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
import { getPublicSiteInfo, getSettings } from '@/api/settings'
import type { Settings } from '@/api/settings'
import Breadcrumb from '@/components/Breadcrumb.vue'
import TagsView from '@/components/TagsView.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()
const menuList = computed(() => userStore.menuList)
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

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

const resolveLogoUrl = (url: string) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return `${window.location.origin}${url}`
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
    const res: any = await getSettings()
    const data = res.data as Settings
    if (data) {
      appStore.setThemeSettings({
        themeColor: data.themeColor || '#409eff',
        sidebarStyle: (data.sidebarStyle || 'light') as 'light' | 'dark',
        tagsView: data.tagsView ?? true,
        breadcrumb: data.breadcrumb ?? true
      })
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
      }).then(() => {
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
  background: #f5f9ff;
}

.sidebar {
  background: linear-gradient(180deg, #f0f7ff 0%, #e6f2ff 100%);
  border-right: 1px solid #d9ecff;
  transition: width 0.3s, background 0.3s;
  display: flex;
  flex-direction: column;

  &.dark {
    background: linear-gradient(180deg, #1a1a2e 0%, #16213e 100%);
    border-right-color: #0f3460;

    .logo {
      border-bottom-color: #0f3460;

      .logo-icon-wrap {
        background: linear-gradient(135deg, rgba(64, 158, 255, 0.25) 0%, rgba(64, 158, 255, 0.15) 100%);
        box-shadow: 0 2px 8px rgba(64, 158, 255, 0.25);
      }

      .logo-text {
        background: linear-gradient(90deg, #ffffff 0%, #bfcbd9 100%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
      }
    }

    .sidebar-menu {
      :deep(.el-menu-item) {
        color: #bfcbd9;

        &:hover {
          background: rgba(255, 255, 255, 0.05) !important;
          color: #fff;
        }

        &.is-active {
          background: rgba(64, 158, 255, 0.2) !important;
          color: var(--el-color-primary);
        }
      }

      :deep(.el-sub-menu__title) {
        color: #bfcbd9;

        &:hover {
          background: rgba(255, 255, 255, 0.05) !important;
          color: #fff;
        }
      }
    }
  }
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  border-bottom: 1px solid #d9ecff;
  padding: 0 16px;
  transition: border-color 0.3s;

  .logo-icon-wrap {
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 10px;
    background: linear-gradient(135deg, var(--el-color-primary-light-8) 0%, var(--el-color-primary-light-9) 100%);
    box-shadow: 0 2px 8px rgba(64, 158, 255, 0.15);
    transition: transform 0.3s;

    &:hover {
      transform: scale(1.05);
    }
  }

  .logo-text {
    font-size: 20px;
    font-weight: 700;
    color: #2c3e50;
    white-space: nowrap;
    transition: color 0.3s;
    letter-spacing: 1px;
    background: linear-gradient(90deg, #2c3e50 0%, #4a6582 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }
}

.menu-scrollbar {
  flex: 1;
}

.sidebar-menu {
  border-right: none;
  background: transparent !important;

  :deep(.el-menu-item) {
    margin: 4px 12px;
    border-radius: 8px;
    transition: background 0.3s, color 0.3s;

    &:hover {
      background: rgba(64, 158, 255, 0.08) !important;
    }

    &.is-active {
      background: rgba(64, 158, 255, 0.12) !important;
      font-weight: 600;
    }
  }

  :deep(.el-sub-menu__title) {
    margin: 4px 12px;
    border-radius: 8px;
    transition: background 0.3s, color 0.3s;

    &:hover {
      background: rgba(64, 158, 255, 0.08) !important;
    }
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
  background: #fff;
  border-bottom: 1px solid #d9ecff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  cursor: pointer;
  color: #606266;
  padding: 6px;
  border-radius: 4px;
  transition: all 0.3s;

  &:hover {
    background: #f5f9ff;
    color: #409eff;
  }
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.header-icon {
  cursor: pointer;
  color: #606266;
  padding: 6px;
  border-radius: 4px;
  transition: all 0.3s;

  &:hover {
    background: #f5f9ff;
    color: #409eff;
  }
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 8px;
  transition: all 0.3s;

  &:hover {
    background: #f5f9ff;
  }

  .username {
    font-size: 14px;
    color: #2c3e50;
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.main-content {
  padding: 20px;
  background: #f5f9ff;
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