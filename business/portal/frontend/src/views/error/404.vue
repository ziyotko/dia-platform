<template>
  <div class="error-page">
    <el-result
      icon="error"
      title="404"
      sub-title="抱歉，您访问的页面不存在"
    >
      <template #icon>
        <el-image
          src="https://gw.alipayobjects.com/zos/antfincdn/zNkK7KS38h/404.svg"
          style="width: 300px;"
        />
      </template>
      <template #extra>
        <el-button type="primary" @click="$router.push('/dashboard')">
          返回首页
        </el-button>
      </template>
    </el-result>
  </div>
</template>

<style scoped>
.error-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f9ff;
}
</style>