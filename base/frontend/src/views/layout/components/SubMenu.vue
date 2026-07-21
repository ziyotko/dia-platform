<template>
  <template v-if="!menu.hidden">
    <el-sub-menu v-if="menu.type === 'directory' && menu.children?.length" :index="menu.path">
      <template #title>
        <el-icon v-if="menu.icon"><component :is="menu.icon" /></el-icon>
        <span>{{ menu.name }}</span>
      </template>
      <sub-menu v-for="child in menu.children" :key="child.id" :menu="child" />
    </el-sub-menu>

    <el-menu-item v-else :index="menu.path">
      <el-icon v-if="menu.icon"><component :is="menu.icon" /></el-icon>
      <template #title>{{ menu.name }}</template>
    </el-menu-item>
  </template>
</template>

<script setup lang="ts">
import type { Menu } from '@/api/menu'

defineProps<{
  menu: Menu
}>()
</script>
