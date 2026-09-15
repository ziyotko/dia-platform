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
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="row.status === 'scored' ? 'success' : 'info'" size="small">{{ reviewStatusMap[row.status] || row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="评分" width="90">
            <template #default="{ row }">{{ row.status === 'scored' ? row.score : '-' }}</template>
          </el-table-column>
          <el-table-column prop="comment" label="评审意见" min-width="220" />
        </el-table>
        <div v-if="scoredReviews.length" style="margin-top:8px;color:#667085">
          综合平均分：<b style="color:#002fa7">{{ app.avgScore }}</b>（{{ scoredReviews.length }}/{{ app.reviews.length }} 位专家已评分）
        </div>
        <div v-else style="margin-top:8px;color:#909399">暂无专家评分</div>
      </template>

      <!-- Final opinion -->
      <template v-if="app.finalOpinion">
        <el-divider content-position="left">评审结果</el-divider>
        <el-alert :closable="false" :title="app.finalOpinion" :type="app.status === 'rejected' ? 'error' : 'success'" />
      </template>

      <!-- Certificate -->
      <template v-if="app.certificate">
        <el-divider content-position="left">证书</el-divider>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="证书编号">{{ app.certificate.certNo || '-' }}</el-descriptions-item>
          <el-descriptions-item label="证书名称">{{ app.certificate.title || '-' }}</el-descriptions-item>
          <el-descriptions-item label="持有人">{{ app.certificate.holder || '-' }}</el-descriptions-item>
          <el-descriptions-item label="颁发时间">{{ fmt(app.certificate.issuedAt) }}</el-descriptions-item>
        </el-descriptions>
        <div style="margin-top:8px">
          <el-button v-if="app.certificate.fileUrl" size="small" type="primary" link @click="downloadCert">下载证书</el-button>
          <el-button size="small" type="warning" link @click="$router.push('/admin/certificates')">作废证书请前往证书管理</el-button>
        </div>
      </template>

      <!-- Actions -->
      <template v-if="actionsVisible">
        <el-divider content-position="left">操作</el-divider>
        <el-alert
          v-if="app.status === 'under_review' && reviewTotal === 0"
          style="margin-bottom:12px"
          type="warning"
          :closable="false"
          title="该申报当前没有评审人（或评审人已全部移除），可直接确定评审结果结案"
        />
        <div style="display:flex;gap:12px;flex-wrap:wrap;align-items:center">
          <template v-if="app.status === 'submitted'">
            <el-button type="success" @click="preliminary(true)">初审通过</el-button>
            <el-button type="danger" @click="preliminary(false)">初审驳回</el-button>
          </template>
          <template v-if="app.status === 'under_review' || app.status === 'reviewed'">
            <el-button type="primary" @click="openAssign">分配评审</el-button>
          </template>
          <template v-if="canDecide">
            <el-button type="success" @click="finalize(true)">确定通过</el-button>
            <el-button type="danger" @click="finalize(false)">确定不通过</el-button>
          </template>
          <template v-if="app.status === 'passed' || app.status === 'rejected'">
            <el-button v-if="!app.publishedAt" type="primary" @click="publishResult">公示结果</el-button>
            <el-tag v-else type="success">结果已公示</el-tag>
            <el-button type="warning" plain @click="revoke">撤回评审结果</el-button>
          </template>
          <template v-if="app.status === 'published'">
            <el-button type="warning" plain @click="revoke">撤回公示</el-button>
            <el-button type="primary" @click="openCertify">颁发证书</el-button>
          </template>
          <template v-if="app.status === 'certified'">
            <el-tag type="success">证书已颁发</el-tag>
            <el-button @click="$router.push('/admin/certificates')">前往证书管理</el-button>
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
      <el-alert
        v-if="app.status === 'under_review' && reviewTotal > 0"
        style="margin-bottom:12px"
        type="info"
        :closable="false"
        title="取消勾选可移除未评分的评审人；已有评分的评审人无法移除"
      />
      <el-select v-model="reviewerIds" multiple placeholder="请选择评审人（不选=清空评审人）" style="width:100%">
        <el-option v-for="r in reviewers" :key="r.id" :label="r.realName" :value="r.id" />
      </el-select>
      <template #footer>
        <el-button @click="assignDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmAssign">确定分配</el-button>
      </template>
    </el-dialog>

    <!-- 颁发证书：已公示后直接在申报详情内完成，不必再跳去证书管理 -->
    <el-dialog v-model="certDialog" title="颁发证书" width="560px">
      <el-form :model="certForm" label-width="90px">
        <el-form-item label="证书编号">
          <el-input v-model="certForm.certNo" placeholder="留空则自动生成（CAAM-年份-申报ID）" />
        </el-form-item>
        <el-form-item label="证书名称">
          <el-input v-model="certForm.title" placeholder="留空则使用项目名称" />
        </el-form-item>
        <el-form-item label="持有人">
          <el-input v-model="certForm.holder" placeholder="留空则使用申报人姓名" />
        </el-form-item>
        <el-form-item label="证书文件">
          <div style="display:flex;gap:8px;width:100%">
            <el-input v-model="certForm.fileUrl" placeholder="可上传或粘贴文件地址" />
            <el-upload :show-file-list="false" :http-request="uploadCert" accept=".pdf,.doc,.docx,.xls,.xlsx,.jpg,.jpeg,.png">
              <el-button :loading="certUploading">上传</el-button>
            </el-upload>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="certDialog = false">取消</el-button>
        <el-button type="primary" @click="submitCertify">颁发</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/admin'
import { applicationStatusMap, applicationStatusType, reviewStatusMap, fileUrl, fmt } from '@/utils/constants'

const route = useRoute()
const router = useRouter()
const id = Number(route.params.id)
const loading = ref(false)
const app = ref<any>({})

const opinionDialog = ref(false)
const opinion = ref('')
const pendingAction = ref('')

const assignDialog = ref(false)
const reviewers = ref<any[]>([])
const reviewerIds = ref<number[]>([])

const certDialog = ref(false)
const certUploading = ref(false)
const certForm = ref<any>({ certNo: '', title: '', holder: '', fileUrl: '' })

const actionsVisible = computed(() => {
  return ['submitted', 'under_review', 'reviewed', 'passed', 'rejected', 'published', 'certified'].includes(app.value.status)
})

// Only assignments that have actually been scored contribute to the average.
const scoredReviews = computed(() => (app.value.reviews || []).filter((r: any) => r.status === 'scored'))
const reviewTotal = computed(() => (app.value.reviews || []).length)
const pendingReviews = computed(() => (app.value.reviews || []).filter((r: any) => r.status !== 'scored'))

// 评审完成，或者根本没人评分（评审人全被移除/未分配）时都能直接确定结果，
// 否则这种申报会永远卡在「待评审」。
const canDecide = computed(() =>
  app.value.status === 'reviewed' ||
  (app.value.status === 'under_review' && pendingReviews.value.length === 0))

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
  const options = [...res.data]
  // Reviewers that are already assigned stay selectable even when their account
  // was disabled, otherwise the operator could neither see nor remove them and
  // the assignment request would fail validation.
  const known = new Set(options.map((r: any) => r.id))
  for (const r of app.value.reviews || []) {
    if (r.reviewer && !known.has(r.reviewerId)) {
      options.push({ ...r.reviewer, realName: `${r.reviewer.realName || r.reviewer.username}（已停用）` })
      known.add(r.reviewerId)
    }
  }
  reviewers.value = options
  reviewerIds.value = (app.value.reviews || []).map((r: any) => r.reviewerId)
  assignDialog.value = true
}

async function confirmAssign() {
  // 清空选择是合法的：它用于把无人评分的评审人全部移除，让申报可以结案。
  if (!reviewerIds.value.length && reviewTotal.value > 0) {
    await ElMessageBox.confirm('未选择任何评审人，将移除全部未评分的评审人，确认？', '提示', { type: 'warning' })
  }
  await adminApi.assignReviewers(id, { reviewerIds: reviewerIds.value })
  ElMessage.success(reviewerIds.value.length ? '分配成功' : '已清空评审人')
  assignDialog.value = false
  fetch()
}

async function publishResult() {
  await adminApi.publishResult(id)
  ElMessage.success('结果已公示')
  fetch()
}

// 撤回评审结果 / 撤回公示，让已确定的结果可以重新处理
async function revoke() {
  const isPublished = !!app.value.publishedAt
  await ElMessageBox.confirm(
    isPublished
      ? '撤回公示后，该结果将不再对申报人公开展示，确认撤回？'
      : '撤回评审结果后，申报将回到「评审完成」，需重新确定结果，确认撤回？',
    '提示',
    { type: 'warning' },
  )
  const res = await adminApi.revokeResult(id)
  ElMessage.success((res as any).message || '已撤回')
  fetch()
}

function openCertify() {
  certForm.value = {
    certNo: '',
    title: app.value.title || '',
    holder: app.value.user?.realName || '',
    fileUrl: '',
  }
  certDialog.value = true
}

async function uploadCert(option: any) {
  certUploading.value = true
  try {
    const res = await adminApi.uploadFile(option.file)
    certForm.value.fileUrl = res.data.fileUrl
    ElMessage.success('文件已上传')
  } finally {
    certUploading.value = false
  }
}

async function submitCertify() {
  await adminApi.issueCertificate({
    applicationId: id,
    certNo: certForm.value.certNo,
    title: certForm.value.title,
    holder: certForm.value.holder,
    fileUrl: certForm.value.fileUrl,
  })
  ElMessage.success('证书已颁发')
  certDialog.value = false
  fetch()
}

function downloadCert() {
  window.open(fileUrl(app.value.certificate?.fileUrl), '_blank')
}

function download(row: any) {
  window.open(fileUrl(row.fileUrl), '_blank')
}

onMounted(fetch)
</script>
