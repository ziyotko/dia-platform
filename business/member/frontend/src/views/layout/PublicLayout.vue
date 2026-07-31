<template>
  <div class="public-layout">
    <header class="public-header">
      <div class="header-inner">
        <div class="logo" @click="$router.push('/')">
          <el-icon :size="28"><OfficeBuilding /></el-icon>
          <span class="logo-text">{{ siteStore.site_name }}</span>
        </div>
        <nav class="nav-links">
          <router-link to="/announcements">公告动态</router-link>
          <router-link to="/login">会员登录</router-link>
          <router-link to="/register">
            <el-button type="primary" size="small">注册会员</el-button>
          </router-link>
        </nav>
      </div>
    </header>
    <main class="public-main">
      <router-view />
    </main>
    <footer class="public-footer">
      <p>© 2024 xxxxxx 版权所有</p>
      <p>地址：北京市 电话：</p>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { useSiteStore } from '@/stores/site'

const siteStore = useSiteStore()
</script>

<style scoped lang="scss">
.public-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.public-header {
  background: #fff;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
  position: sticky;
  top: 0;
  z-index: 100;
  .header-inner {
    max-width: 1200px;
    margin: 0 auto;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 24px;
    height: 64px;
  }
  .logo {
    display: flex;
    align-items: center;
    gap: 10px;
    cursor: pointer;
    color: #1a6fb5;
    .logo-text {
      font-size: 18px;
      font-weight: 700;
    }
  }
  .nav-links {
    display: flex;
    align-items: center;
    gap: 24px;
    a {
      color: #4b5563;
      text-decoration: none;
      font-size: 15px;
      &:hover { color: #1a6fb5; }
    }
  }
}
.public-main {
  flex: 1;
}
.public-footer {
  text-align: center;
  padding: 32px 24px;
  background: #1f2937;
  color: #9ca3af;
  font-size: 13px;
  line-height: 1.8;
}
</style>
