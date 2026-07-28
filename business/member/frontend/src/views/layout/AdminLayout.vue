<template>
  <el-container class="admin-layout">
    <el-aside :width="collapsed ? '80px' : '200px'" class="sidebar">
      <div class="sidebar-header" @click="$router.push('/admin/dashboard')">
        <el-icon :size="24" color="#3b82f6"><Setting /></el-icon>
        <span v-show="!collapsed" class="title">会员管理后台</span>
      </div>
      <el-menu
        router
        :collapse="collapsed"
        :collapse-transition="false"
        :default-active="route.path"
        background-color="#fff"
        text-color="#4b5563"
        active-text-color="#3b82f6"
      >
        <el-menu-item index="/admin/dashboard"><el-icon><DataAnalysis /></el-icon><span>控制台</span></el-menu-item>
        <el-menu-item index="/admin/members"><el-icon><UserFilled /></el-icon><span>会员管理</span></el-menu-item>
        <el-menu-item index="/admin/applications"><el-icon><DocumentChecked /></el-icon><span>入会审核</span></el-menu-item>
        <el-menu-item index="/admin/fees"><el-icon><Money /></el-icon><span>会费管理</span></el-menu-item>
        <el-menu-item index="/admin/certificates"><el-icon><Medal /></el-icon><span>证书管理</span></el-menu-item>
        <el-menu-item index="/admin/organizations"><el-icon><Connection /></el-icon><span>组织机构</span></el-menu-item>
        <el-menu-item index="/admin/member-levels"><el-icon><Sort /></el-icon><span>会员等级</span></el-menu-item>
        <el-menu-item index="/admin/fee-standards"><el-icon><Coin /></el-icon><span>会费标准</span></el-menu-item>
        <el-menu-item index="/admin/messages"><el-icon><ChatDotRound /></el-icon><span>会员留言</span></el-menu-item>
        <el-menu-item index="/admin/articles"><el-icon><Document /></el-icon><span>文章管理</span></el-menu-item>
        <el-menu-item index="/admin/announcements"><el-icon><Notification /></el-icon><span>公告管理</span></el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="topbar">
        <div class="topbar-left">
          <el-icon class="collapse-btn" @click="toggleCollapse" :size="20">
            <Fold v-if="!collapsed" /><Expand v-else />
          </el-icon>
        </div>
        <el-dropdown @command="handleCommand">
          <span class="user-info">
            <el-avatar :size="32" :icon="UserFilled" />
            <span>{{ userStore.userInfo?.username }}</span>
          </span>
          <template #dropdown>
            <el-dropdown-item command="member">会员中心</el-dropdown-item>
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
import { UserFilled, Fold, Expand, Coin } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const collapsed = ref(false)

function toggleCollapse() {
  collapsed.value = !collapsed.value
}

function handleCommand(cmd: string) {
  if (cmd === 'member') router.push('/member/dashboard')
  else if (cmd === 'logout') userStore.logout()
}
</script>

<style scoped lang="scss">
.admin-layout { height: 100vh; }

.sidebar {
  background: #fff;
  border-right: 1px solid #f0f0f0;
  overflow: hidden;
  box-shadow: 2px 0 8px rgba(0,0,0,0.02);
  transition: width 0.3s;

  .sidebar-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 20px;
    cursor: pointer;
    border-bottom: 1px solid #f5f5f5;
    .title { font-size: 16px; font-weight: 600; color: #1f2937; white-space: nowrap; }
  }

  :deep(.el-menu) {
    border-right: none;
    .el-menu-item {
      margin: 2px 8px;
      border-radius: 8px;
      white-space: nowrap;
      &:hover { background: #f0f7ff; }
      &.is-active { background: #e8f4fd; font-weight: 600; }
    }
  }
}

.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  padding: 0 24px;
  height: 56px;

  .topbar-left {
    display: flex;
    align-items: center;
  }

  .collapse-btn {
    cursor: pointer;
    color: #6b7280;
    transition: color 0.2s;
    &:hover { color: #3b82f6; }
  }

  .user-info { display: flex; align-items: center; gap: 8px; cursor: pointer; }
}

.main-content {
  background: #f8fafc;
  padding: 24px;
}
</style>
