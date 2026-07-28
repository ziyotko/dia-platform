<template>
  <div class="not-found">
    <h1>404</h1>
    <p>页面未找到</p>
    <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
  </div>
</template>

<style scoped lang="scss">
.not-found {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  height: 100vh; gap: 16px;
  h1 { font-size: 72px; color: #1a6fb5; }
  p { font-size: 18px; color: #6b7280; }
}
</style>
