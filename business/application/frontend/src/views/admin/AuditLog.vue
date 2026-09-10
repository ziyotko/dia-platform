<template>
  <div class="page-card">
    <div class="page-toolbar">
      <el-input v-model="keyword" placeholder="搜索操作/管理员" clearable style="width:240px" @keyup.enter="fetch" @clear="fetch" />
      <el-select v-model="module" placeholder="全部模块" clearable style="width:160px" @change="fetch">
        <el-option label="批次管理" value="批次管理" />
        <el-option label="类别管理" value="类别管理" />
        <el-option label="初审管理" value="初审管理" />
        <el-option label="评审管理" value="评审管理" />
        <el-option label="结果公示" value="结果公示" />
        <el-option label="证书管理" value="证书管理" />
        <el-option label="用户管理" value="用户管理" />
        <el-option label="账号管理" value="账号管理" />
        <el-option label="通知管理" value="通知管理" />
      </el-select>
      <el-button type="primary" @click="fetch">查询</el-button>
    </div>

    <el-table :data="list" v-loading="loading">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="admin" label="操作人" width="120" />
      <el-table-column prop="module" label="模块" width="120" />
      <el-table-column prop="action" label="操作" min-width="160" />
      <el-table-column prop="ip" label="IP" width="130" />
      <el-table-column label="时间" width="170">
        <template #default="{ row }">{{ fmt(row.createdAt) }}</template>
      </el-table-column>
    </el-table>

    <el-pagination
      style="margin-top:16px;justify-content:flex-end"
      layout="total, prev, pager, next"
      :total="total" :page-size="pageSize" :current-page="page"
      @current-change="onPage"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { fmt } from '@/utils/constants'

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const keyword = ref('')
const module = ref('')
const loading = ref(false)

async function fetch() {
  loading.value = true
  try {
    const res = await adminApi.getAuditLogs({ page: page.value, pageSize, keyword: keyword.value, module: module.value })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }

onMounted(fetch)
</script>
