<template>
  <el-container style="height:100vh">
    <el-aside :width="collapsed ? '64px' : '220px'" class="sidebar">
      <div class="logo-area" @click="$router.push('/admin/dashboard')">
        <span v-if="!collapsed">管理后台</span>
        <span v-else>管</span>
      </div>
      <el-menu :default-active="route.path" router :collapse="collapsed" background-color="#fff" text-color="#374151" active-text-color="#2563eb">
        <el-menu-item index="/admin/dashboard"><el-icon><DataAnalysis /></el-icon><span>数据看板</span></el-menu-item>
        <el-menu-item index="/admin/meetings"><el-icon><Calendar /></el-icon><span>会议管理</span></el-menu-item>
        <el-menu-item index="/admin/registrations"><el-icon><Tickets /></el-icon><span>报名管理</span></el-menu-item>
        <el-menu-item index="/admin/sign-in"><el-icon><Checked /></el-icon><span>签到管理</span></el-menu-item>
        <el-menu-item index="/admin/votes"><el-icon><Histogram /></el-icon><span>投票管理</span></el-menu-item>
        <el-menu-item index="/admin/finance"><el-icon><Money /></el-icon><span>财务管理</span></el-menu-item>
        <el-menu-item index="/admin/live"><el-icon><VideoCamera /></el-icon><span>直播管理</span></el-menu-item>
        <el-menu-item index="/admin/surveys"><el-icon><Document /></el-icon><span>问卷管理</span></el-menu-item>
        <el-menu-item index="/admin/credits"><el-icon><Star /></el-icon><span>学分管理</span></el-menu-item>
        <el-menu-item index="/admin/archives"><el-icon><FolderOpened /></el-icon><span>档案归档</span></el-menu-item>
        <el-menu-item index="/admin/notifications"><el-icon><Message /></el-icon><span>通知管理</span></el-menu-item>
        <el-menu-item index="/admin/users"><el-icon><UserFilled /></el-icon><span>用户管理</span></el-menu-item>
        <el-menu-item index="/admin/audit"><el-icon><List /></el-icon><span>系统日志</span></el-menu-item>
        <el-menu-item index="/admin/profile"><el-icon><User /></el-icon><span>个人资料</span></el-menu-item>
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
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { useAdminStore } from '@/stores/admin'

const route = useRoute()
const adminStore = useAdminStore()
const collapsed = ref(false)
</script>

<style scoped>
.sidebar { background: #fff; border-right: 1px solid #e5e7eb; overflow: hidden; transition: width .3s; }
.logo-area { height: 60px; display: flex; align-items: center; justify-content: center; color: #2563eb; font-size: 18px; font-weight: bold; cursor: pointer; border-bottom: 1px solid #e5e7eb; }
.topbar { background: #fff; box-shadow: 0 1px 4px rgba(0,0,0,.08); display: flex; align-items: center; justify-content: space-between; padding: 0 20px; }
.collapse-btn { font-size: 20px; cursor: pointer; }
.topbar-right { display: flex; align-items: center; gap: 20px; }
.user-info { display: flex; align-items: center; gap: 4px; cursor: pointer; }
.main-content { padding: 20px; background: #f3f4f6; }
</style>
