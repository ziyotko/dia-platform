<template>
  <router-view />
</template>

<script setup lang="ts">
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

/* 字体统一由 styles/global.scss 的 `--el-font-family` 提供（Element Plus 组件与正文同一字体栈）。
   此处不得再声明 body 字体：本样式非 scoped 且后加载，同优先级下会覆盖字体令牌，
   曾导致组件（读 --el-font-family）与正文（读 body）出现两套字体栈。 */
</style>