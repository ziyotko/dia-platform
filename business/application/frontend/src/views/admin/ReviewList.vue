<template>
  <div class="page-card">
    <!-- Reviewer view: my review tasks -->
    <template v-if="isReviewer">
      <div class="page-toolbar">
        <el-select v-model="status" placeholder="全部状态" clearable style="width:150px" @change="fetch">
          <el-option label="待评审" value="pending" />
          <el-option label="已评分" value="scored" />
        </el-select>
      </div>
      <el-table :data="assignments" v-loading="loading">
        <el-table-column label="项目名称" min-width="220">
          <template #default="{ row }">{{ row.application?.title || '-' }}</template>
        </el-table-column>
        <el-table-column label="所属批次" min-width="160">
          <template #default="{ row }">{{ row.application?.batch?.title || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }"><el-tag :type="row.status === 'scored' ? 'success' : 'warning'" size="small">{{ reviewStatusMap[row.status] }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="score" label="我的评分" width="100" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="$router.push(`/admin/reviews/${row.id}`)">{{ row.status === 'scored' ? '查看' : '评审' }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </template>

    <!-- Manager view: applications in review -->
    <template v-else>
      <div class="page-toolbar">
        <el-select v-model="status" placeholder="全部状态" clearable style="width:160px" @change="fetch">
          <el-option label="待评审" value="under_review" />
          <el-option label="评审完成" value="reviewed" />
        </el-select>
      </div>
      <el-table :data="list" v-loading="loading">
        <el-table-column prop="title" label="项目名称" min-width="220" />
        <el-table-column label="申报人" width="110">
          <template #default="{ row }">{{ row.user?.realName || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }"><el-tag :type="applicationStatusType[row.status]">{{ applicationStatusMap[row.status] }}</el-tag></template>
        </el-table-column>
        <el-table-column label="平均分" width="90">
          <template #default="{ row }">{{ row.avgScore || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="$router.push(`/admin/applications/${row.id}`)">分配/评审</el-button>
          </template>
        </el-table-column>
      </el-table>
    </template>

    <el-pagination
      style="margin-top:16px;justify-content:flex-end"
      layout="total, prev, pager, next"
      :total="total" :page-size="pageSize" :current-page="page"
      @current-change="onPage"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAdminStore } from '@/stores/admin'
import { adminApi } from '@/api/admin'
import { applicationStatusMap, applicationStatusType, reviewStatusMap } from '@/utils/constants'

const adminStore = useAdminStore()
const isReviewer = computed(() => adminStore.roleCode === 'reviewer')

const assignments = ref<any[]>([])
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const status = ref('')
const loading = ref(false)

async function fetch() {
  loading.value = true
  try {
    if (isReviewer.value) {
      const res = await adminApi.getMyReviews({ page: page.value, pageSize, status: status.value })
      assignments.value = res.data.list
      total.value = res.data.total
    } else {
      const res = await adminApi.getApplications({
        page: page.value, pageSize,
        status: status.value || undefined,
      })
      // only show review-stage applications
      const filtered = status.value ? res.data.list : res.data.list.filter((a: any) => ['under_review', 'reviewed'].includes(a.status))
      list.value = filtered
      total.value = status.value ? res.data.total : filtered.length
    }
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }

onMounted(fetch)
</script>
