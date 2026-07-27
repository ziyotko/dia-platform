<template>
  <template v-if="!menu.hidden">
    <el-sub-menu v-if="menu.type === 'directory' && menu.children?.length" :index="menu.path" popper-class="dark-popper">
      <template #title>
        <el-icon v-if="menu.icon" class="menu-icon"><component :is="menu.icon" /></el-icon>
        <span class="menu-title">{{ menu.name }}</span>
      </template>
      <sub-menu v-for="child in menu.children" :key="child.id" :menu="child" />
    </el-sub-menu>

    <el-menu-item v-else :index="menu.path">
      <el-icon v-if="menu.icon" class="menu-icon"><component :is="menu.icon" /></el-icon>
      <template #title>
        <span class="menu-title">{{ menu.name }}</span>
      </template>
    </el-menu-item>
  </template>
</template>

<script setup lang="ts">
import type { Menu } from '@/api/menu'

defineProps<{
  menu: Menu
}>()
</script>

<style scoped lang="scss">
.menu-icon {
  font-size: 18px;
  margin-right: 4px;
}

.menu-title {
  font-size: 14px;
}

:deep(.el-menu-item),
:deep(.el-sub-menu__title) {
  height: 46px;
  line-height: 46px;
  margin-bottom: 4px;
  border-radius: 8px;
  transition: all 0.25s ease;

  &:hover {
    background: rgba(59, 130, 246, 0.06) !important;
    color: #2563eb !important;
  }
}

:deep(.el-menu-item.is-active) {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%) !important;
  color: #fff !important;
  box-shadow: 0 6px 16px rgba(37, 99, 235, 0.35);
}

:deep(.el-sub-menu.is-active > .el-sub-menu__title) {
  color: #2563eb !important;
}
</style>
