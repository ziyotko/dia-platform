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

const route = useRoute()
const userStore = useUserStore()

// 与 integration/README.md 约定一致：加载子应用时在 URL 追加底座会话参数，
// 子应用可用 base_token 调 /base/api/v1/auth/info 校验用户身份。
const url = computed(() => {
  const base = (route.meta.url as string) || ''
  if (!base) return ''

  const params = new URLSearchParams({
    base_token: userStore.token || '',
    user_id: String(userStore.userInfo?.id ?? ''),
    username: userStore.userInfo?.username || '',
    tenant_id: String(userStore.userInfo?.tenantId ?? 0)
  })
  return base + (base.includes('?') ? '&' : '?') + params.toString()
})
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
