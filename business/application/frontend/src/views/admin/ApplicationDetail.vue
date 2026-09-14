<template>
  <div v-loading="loading">
    <div class="page-card" style="max-width:1000px">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
        <h3>{{ app.title || '申报详情' }}</h3>
        <el-tag :type="applicationStatusType[app.status]" size="large">{{ applicationStatusMap[app.status] || app.status }}</el-tag>
      </div>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="申报人">{{ app.user?.realName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="所在单位">{{ app.user?.organization || '-' }}</el-descriptions-item>
        <el-descriptions-item label="申报批次">{{ app.batch?.title || '-' }}</el-descriptions-item>
        <el-descriptions-item label="项目类别">{{ app.category?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="项目名称" :span="2">{{ app.title }}</el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ fmt(app.submittedAt) }}</el-descriptions-item>
        <el-descriptions-item label="公示时间">{{ fmt(app.publishedAt) }}</el-descriptions-item>
        <el-descriptions-item label="项目简介" :span="2">{{ app.projectBrief || '-' }}</el-descriptions-item>
        <el-descriptions-item label="申报内容" :span="2">{{ app.content || '-' }}</el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">申报材料</el-divider>
      <el-table :data="app.materials || []" size="small">
        <el-table-column prop="name" label="材料名称" min-width="220" />
        <el-table-column prop="fileType" label="类型" width="90" />
        <el-table-column label="大小" width="110">
          <template #default="{ row }">{{ (row.fileSize / 1024).toFixed(1) }} KB</template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }"><el-button size="small" type="primary" link @click="download(row)">下载</el-button></template>
        </el-table-column>
      </el-table>

      <template v-if="app.preliminaryOpinion">
        <el-divider content-position="left">初审意见</el-divider>
        <el-alert :closable="false" :title="app.preliminaryOpinion" :type="app.status === 'preliminary_rejected' ? 'error' : 'info'" />
      </template>

      <template v-if="app.reviews && app.reviews.length">
        <el-divider content-position="left">专家评审</el-divider>
        <el-table :data="app.reviews" size="small">
          <el-table-column label="评审人" width="140">
            <template #default="{ row }">{{ row.reviewer?.realName || '-' }}</template>
          </el-table-column>
          <el-table-column prop="score" label="评分" width="100" />
          <el-table-column prop="comment" label="评审意见" min-width="220" />
        </el-table>
        <div style="margin-top:8px;color:#667085">综合平均分：<b style="color:#002fa7">{{ app.avgScore }}</b></div>
      </template>

      <!-- Final opinion -->
      <template v-if="app.finalOpinion">
        <el-divider content-position="left">评审结果</el-divider>
        <el-alert :closable="false" :title="app.finalOpinion" :type="app.status === 'rejected' ? 'error' : 'success'" />
      </template>

      <!-- Actions -->
      <template v-if="actionsVisible">
        <el-divider content-position="left">操作</el-divider>
        <div style="display:flex;gap:12px;flex-wrap:wrap">
          <template v-if="app.status === 'submitted'">
            <el-button type="success" @click="preliminary(true)">初审通过</el-button>
            <el-button type="danger" @click="preliminary(false)">初审驳回</el-button>
          </template>
          <template v-if="app.status === 'under_review' || app.status === 'reviewed'">
            <el-button type="primary" @click="openAssign">分配评审</el-button>
          </template>
          <template v-if="app.status === 'reviewed'">
            <el-button type="success" @click="finalize(true)">确定通过</el-button>
            <el-button type="danger" @click="finalize(false)">确定不通过</el-button>
          </template>
          <template v-if="app.status === 'passed' || app.status === 'rejected'">
            <el-button v-if="!app.publishedAt" type="primary" @click="publishResult">公示结果</el-button>
            <el-tag v-else type="success">结果已公示</el-tag>
          </template>
        </div>
      </template>
    </div>

    <el-dialog v-model="opinionDialog" title="填写意见" width="520px">
      <el-input v-model="opinion" type="textarea" :rows="3" placeholder="请输入评审/初审意见" />
      <template #footer>
        <el-button @click="opinionDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmOpinion">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="assignDialog" title="分配评审专家" width="520px">
      <el-select v-model="reviewerIds" multiple placeholder="请选择评审人" style="width:100%">
        <el-option v-for="r in reviewers" :key="r.id" :label="r.realName" :value="r.id" />
      </el-select>
      <template #footer>
        <el-button @click="assignDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmAssign">确定分配</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'
import { applicationStatusMap, applicationStatusType, fileUrl, fmt } from '@/utils/constants'

const route = useRoute()
const id = Number(route.params.id)
const loading = ref(false)
const app = ref<any>({})

const opinionDialog = ref(false)
const opinion = ref('')
const pendingAction = ref('')

const assignDialog = ref(false)
const reviewers = ref<any[]>([])
const reviewerIds = ref<number[]>([])

const actionsVisible = computed(() => {
  return ['submitted', 'under_review', 'reviewed', 'passed', 'rejected'].includes(app.value.status)
})

async function fetch() {
  loading.value = true
  try {
    const res = await adminApi.getApplication(id)
    app.value = res.data
  } finally {
    loading.value = false
  }
}

function preliminary(pass: boolean) {
  pendingAction.value = pass ? 'preliminary_pass' : 'preliminary_reject'
  opinion.value = ''
  opinionDialog.value = true
}

function finalize(pass: boolean) {
  pendingAction.value = pass ? 'finalize_pass' : 'finalize_reject'
  opinion.value = ''
  opinionDialog.value = true
}

async function confirmOpinion() {
  if (pendingAction.value === 'preliminary_pass' || pendingAction.value === 'preliminary_reject') {
    await adminApi.preliminaryReview(id, { pass: pendingAction.value === 'preliminary_pass', opinion: opinion.value })
    ElMessage.success('初审完成')
  } else {
    await adminApi.finalize(id, { pass: pendingAction.value === 'finalize_pass', opinion: opinion.value })
    ElMessage.success('评审结果已确定')
  }
  opinionDialog.value = false
  fetch()
}

async function openAssign() {
  const res = await adminApi.getReviewers()
  reviewers.value = res.data
  reviewerIds.value = (app.value.reviews || []).map((r: any) => r.reviewerId)
  assignDialog.value = true
}

async function confirmAssign() {
  await adminApi.assignReviewers(id, { reviewerIds: reviewerIds.value })
  ElMessage.success('分配成功')
  assignDialog.value = false
  fetch()
}

async function publishResult() {
  await adminApi.publishResult(id)
  ElMessage.success('结果已公示')
  fetch()
}

function download(row: any) {
  window.open(fileUrl(row.fileUrl), '_blank')
}

onMounted(fetch)
</script>
