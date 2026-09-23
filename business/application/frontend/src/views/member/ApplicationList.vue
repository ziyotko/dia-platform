<template>
  <div class="page-card">
    <div class="page-toolbar">
      <div style="display:flex;gap:12px">
        <el-input v-model="keyword" placeholder="搜索项目名称" clearable style="width:220px" @keyup.enter="fetch" @clear="fetch" />
        <el-select v-model="status" placeholder="全部状态" clearable style="width:160px" @change="onFilterChange">
          <el-option v-for="(label, key) in applicationStatusMap" :key="key" :label="label" :value="key" />
        </el-select>
        <el-button type="primary" @click="fetch">查询</el-button>
      </div>
      <el-button type="primary" @click="$router.push('/member/applications/create')">新建申报</el-button>
    </div>

    <el-table :data="list" v-loading="loading">
      <el-table-column prop="title" label="项目名称" min-width="220" />
      <el-table-column label="申报批次" min-width="160">
        <template #default="{ row }">{{ row.batch?.title || '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="applicationStatusType[row.status]">{{ applicationStatusMap[row.status] || row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="评审平均分" width="110">
        <template #default="{ row }">{{ avgScoreText(row) }}</template>
      </el-table-column>
      <el-table-column label="提交时间" width="160">
        <template #default="{ row }">{{ fmt(row.submittedAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="$router.push(`/member/applications/${row.id}`)">查看</el-button>
          <el-button v-if="canEdit(row)" size="small" type="primary" @click="$router.push(`/member/applications/${row.id}`)">修改</el-button>
          <el-button v-if="row.status === 'submitted'" size="small" type="warning" @click="withdraw(row)">撤回</el-button>
          <el-button v-if="canEdit(row)" size="small" type="danger" @click="remove(row)">删除</el-button>
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { memberApi } from '@/api/member'
import { applicationStatusMap, applicationStatusType, avgScoreText, fmt } from '@/utils/constants'

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const keyword = ref('')
const status = ref('')
const loading = ref(false)

async function fetch() {
  loading.value = true
  try {
    const res = await memberApi.getMyApplications({ page: page.value, pageSize, keyword: keyword.value, status: status.value })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }
// 切换筛选条件回到第 1 页，否则在第 2 页切筛选会看到空列表
function onFilterChange() { page.value = 1; fetch() }

// Drafts and preliminary-rejected applications can still be modified.
function canEdit(row: any) {
  return row.status === 'draft' || row.status === 'preliminary_rejected'
}

async function remove(row: any) {
  await ElMessageBox.confirm('确认删除该申报？', '提示', { type: 'warning' })
  await memberApi.deleteApplication(row.id)
  ElMessage.success('删除成功')
  fetch()
}

// Withdrawing returns the submission to draft so it can be edited again.
async function withdraw(row: any) {
  await ElMessageBox.confirm('撤回后申报将回到草稿状态，确认撤回？', '提示', { type: 'warning' })
  await memberApi.withdrawApplication(row.id)
  ElMessage.success('已撤回，可继续修改')
  fetch()
}

onMounted(fetch)
</script>
