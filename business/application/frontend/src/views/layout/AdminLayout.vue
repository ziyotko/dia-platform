<template>
  <el-container style="height:100vh">
    <el-aside :width="collapsed ? '64px' : '220px'" class="sidebar">
      <div class="logo-area" @click="$router.push('/admin/dashboard')">
        <span v-if="!collapsed">评审管理后台</span>
        <span v-else>评</span>
      </div>
      <el-menu :default-active="route.path" router :collapse="collapsed" background-color="#fff" text-color="#374151" active-text-color="#002fa7">
        <el-menu-item v-for="item in menus" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="topbar">
        <div class="topbar-left">
          <el-icon class="collapse-btn" @click="collapsed = !collapsed"><Fold /></el-icon>
        </div>
        <div class="topbar-right">
          <el-dropdown>
            <span class="user-info">{{ adminStore.adminInfo?.realName || '管理员' }} <el-icon><ArrowDown /></el-icon></span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="$router.push('/admin/profile')">个人资料</el-dropdown-item>
                <el-dropdown-item divided @click="adminStore.logout()">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main-content"><router-view /></el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAdminStore } from '@/stores/admin'

const route = useRoute()
const adminStore = useAdminStore()
const collapsed = ref(false)

onMounted(() => {
  if (!adminStore.adminInfo) adminStore.fetchAdminInfo()
})

const fullMenus = [
  { path: '/admin/dashboard', label: '管理看板', icon: 'DataAnalysis' },
  { path: '/admin/categories', label: '类别管理', icon: 'FolderOpened' },
  { path: '/admin/experts', label: '专家库', icon: 'School' },
  { path: '/admin/batches', label: '批次管理', icon: 'Files' },
  { path: '/admin/applications', label: '申报管理', icon: 'Tickets' },
  { path: '/admin/reviews', label: '评审管理', icon: 'Histogram' },
  { path: '/admin/review-tasks', label: '评审任务', icon: 'DocumentChecked' },
  { path: '/admin/announcements', label: '结果公示', icon: 'Bell' },
  { path: '/admin/certificates', label: '证书管理', icon: 'Medal' },
  { path: '/admin/notifications', label: '通知管理', icon: 'Message' },
  { path: '/admin/users', label: '申报人管理', icon: 'UserFilled' },
  { path: '/admin/admins', label: '账号管理', icon: 'Avatar' },
  { path: '/admin/audit', label: '系统日志', icon: 'List' },
  { path: '/admin/profile', label: '个人资料', icon: 'User' },
]

const reviewerMenus = [
  { path: '/admin/dashboard', label: '管理看板', icon: 'DataAnalysis' },
  { path: '/admin/reviews', label: '我的评审', icon: 'Histogram' },
  { path: '/admin/profile', label: '个人资料', icon: 'User' },
]

const menus = computed(() => (adminStore.roleCode === 'reviewer' ? reviewerMenus : fullMenus))
</script>

<style scoped>
.sidebar { background: #fff; border-right: 1px solid var(--app-border); overflow: hidden; transition: width .3s; }
.logo-area { height: 60px; display: flex; align-items: center; justify-content: center; color: #002fa7; font-size: 18px; font-weight: bold; cursor: pointer; border-bottom: 1px solid var(--app-border); }
.topbar { background: #fff; box-shadow: 0 1px 4px rgba(0,0,0,.08); display: flex; align-items: center; justify-content: space-between; padding: 0 20px; }
.collapse-btn { font-size: 20px; cursor: pointer; }
.topbar-right { display: flex; align-items: center; gap: 20px; }
.user-info { display: flex; align-items: center; gap: 4px; cursor: pointer; }
.main-content { padding: 20px; background: var(--app-bg); }
</style>
