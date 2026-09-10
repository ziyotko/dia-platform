<template>
  <div class="admin-level-changes" v-loading="loading">
    <div class="page-header">
      <h3>会籍变更记录</h3>
      <div class="filters">
        <el-input v-model="keyword" placeholder="搜索用户名/公司名/等级" clearable style="width:220px" @clear="search" @keyup.enter="search" />
        <el-select v-model="filterType" placeholder="会员类型" clearable style="width:140px" @change="search">
          <el-option label="单位会员" value="unit" />
          <el-option label="个人会员" value="personal" />
        </el-select>
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </div>

    <el-card>
      <el-table :data="list" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" min-width="70" />
        <el-table-column prop="username" label="变更用户名" min-width="120" />
        <el-table-column prop="member_name" label="公司名称/姓名" min-width="180" />
        <el-table-column label="类型" min-width="90">
          <template #default="{ row }">{{ typeLabel(row.member_type) }}</template>
        </el-table-column>
        <el-table-column label="变更时间" min-width="150">
          <template #default="{ row }">{{ fmt(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="change_year" label="变更年份" min-width="90" />
        <el-table-column prop="org_name" label="入会机构" min-width="150">
          <template #default="{ row }">{{ row.org_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="原始会籍" min-width="150">
          <template #default="{ row }">
            <span v-if="row.old_level_id">{{ row.old_level_name }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="新的会籍" min-width="150">
          <template #default="{ row }">
            <el-tag type="warning">{{ row.new_level_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="变更原因" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.reason || '-' }}</template>
        </el-table-column>
        <el-table-column prop="operator" label="变更人" min-width="110">
          <template #default="{ row }">{{ row.operator || '-' }}</template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length === 0" description="暂无会籍变更记录" />
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
const keyword = ref(''); const filterType = ref('')

function typeLabel(t: string) {
  return t === 'unit' ? '单位' : t === 'personal' ? '个人' : (t || '-')
}
function fmt(d: string) { return d ? d.replace('T', ' ').slice(0, 19) : '-' }

onMounted(() => fetchData())

async function fetchData() {
  loading.value = true
  try {
    const res = await adminApi.getMemberLevelChanges({
      page: page.value,
      size: size.value,
      keyword: keyword.value,
      member_type: filterType.value
    })
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch {} finally { loading.value = false }
}

function search() { page.value = 1; fetchData() }
</script>

<style scoped lang="scss">
.admin-level-changes {
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
</style>
