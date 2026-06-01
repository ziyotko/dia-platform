<template>
  <el-container class="layout-container">
    <el-aside
      :width="appStore.sidebarCollapsed ? '64px' : '220px'"
      class="sidebar"
    >
      <div class="logo">
        <el-icon size="28" color="#409eff"><Platform /></el-icon>
        <span v-show="!appStore.sidebarCollapsed" class="logo-text">管理后台</span>
      </div>
      <el-scrollbar class="menu-scrollbar">
        <el-menu
          :default-active="route.path"
          :collapse="appStore.sidebarCollapsed"
          :collapse-transition="false"
          router
          class="sidebar-menu"
          background-color="transparent"
          text-color="#2c3e50"
          active-text-color="#409eff"
        >
          <el-menu-item index="/dashboard">
            <el-icon><HomeFilled /></el-icon>
            <template #title>欢迎首页</template>
          </el-menu-item>

          <el-sub-menu index="/system">
            <template #title>
              <el-icon><Tools /></el-icon>
              <span>系统管理</span>
            </template>
            <el-menu-item index="/users">
              <el-icon><UserFilled /></el-icon>
              <template #title>用户管理</template>
            </el-menu-item>
            <el-menu-item index="/roles">
              <el-icon><Avatar /></el-icon>
              <template #title>角色管理</template>
            </el-menu-item>
            <el-menu-item index="/menus">
              <el-icon><Menu /></el-icon>
              <template #title>菜单管理</template>
            </el-menu-item>
            <el-menu-item index="/logs">
              <el-icon><List /></el-icon>
              <template #title>操作日志</template>
            </el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="/personal">
            <template #title>
              <el-icon><User /></el-icon>
              <span>个人中心</span>
            </template>
            <el-menu-item index="/profile">
              <el-icon><Document /></el-icon>
              <template #title>个人信息</template>
            </el-menu-item>
            <el-menu-item index="/settings">
              <el-icon><Setting /></el-icon>
              <template #title>系统设置</template>
            </el-menu-item>
          </el-sub-menu>
        </el-menu>
      </el-scrollbar>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-icon
            class="collapse-btn"
            size="20"
            @click="appStore.toggleSidebar"
          >
            <Fold v-if="!appStore.sidebarCollapsed" />
            <Expand v-else />
          </el-icon>
          <breadcrumb />
        </div>
        <div class="header-right">
          <el-tooltip content="全屏" placement="bottom">
            <el-icon class="header-icon" size="18" @click="toggleFullScreen">
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
                <el-dropdown-item command="settings">
                  <el-icon><Setting /></el-icon>系统设置
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

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
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Platform,
  HomeFilled,
  Tools,
  UserFilled,
  Avatar,
  Menu,
  List,
  User,
  Document,
  Setting,
  Fold,
  Expand,
  FullScreen,
  ArrowDown,
  SwitchButton
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import Breadcrumb from '@/components/Breadcrumb.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()

const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

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
    case 'settings':
      router.push('/settings')
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
  transition: width 0.3s;
  display: flex;
  flex-direction: column;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-bottom: 1px solid #d9ecff;
  padding: 0 16px;

  .logo-text {
    font-size: 18px;
    font-weight: 600;
    color: #2c3e50;
    white-space: nowrap;
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

    &:hover {
      background: rgba(64, 158, 255, 0.08) !important;
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