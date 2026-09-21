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
        <el-button type="primary" @click="goHome">
          返回首页
        </el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { resolveHomePath } from '@/utils/permission'

const router = useRouter()
const userStore = useUserStore()

// 硬编码 /dashboard 在未授「管理首页」菜单时会再次落到 404
const goHome = () => {
  router.push(resolveHomePath(userStore.menuList))
}
</script>

<style scoped>
.error-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f9ff;
}
</style>