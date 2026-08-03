<template>
  <div class="public-layout">
    <header class="public-header">
      <div class="header-inner">
        <h1 class="logo" @click="$router.push('/')">会议管理系统</h1>
        <nav class="nav-links">
          <router-link to="/login">会员登录</router-link>
          <router-link to="/admin/login">管理后台</router-link>
        </nav>
      </div>
    </header>
    <main class="public-main">
      <router-view />
    </main>
    <footer class="public-footer">
      <p>© 2024 会议管理系统. All rights reserved.</p>
    </footer>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.public-layout { display: flex; flex-direction: column; min-height: 100vh; }
.public-header { background: #fff; box-shadow: 0 1px 4px rgba(0,0,0,.08); padding: 0 24px; height: 60px; display: flex; align-items: center; }
.header-inner { width: 100%; max-width: 1200px; margin: 0 auto; display: flex; justify-content: space-between; align-items: center; }
.logo { font-size: 20px; color: #2563eb; cursor: pointer; }
.nav-links a { margin-left: 24px; color: #666; text-decoration: none; }
.nav-links a:hover { color: #2563eb; }
.public-main { flex: 1; width: 100%; max-width: 1200px; margin: 0 auto; padding: 24px; }
.public-footer { background: #f5f5f5; padding: 16px; text-align: center; color: #999; font-size: 13px; }
</style>
