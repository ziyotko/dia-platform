<template>
  <el-breadcrumb separator="/">
    <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
    <template v-if="route.matched.length > 1">
      <el-breadcrumb-item
        v-for="item in route.matched.slice(0, -1)"
        :key="item.path"
        :to="{ path: item.path }"
      >
        {{ item.meta.title }}
      </el-breadcrumb-item>
    </template>
    <el-breadcrumb-item v-if="route.meta.title && route.path !== '/dashboard'">
      {{ route.meta.title }}
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>