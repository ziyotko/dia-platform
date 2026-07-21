<template>
  <el-breadcrumb separator="/">
    <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path" :to="item.path">
      {{ item.title }}
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const breadcrumbs = computed(() => {
  const list: { title: string; path: string }[] = []
  const matched = route.matched.filter((r) => r.meta?.title)
  for (const r of matched) {
    list.push({ title: r.meta.title as string, path: r.path })
  }
  if (list.length === 0) {
    list.push({ title: '首页', path: '/' })
  }
  return list
})
</script>
