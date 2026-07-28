<template>
  <el-container class="admin-layout">
    <el-aside width="220px" class="sidebar">
      <div class="sidebar-header" @click="$router.push('/admin/dashboard')">
        <el-icon :size="24"><Setting /></el-icon>
        <span class="title">管理后台</span>
      </div>
      <el-menu router background-color="#1f2937" text-color="#9ca3af" active-text-color="#fff">
        <el-menu-item index="/admin/dashboard"><el-icon><DataAnalysis /></el-icon>控制台</el-menu-item>
        <el-menu-item index="/admin/members"><el-icon><UserFilled /></el-icon>会员管理</el-menu-item>
        <el-menu-item index="/admin/applications"><el-icon><DocumentChecked /></el-icon>入会审核</el-menu-item>
        <el-menu-item index="/admin/fees"><el-icon><Money /></el-icon>会费管理</el-menu-item>
        <el-menu-item index="/admin/certificates"><el-icon><Medal /></el-icon>证书管理</el-menu-item>
        <el-menu-item index="/admin/organizations"><el-icon><Connection /></el-icon>组织机构</el-menu-item>
        <el-menu-item index="/admin/messages"><el-icon><ChatDotRound /></el-icon>会员留言</el-menu-item>
        <el-menu-item index="/admin/articles"><el-icon><Document /></el-icon>文章管理</el-menu-item>
        <el-menu-item index="/admin/announcements"><el-icon><Notification /></el-icon>公告管理</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="topbar">
        <span class="title">管理后台</span>
        <el-dropdown @command="handleCommand">
          <span class="user-info">
            <el-avatar :size="32" icon="UserFilled" />
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
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
const router = useRouter()
const userStore = useUserStore()
function handleCommand(cmd: string) {
  if (cmd === 'member') router.push('/dashboard')
  else if (cmd === 'logout') userStore.logout()
}
</script>

<style scoped lang="scss">
.admin-layout { height: 100vh; }
.sidebar { background: #1f2937; overflow: hidden;
  .sidebar-header { display: flex; align-items: center; gap: 10px; padding: 20px; color: #fff; cursor: pointer;
    .title { font-size: 16px; font-weight: 600; }
  }
}
.topbar { display: flex; justify-content: space-between; align-items: center; background: #fff;
  border-bottom: 1px solid #e5e7eb; padding: 0 24px; height: 56px;
  .title { font-size: 16px; font-weight: 600; }
  .user-info { display: flex; align-items: center; gap: 8px; cursor: pointer; }
}
.main-content { background: #f5f7fa; padding: 24px; }
</style>
