<template>
  <el-container class="layout-container">
    <el-aside :width="appStore.collapsed ? '64px' : '220px'" class="sidebar">
      <div class="logo">
        <el-icon size="28" color="#fff"><MenuIcon /></el-icon>
        <span v-show="!appStore.collapsed" class="logo-text">Base 平台</span>
      </div>
      <el-scrollbar class="menu-scroll">
        <el-menu
          :default-active="activeMenu"
          :collapse="appStore.collapsed"
          :collapse-transition="false"
          router
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
        >
          <sub-menu v-for="menu in userStore.menus" :key="menu.id" :menu="menu" />
        </el-menu>
      </el-scrollbar>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="appStore.toggleCollapse">
            <Fold v-if="!appStore.collapsed" />
            <Expand v-else />
          </el-icon>
          <breadcrumb />
        </div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              {{ userStore.userInfo?.realName || userStore.userInfo?.username }}
              <el-icon class="el-icon--right"><arrow-down /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item command="password">修改密码</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main">
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
import SubMenu from './components/SubMenu.vue'
import Breadcrumb from '@/components/Breadcrumb.vue'
import { Menu as MenuIcon, Fold, Expand, ArrowDown } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()

const activeMenu = computed(() => route.path)

const handleCommand = (command: string) => {
  if (command === 'logout') {
    userStore.logout()
  } else if (command === 'profile') {
    router.push('/base/profile')
  } else if (command === 'password') {
    router.push('/base/profile')
  }
}
</script>

<style scoped lang="scss">
.layout-container {
  height: 100vh;
}
.sidebar {
  background-color: #304156;
  transition: width 0.3s;
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  border-bottom: 1px solid #1f2d3d;
}
.logo-text {
  margin-left: 12px;
}
.menu-scroll {
  height: calc(100vh - 60px);
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}
.header-left {
  display: flex;
  align-items: center;
}
.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  margin-right: 16px;
}
.user-info {
  cursor: pointer;
  display: flex;
  align-items: center;
}
.main {
  background: #f0f2f5;
  padding: 16px;
  overflow: auto;
}
</style>
