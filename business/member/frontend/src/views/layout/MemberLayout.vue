<template>
  <el-container class="member-layout">
    <el-aside :width="collapsed ? '64px' : '220px'" class="sidebar">
      <div class="sidebar-header" @click="$router.push('/dashboard')">
        <el-icon :size="24"><OfficeBuilding /></el-icon>
        <span v-show="!collapsed" class="title">会员中心</span>
      </div>
      <el-menu
        :default-active="route.path"
        :collapse="collapsed"
        :collapse-transition="false"
        router
        background-color="#1f2937"
        text-color="#9ca3af"
        active-text-color="#fff"
      >
        <el-menu-item index="/dashboard">
          <el-icon><HomeFilled /></el-icon>
          <span>会员首页</span>
        </el-menu-item>
        <el-menu-item index="/profile">
          <el-icon><User /></el-icon>
          <span>我的资料</span>
        </el-menu-item>
        <el-menu-item index="/applications">
          <el-icon><Document /></el-icon>
          <span>我的申请</span>
        </el-menu-item>
        <el-menu-item index="/fees">
          <el-icon><Money /></el-icon>
          <span>会费管理</span>
        </el-menu-item>
        <el-menu-item index="/certificates">
          <el-icon><Medal /></el-icon>
          <span>我的证书</span>
        </el-menu-item>
        <el-menu-item index="/organizations">
          <el-icon><Connection /></el-icon>
          <span>参加的组织</span>
        </el-menu-item>
        <el-menu-item index="/articles">
          <el-icon><EditPen /></el-icon>
          <span>我的文章</span>
        </el-menu-item>
        <el-menu-item index="/messages">
          <el-icon><ChatDotRound /></el-icon>
          <span>会员留言</span>
        </el-menu-item>
        <el-menu-item index="/service">
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
            <el-breadcrumb-item>会员中心</el-breadcrumb-item>
            <el-breadcrumb-item>{{ route.meta.title || route.name }}</el-breadcrumb-item>
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
                <el-dropdown-item v-if="userStore.isAdmin" command="admin">管理后台</el-dropdown-item>
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
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()
const collapsed = computed(() => appStore.collapsed)

function handleCommand(cmd: string) {
  if (cmd === 'profile') router.push('/profile')
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
  background: #1f2937;
  overflow: hidden;
  transition: width 0.3s;
  .sidebar-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 20px;
    color: #fff;
    cursor: pointer;
    .title { font-size: 16px; font-weight: 600; white-space: nowrap; }
  }
}
.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
  padding: 0 24px;
  height: 56px;
  .topbar-left { display: flex; align-items: center; gap: 16px; }
  .collapse-btn { cursor: pointer; &:hover { color: #1a6fb5; } }
  .user-info {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    .username { font-size: 14px; color: #374151; }
  }
}
.main-content {
  background: #f5f7fa;
  padding: 24px;
}
</style>
