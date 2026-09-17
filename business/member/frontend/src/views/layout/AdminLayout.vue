<template>
  <el-container class="admin-layout">
    <el-aside :width="collapsed ? '64px' : '160px'" class="sidebar">
      <div class="sidebar-header" @click="$router.push('/admin/dashboard')">
        <el-icon :size="24" color="#002fa7"><Setting /></el-icon>
        <span v-show="!collapsed" class="title">管理后台</span>
      </div>
      <el-menu
        router
        :collapse="collapsed"
        :collapse-transition="false"
        :default-active="route.path"
        background-color="transparent"
        text-color="#344054"
        active-text-color="#002fa7"
        class="sidebar-menu"
      >
        <el-menu-item index="/admin/dashboard"><el-icon><DataAnalysis /></el-icon><span>管理首页</span></el-menu-item>
        <el-menu-item index="/admin/applications"><el-icon><DocumentChecked /></el-icon><span>入会审核</span></el-menu-item>
        <el-menu-item index="/admin/members"><el-icon><UserFilled /></el-icon><span>会员管理</span></el-menu-item>
        <el-menu-item index="/admin/fees"><el-icon><Money /></el-icon><span>会费管理</span></el-menu-item>
        <el-menu-item index="/admin/certificates"><el-icon><Medal /></el-icon><span>证书管理</span></el-menu-item>
        <el-menu-item index="/admin/organizations"><el-icon><Connection /></el-icon><span>组织机构</span></el-menu-item>
        <el-menu-item index="/admin/member-levels"><el-icon><Sort /></el-icon><span>会员等级</span></el-menu-item>
        <el-menu-item index="/admin/fee-standards"><el-icon><Coin /></el-icon><span>会费标准</span></el-menu-item>
        <el-menu-item index="/admin/messages"><el-icon><ChatDotRound /></el-icon><span>会员留言</span></el-menu-item>
        <el-menu-item index="/admin/articles"><el-icon><Document /></el-icon><span>文章管理</span></el-menu-item>
        <el-menu-item index="/admin/announcements"><el-icon><Notification /></el-icon><span>公告管理</span></el-menu-item>
        <el-menu-item index="/admin/charter"><el-icon><Notebook /></el-icon><span>协会章程</span></el-menu-item>
        <el-menu-item index="/admin/system-config"><el-icon><Setting /></el-icon><span>系统管理</span></el-menu-item>
        <el-menu-item index="/admin/profile"><el-icon><User /></el-icon><span>我的资料</span></el-menu-item>
        <el-menu-item index="/admin/member-level-changes"><el-icon><Switch /></el-icon><span>会籍记录</span></el-menu-item>
        <el-menu-item index="/admin/profile-changes"><el-icon><EditPen /></el-icon><span>资料记录</span></el-menu-item>
        <el-menu-item index="/admin/operation-logs"><el-icon><Tickets /></el-icon><span>操作日志</span></el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="topbar">
        <div class="topbar-left">
          <el-icon class="collapse-btn" @click="toggleCollapse" :size="20">
            <Fold v-if="!collapsed" /><Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item v-if="route.meta?.title">{{ route.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <el-dropdown @command="handleCommand">
          <span class="user-info">
            <el-avatar :size="32" :icon="UserFilled" />
            <span>{{ userStore.userInfo?.username }}</span>
          </span>
          <template #dropdown>
            <el-dropdown-item command="member">个人资料</el-dropdown-item>
            <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
          </template>
        </el-dropdown>
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { UserFilled, Fold, Expand, Coin, User, Switch, EditPen } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const siteStore = useSiteStore()
const collapsed = ref(false)

function toggleCollapse() {
  collapsed.value = !collapsed.value
}

function handleCommand(cmd: string) {
  if (cmd === 'member') router.push('/admin/profile')
  else if (cmd === 'logout') userStore.logout()
}
</script>

<style scoped lang="scss">
.admin-layout { height: 100vh; }

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
