<template>
  <el-breadcrumb separator="/" class="app-breadcrumb">
    <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path" :to="item.path">
      <el-icon v-if="item.path === '/'" class="breadcrumb-home"><House /></el-icon>
      <span>{{ item.title }}</span>
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { House } from '@element-plus/icons-vue'

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

<style scoped lang="scss">
.app-breadcrumb {
  :deep(.el-breadcrumb__item) {
    .el-breadcrumb__inner {
      color: #64748b;
      font-size: 14px;
      font-weight: 500;
      display: inline-flex;
      align-items: center;
      gap: 4px;
      transition: color 0.25s ease;

      &:hover {
        color: #2563eb;
      }
    }

    &:last-child .el-breadcrumb__inner {
      color: #1e293b;
      font-weight: 600;
    }
  }

  .breadcrumb-home {
    font-size: 14px;
  }
}
</style>
