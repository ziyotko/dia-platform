<template>
  <div class="page-card">
    <div class="page-toolbar">
      <div style="display:flex;gap:12px;flex-wrap:wrap">
        <el-input v-model="keyword" placeholder="搜索项目名称" clearable style="width:200px" @keyup.enter="fetch" @clear="fetch" />
        <el-select v-model="batchId" placeholder="全部批次" clearable filterable style="width:220px" @change="fetch">
          <el-option v-for="b in batches" :key="b.id" :label="b.title" :value="b.id" />
        </el-select>
        <el-select v-model="status" placeholder="全部状态" clearable style="width:150px" @change="fetch">
          <el-option v-for="(label, key) in applicationStatusMap" :key="key" :label="label" :value="key" />
        </el-select>
        <el-button type="primary" @click="fetch">查询</el-button>
      </div>
    </div>

    <el-table :data="list" v-loading="loading">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="title" label="项目名称" min-width="200" />
      <el-table-column label="申报人" width="110">
        <template #default="{ row }">{{ row.user?.realName || '-' }}</template>
      </el-table-column>
      <el-table-column label="批次" min-width="150">
        <template #default="{ row }">{{ row.batch?.title || '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="110">
        <template #default="{ row }"><el-tag :type="applicationStatusType[row.status]">{{ applicationStatusMap[row.status] || row.status }}</el-tag></template>
      </el-table-column>
      <el-table-column label="评审平均分" width="110">
        <template #default="{ row }">{{ avgScoreText(row) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="$router.push(`/admin/applications/${row.id}`)">详情</el-button>
          <el-button v-if="row.status === 'submitted'" size="small" type="warning" @click="quickPreliminary(row, true)">初审通过</el-button>
          <el-button v-if="row.status === 'submitted'" size="small" type="danger" @click="quickPreliminary(row, false)">初审驳回</el-button>
        </template>
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
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'
import { applicationStatusMap, applicationStatusType, avgScoreText } from '@/utils/constants'

const list = ref<any[]>([])
const batches = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const keyword = ref('')
const batchId = ref<any>('')
const status = ref('')
const loading = ref(false)

async function fetch() {
  loading.value = true
  try {
    const res = await adminApi.getApplications({
      page: page.value, pageSize, keyword: keyword.value, status: status.value,
      batchId: batchId.value || undefined,
    })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }

async function quickPreliminary(row: any, pass: boolean) {
  await adminApi.preliminaryReview(row.id, { pass, opinion: pass ? '初审通过' : '初审不通过' })
  ElMessage.success('初审完成')
  fetch()
}

onMounted(async () => {
  fetch()
  const res = await adminApi.getBatches({ page: 1, pageSize: 100 })
  batches.value = res.data.list
})
</script>
