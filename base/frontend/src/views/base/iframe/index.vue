<template>
  <div class="iframe-page">
    <el-alert
      v-if="ticketFailed"
      title="未能签发子应用接入票据，子应用可能无法识别当前登录用户"
      description="请刷新页面重试；若持续失败，请检查登录状态是否已过期。"
      type="warning"
      :closable="false"
      show-icon
      class="ticket-alert"
    />
    <iframe v-if="url" :src="url" frameborder="0" class="iframe-content"></iframe>
    <el-empty v-else description="未配置子应用地址（菜单的「组件路径」需填写 iframe 地址）" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { buildAppEntryUrl } from '@/utils/appEntry'

const route = useRoute()
const url = ref('')
const ticketFailed = ref(false)

// 与 base/DEPLOY.md「五、子应用接入」约定一致：加载子应用时现场签一张一次性票据（base_ticket），
// 不再把 access token 放进 URL；子应用用它换会话。
const loadUrl = async () => {
  const base = (route.meta.url as string) || ''
  ticketFailed.value = false
  if (!base) {
    url.value = ''
    return
  }
  const target = await buildAppEntryUrl(base)
  ticketFailed.value = target.includes('base_ticket_error=1')
  url.value = target
}

watch(() => route.meta.url, loadUrl, { immediate: true })
</script>

<style scoped lang="scss">
.iframe-page {
  height: calc(100vh - 96px);
  background: #fff;
  margin: -16px;
}
.ticket-alert {
  margin: 12px 16px;
}
.iframe-content {
  width: 100%;
  height: 100%;
}
</style>
