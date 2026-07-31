<template>
  <router-view />
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useRoute } from 'vue-router'
import { useSiteStore } from '@/stores/site'

const route = useRoute()
const siteStore = useSiteStore()

// Ensure site config is loaded
siteStore.load()

// Update browser tab title from site_name + current page title
watch(
  () => [route.fullPath, siteStore.site_name],
  () => {
    const name = siteStore.site_name || '会员系统'
    document.title = route.meta.title ? `${name} - ${route.meta.title}` : name
  },
  { immediate: true }
)
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}
html, body, #app {
  height: 100%;
  font-family: 'Helvetica Neue', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>
