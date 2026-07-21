<template>
  <div class="iframe-page">
    <iframe :src="url" frameborder="0" class="iframe-content"></iframe>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const url = computed(() => route.meta.url as string)
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
