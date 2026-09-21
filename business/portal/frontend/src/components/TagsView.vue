<template>
  <div class="tags-view">
    <div class="tags-list">
      <div
        v-for="tag in visitedViews"
        :key="tag.path"
        class="tags-view-item"
        :class="{ active: isActive(tag) }"
        @click="handleClick(tag)"
      >
        {{ tag.title }}
        <el-icon
          v-if="tag.path !== homePath"
          class="close-icon"
          @click.stop="handleClose(tag)"
        >
          <Close />
        </el-icon>
      </div>
    </div>
    <div class="tags-action">
      <el-tooltip content="关闭所有页面" placement="bottom">
        <el-icon class="close-all-icon" @click="handleCloseAll">
          <CircleClose />
        </el-icon>
      </el-tooltip>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Close, CircleClose } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { resolveHomePath } from '@/utils/permission'

interface TagView {
  path: string
  title: string
}

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 首页标签跟随当前用户已授权菜单：无「管理首页」菜单（如仅授「待审核」的审核角色）时退回首个已授权菜单
const homePath = resolveHomePath(userStore.menuList)

const visitedViews = ref<TagView[]>([
  { path: homePath, title: '首页' }
])

const isActive = (tag: TagView) => tag.path === route.path

const addView = () => {
  const { path, meta } = route
  if (visitedViews.value.some(v => v.path === path)) return
  if (meta?.title) {
    visitedViews.value.push({ path, title: meta.title as string })
  }
}

const handleClick = (tag: TagView) => {
  if (tag.path !== route.path) {
    router.push(tag.path)
  }
}

const handleClose = (tag: TagView) => {
  const index = visitedViews.value.findIndex(v => v.path === tag.path)
  visitedViews.value.splice(index, 1)
  if (isActive(tag) && visitedViews.value.length > 0) {
    const lastTag = visitedViews.value[visitedViews.value.length - 1]
    router.push(lastTag.path)
  }
}

const handleCloseAll = () => {
  visitedViews.value = [{ path: homePath, title: '首页' }]
  if (route.path !== homePath) {
    router.push(homePath)
  }
}

watch(() => route.path, addView, { immediate: true })
</script>

<style scoped lang="scss">
.tags-view {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 20px;
  background: #fff;
  border-bottom: 1px solid var(--app-border);

  .tags-list {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: 1;
    overflow-x: auto;
  }

  .tags-action {
    display: flex;
    align-items: center;
    margin-left: 12px;
    padding-left: 12px;
    border-left: 1px solid #eef1f6;
    color: #909399;

    .close-all-icon {
      font-size: 16px;
      cursor: pointer;
      transition: all 0.2s;

      &:hover {
        color: #f56c6c;
      }
    }
  }

  .tags-view-item {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 12px;
    border-radius: 8px;
    font-size: 13px;
    color: var(--app-text-secondary);
    background: #f5f7fb;
    border: 1px solid #eef1f6;
    cursor: pointer;
    white-space: nowrap;
    transition: all 0.25s;

    &:hover {
      color: var(--el-color-primary);
      border-color: var(--el-color-primary-light-7);
    }

    &.active {
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9, #e6eaf6);
      border-color: var(--el-color-primary-light-7, #b3c1e5);
      font-weight: 600;
      box-shadow: 0 2px 6px rgba(0, 47, 167, 0.15);
    }

    .close-icon {
      font-size: 12px;
      color: #909399;
      width: 14px;
      height: 14px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 50%;
      transition: all 0.2s;

      &:hover {
        color: #f56c6c;
        background: rgba(245, 108, 108, 0.1);
      }
    }
  }
}
</style>
