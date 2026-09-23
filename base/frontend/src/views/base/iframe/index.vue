<template>
  <div class="iframe-page">
    <iframe v-if="url" :src="url" frameborder="0" class="iframe-content"></iframe>
    <el-empty v-else description="未配置子应用地址（菜单的「组件路径」需填写 iframe 地址）" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { buildAppEntryUrl } from '@/utils/appEntry'

const route = useRoute()
const userStore = useUserStore()

// 与 base/DEPLOY.md「五、子应用接入」约定一致：加载子应用时追加底座会话参数
const url = computed(() =>
  buildAppEntryUrl((route.meta.url as string) || '', {
    token: userStore.token,
    userId: userStore.userInfo?.id,
    username: userStore.userInfo?.username,
    tenantId: userStore.userInfo?.tenantId
  })
)
</script>

<style scoped lang="scss">
.iframe-page {
  height: calc(100vh - 96px);
  background: #fff;
  margin: -16px;
}
.iframe-content {
  width: 100%;
  height: 100%;
}
</style>
