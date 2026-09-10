<template>
  <el-container class="member-layout">
    <el-aside :width="collapsed ? '64px' : '160px'" class="sidebar">
      <div class="sidebar-header" @click="$router.push('/member/dashboard')">
        <el-icon :size="24"><OfficeBuilding /></el-icon>
        <span v-show="!collapsed" class="title">会员中心</span>
      </div>
      <el-menu
        :default-active="route.path"
        :collapse="collapsed"
        :collapse-transition="false"
        router
        background-color="transparent"
        text-color="#344054"
        active-text-color="#002fa7"
        class="sidebar-menu"
      >
        <el-menu-item index="/member/dashboard">
          <el-icon><HomeFilled /></el-icon>
          <span>会员首页</span>
        </el-menu-item>
        <el-menu-item index="/member/profile">
          <el-icon><User /></el-icon>
          <span>我的资料</span>
        </el-menu-item>
        <el-menu-item index="/member/applications">
          <el-icon><Document /></el-icon>
          <span>我的申请</span>
        </el-menu-item>
        <el-menu-item index="/member/fees">
          <el-icon><Money /></el-icon>
          <span>会费管理</span>
        </el-menu-item>
        <el-menu-item index="/member/certificates">
          <el-icon><Medal /></el-icon>
          <span>我的证书</span>
        </el-menu-item>
        <el-menu-item index="/member/organizations">
          <el-icon><Connection /></el-icon>
          <span>加入信息</span>
        </el-menu-item>
        <el-menu-item index="/member/articles">
          <el-icon><EditPen /></el-icon>
          <span>我的文章</span>
        </el-menu-item>
        <el-menu-item index="/member/messages">
          <el-icon><ChatDotRound /></el-icon>
          <span>会员留言</span>
        </el-menu-item>
        <el-menu-item index="/member/service">
          <el-icon><Reading /></el-icon>
          <span>服务中心</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="topbar">
        <div class="topbar-left">
          <el-icon class="collapse-btn" @click="appStore.toggleCollapse()" :size="20">
            <Fold v-if="!collapsed" /><Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item v-if="route.meta?.title">{{ route.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="topbar-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="32" icon="UserFilled" />
              <span class="username">{{ userStore.userInfo?.username }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人资料</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import { useSiteStore } from '@/stores/site'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()
const siteStore = useSiteStore()
const collapsed = computed(() => appStore.collapsed)

function handleCommand(cmd: string) {
  if (cmd === 'profile') router.push('/member/profile')
  else if (cmd === 'admin') router.push('/admin/dashboard')
  else if (cmd === 'logout') {
    userStore.logout()
    ElMessage.success('已退出登录')
  }
}
</script>

<style scoped lang="scss">
.member-layout { height: 100vh; }
.sidebar {
  background: #fff;
  border-right: 1px solid var(--app-border);
  box-shadow: 2px 0 12px rgba(16, 24, 40, 0.04);
  overflow: hidden;
  transition: width 0.3s;
  .sidebar-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 20px;
    color: var(--app-text-primary);
    cursor: pointer;
    .title { flex: 1; min-width: 0; font-size: 16px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  }
  .sidebar-menu {
    border-right: none;
    &.el-menu--collapse {
      :deep(.el-menu-item) {
        margin: 2px 0;
        padding: 0;
        justify-content: center;
        .el-icon { margin: 0; }
      }
    }
    :deep(.el-menu-item) {
      border-radius: 8px;
      margin: 2px 8px;
      &:hover { background: var(--el-color-primary-light-9); }
      &.is-active {
        background: var(--el-color-primary-light-9);
        color: var(--el-color-primary);
        font-weight: 600;
      }
    }
  }
}
.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  border-bottom: 1px solid var(--app-border);
  padding: 0 24px;
  height: 56px;
  .topbar-left { display: flex; align-items: center; gap: 16px; }
  .collapse-btn { cursor: pointer; &:hover { color: var(--el-color-primary); } }
  .user-info {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    user-select: none;
    -webkit-user-select: none;
    outline: none;
    &:focus, &:focus-visible { outline: none; }
    .username { font-size: 14px; color: var(--app-text-secondary); }
  }
  :deep(.el-dropdown) {
    &:focus, &:focus-visible, &:focus-within { outline: none; }
  }
}
.main-content {
  background: var(--app-bg);
  padding: 24px;
}
</style>
