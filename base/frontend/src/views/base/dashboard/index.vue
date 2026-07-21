<template>
  <div class="dashboard-page">
    <el-row :gutter="16">
      <el-col :span="6">
        <el-card>
          <stat-card title="租户数" :value="stats.tenantCount" icon="OfficeBuilding" color="#409EFF" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <stat-card title="应用数" :value="stats.appCount" icon="Grid" color="#67C23A" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <stat-card title="用户数" :value="stats.userCount" icon="User" color="#E6A23C" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <stat-card title="角色数" :value="stats.roleCount" icon="UserFilled" color="#F56C6C" />
        </el-card>
      </el-col>
    </el-row>

    <el-card class="mt-16">
      <template #header>
        <span>欢迎使用 Base 底座平台</span>
      </template>
      <p>Base 底座平台提供统一的登录、租户、应用、用户、角色、菜单、权限、审计等基础能力。</p>
      <p>已有系统可通过 IFrame、API 代理或微应用方式接入底座，实现一站式管理。</p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import StatCard from './components/StatCard.vue'

const stats = reactive({
  tenantCount: 12,
  appCount: 5,
  userCount: 128,
  roleCount: 18
})
</script>

<style scoped lang="scss">
.dashboard-page {
  .mt-16 {
    margin-top: 16px;
  }
  p {
    line-height: 2;
    color: #606266;
  }
}
</style>
