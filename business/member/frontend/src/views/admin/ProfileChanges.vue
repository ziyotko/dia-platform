<template>
  <div class="admin-profile-changes" v-loading="loading">
    <div class="page-header">
      <h3>资料变更记录</h3>
      <div class="filters">
        <el-input v-model="keyword" placeholder="搜索修改人/字段名称" clearable style="width:260px" @clear="search" @keyup.enter="search" />
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </div>

    <el-card>
      <el-table :data="list" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" min-width="70" />
        <el-table-column prop="username" label="用户名" min-width="120">
          <template #default="{ row }">{{ row.username || '-' }}</template>
        </el-table-column>
        <el-table-column prop="operator" label="修改人" min-width="120">
          <template #default="{ row }">{{ row.operator || '-' }}</template>
        </el-table-column>
        <el-table-column label="字段" min-width="120">
          <template #default="{ row }">{{ row.name || '-' }}</template>
        </el-table-column>
        <el-table-column label="原始内容" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            <span :class="{ 'dim': !row.old_content }">{{ row.old_content || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="变更后内容" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            <span :class="{ 'changed': row.new_content }">{{ row.new_content || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="变更时间" min-width="160">
          <template #default="{ row }">{{ fmt(row.created_at) }}</template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length === 0" description="暂无资料变更记录" />
      <div class="pagination">
        <el-pagination background layout="prev, pager, next, total" :total="total" :page-size="size" v-model:current-page="page" @change="fetchData" />
      </div>
    </el-card>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([])
const loading = ref(true)
const page = ref(1); const size = ref(10); const total = ref(0)
const keyword = ref('')

function fmt(d: string) { return d ? d.replace('T', ' ').slice(0, 19) : '-' }

onMounted(() => fetchData())

async function fetchData() {
  loading.value = true
  try {
    const res = await adminApi.getProfileChanges({
      page: page.value,
      size: size.value,
      keyword: keyword.value
    })
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch {} finally { loading.value = false }
}

function search() { page.value = 1; fetchData() }
</script>

<style scoped lang="scss">
.admin-profile-changes {
  width: 100%;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 12px;
  h3 {
    font-size: 22px;
    font-weight: 600;
    color: #1d2739;
    margin: 0;
  }
}
.filters { display: flex; gap: 12px; }
.el-card { border-radius: 10px; }
.pagination { display: flex; justify-content: center; padding: 20px 0; }
.dim { color: #98a2b3; }
.changed { color: #002fa7; font-weight: 600; }
</style>
