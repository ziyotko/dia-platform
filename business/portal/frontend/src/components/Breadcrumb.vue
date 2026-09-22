<template>
  <el-breadcrumb separator="/">
    <el-breadcrumb-item :to="{ path: homePath }">首页</el-breadcrumb-item>
    <template v-if="route.matched.length > 1">
      <template v-for="item in route.matched.slice(0, -1)" :key="item.path">
        <el-breadcrumb-item
          v-if="item.meta?.title"
          :to="item.meta.type === 'menu' ? { path: item.path } : undefined"
        >
          {{ item.meta.title }}
        </el-breadcrumb-item>
      </template>
    </template>
    <el-breadcrumb-item v-if="route.meta.title && route.path !== '/dashboard'">
      {{ route.meta.title }}
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { resolveHomePath } from '@/utils/permission'

const route = useRoute()
const userStore = useUserStore()

// 硬编码 /dashboard：未授「管理首页」菜单的用户点面包屑「首页」会落到 404（与 TagsView/404 页统一口径）
const homePath = computed(() => resolveHomePath(userStore.menuList))
</script>