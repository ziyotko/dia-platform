<template>
  <div class="admin-profile-changes" v-loading="loading">
    <div class="page-header">
      <h3>资料变更记录</h3>
      <div class="filters">
        <el-input v-model="keyword" placeholder="搜索修改人/原始名称/变更名称" clearable style="width:260px" @clear="search" @keyup.enter="search" />
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </div>

    <el-card>
      <el-table :data="list" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" min-width="70" />
        <el-table-column prop="operator" label="修改人" min-width="120">
          <template #default="{ row }">{{ row.operator || '-' }}</template>
        </el-table-column>
        <el-table-column label="原始name" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.old_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="变更name" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.new_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="变更时间" min-width="160">
          <template #default="{ row }">{{ fmt(row.changed_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" min-width="100" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" size="small" @click="openDetail(row)">查看详情</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length === 0" description="暂无资料变更记录" />
      <div class="pagination">
        <el-pagination background layout="prev, pager, next, total" :total="total" :page-size="size" v-model:current-page="page" @change="fetchData" />
      </div>
    </el-card>

    <el-dialog v-model="detailVisible" title="资料变更详情" width="720px" append-to-body>
      <div v-if="currentRow" class="detail-head">
        <div><span class="label">修改人：</span>{{ currentRow.operator || '-' }}</div>
        <div><span class="label">变更时间：</span>{{ fmt(currentRow.changed_at) }}</div>
        <div><span class="label">原始name：</span>{{ currentRow.old_name || '-' }}</div>
        <div><span class="label">变更name：</span>{{ currentRow.new_name || '-' }}</div>
      </div>
      <el-table :data="diff" stripe size="small" style="width: 100%">
        <el-table-column prop="field" label="字段" min-width="120" />
        <el-table-column label="原始内容" min-width="200">
          <template #default="{ row }">
            <span :class="{ 'dim': isEmpty(row.oldValue) }">{{ displayValue(row.oldValue) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="变更后内容" min-width="200">
          <template #default="{ row }">
            <span :class="{ 'changed': !isEmpty(row.newValue) }">{{ displayValue(row.newValue) }}</span>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="diff.length === 0" description="无字段变更" :image-size="80" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([])
const loading = ref(true)
const page = ref(1); const size = ref(10); const total = ref(0)
const keyword = ref('')

const detailVisible = ref(false)
const currentRow = ref<any>(null)
const diff = ref<{ field: string; oldValue: any; newValue: any }[]>([])

function fmt(d: string) { return d ? d.replace('T', ' ').slice(0, 19) : '-' }
function parseContent(json: string): Record<string, any> {
  if (!json) return {}
  try { return JSON.parse(json) } catch { return {} }
}
function isEmpty(v: any) { return v === null || v === undefined || v === '' }
function displayValue(v: any) { return isEmpty(v) ? '-' : String(v) }

function openDetail(row: any) {
  currentRow.value = row
  const oldObj = parseContent(row.old_content)
  const newObj = parseContent(row.new_content)
  const keys = new Set([...Object.keys(oldObj), ...Object.keys(newObj)])
  const items: { field: string; oldValue: any; newValue: any }[] = []
  keys.forEach((k) => {
    const o = oldObj[k]
    const n = newObj[k]
    if (String(o ?? '') !== String(n ?? '')) {
      items.push({ field: k, oldValue: o, newValue: n })
    }
  })
  diff.value = items
  detailVisible.value = true
}

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

.detail-head {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 24px;
  margin-bottom: 16px;
  padding: 12px;
  background: #f5f7fb;
  border-radius: 8px;
  color: #344054;
  .label { color: #667085; }
}
.dim { color: #98a2b3; }
.changed { color: #002fa7; font-weight: 600; }
</style>
