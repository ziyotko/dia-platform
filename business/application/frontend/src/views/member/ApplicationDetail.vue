<template>
  <div v-loading="loading">
    <div class="page-card" style="max-width:960px">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
        <h3>{{ app.title || '申报详情' }}</h3>
        <el-tag :type="applicationStatusType[app.status]" size="large">{{ applicationStatusMap[app.status] || app.status }}</el-tag>
      </div>

      <!-- Progress timeline -->
      <el-steps :active="activeStep" align-center style="margin:24px 0" finish-status="success">
        <el-step title="提交申报" />
        <el-step title="在线初审" />
        <el-step title="专家评审" />
        <el-step title="评审结果" />
        <el-step title="结果公示" />
        <el-step title="颁发证书" />
      </el-steps>

      <!-- Basic info (editable in draft / after preliminary rejection) -->
      <el-descriptions :column="2" border v-if="!isEditable">
        <el-descriptions-item label="申报批次">{{ app.batch?.title || '-' }}</el-descriptions-item>
        <el-descriptions-item label="项目类别">{{ app.category?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="项目名称">{{ app.title }}</el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ fmt(app.submittedAt) }}</el-descriptions-item>
        <el-descriptions-item label="项目简介" :span="2">{{ app.projectBrief || '-' }}</el-descriptions-item>
        <el-descriptions-item label="申报内容" :span="2">{{ app.content || '-' }}</el-descriptions-item>
      </el-descriptions>

      <el-form v-else :model="editForm" label-width="90px">
        <el-form-item label="项目类别">
          <el-select v-model="editForm.categoryId" placeholder="-" style="width:100%" disabled>
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
          <div class="form-hint">项目类别由申报批次决定，不可修改</div>
        </el-form-item>
        <el-form-item label="项目名称"><el-input v-model="editForm.title" /></el-form-item>
        <el-form-item label="项目简介"><el-input v-model="editForm.projectBrief" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="申报内容"><el-input v-model="editForm.content" type="textarea" :rows="5" /></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveInfo">保存信息</el-button>
        </el-form-item>
      </el-form>

      <!-- Materials -->
      <el-divider content-position="left">申报材料</el-divider>
      <el-table :data="materials" size="small">
        <el-table-column prop="name" label="材料名称" min-width="220" />
        <el-table-column prop="fileType" label="类型" width="90" />
        <el-table-column label="大小" width="110">
          <template #default="{ row }">{{ (row.fileSize / 1024).toFixed(1) }} KB</template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="download(row)">下载</el-button>
            <el-button v-if="isEditable" size="small" type="danger" link @click="removeMaterial(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="isEditable" style="margin-top:12px">
        <el-upload :show-file-list="false" :http-request="handleUpload" accept=".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.zip,.rar,.jpg,.jpeg,.png,.gif">
          <el-button size="small">上传材料</el-button>
        </el-upload>
      </div>

      <!-- Preliminary opinion -->
      <template v-if="app.preliminaryOpinion">
        <el-divider content-position="left">初审意见</el-divider>
        <el-alert :closable="false" :title="app.preliminaryOpinion" :type="app.status === 'preliminary_rejected' ? 'error' : 'info'" />
        <div v-if="app.status === 'preliminary_rejected'" style="margin-top:8px;color:#c45656">
          初审未通过，请修改申报信息或材料后重新提交（需在申报截止前）。
        </div>
      </template>

      <!-- Review scores -->
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
        <div v-if="app.publishedAt" style="margin-top:8px;color:#667085">公示时间：{{ fmt(app.publishedAt) }}</div>
      </template>

      <div v-if="isEditable" style="margin-top:20px;text-align:center">
        <el-button type="primary" size="large" @click="submit" :loading="submitting">
          {{ app.status === 'preliminary_rejected' ? '修改后重新提交' : '提交申报' }}
        </el-button>
      </div>
      <div v-else-if="app.status === 'submitted'" style="margin-top:20px;text-align:center">
        <el-button size="large" @click="withdraw" :loading="withdrawing">撤回申报</el-button>
        <div class="form-hint" style="margin-top:8px">撤回后申报将回到草稿状态，可继续修改后重新提交</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { memberApi } from '@/api/member'
import { applicationStatusMap, applicationStatusType, fileUrl, fmt } from '@/utils/constants'

const route = useRoute()
const id = Number(route.params.id)
const loading = ref(false)
const submitting = ref(false)
const withdrawing = ref(false)
const app = ref<any>({})
const materials = ref<any[]>([])
const categories = ref<any[]>([])

const editForm = reactive({ categoryId: '' as any, title: '', projectBrief: '', content: '' })

// Drafts and preliminary-rejected applications can still be edited / resubmitted.
const isEditable = computed(() => app.value.status === 'draft' || app.value.status === 'preliminary_rejected')

const activeStep = computed(() => {
  const map: Record<string, number> = {
    draft: 0, submitted: 1, preliminary_rejected: 1, under_review: 2, reviewed: 2,
    passed: 3, rejected: 3, published: 4, certified: 5,
  }
  return map[app.value.status] ?? 0
})

async function fetch() {
  loading.value = true
  try {
    const res = await memberApi.getApplication(id)
    app.value = res.data
    materials.value = res.data.materials || []
    editForm.categoryId = res.data.categoryId
    editForm.title = res.data.title
    editForm.projectBrief = res.data.projectBrief
    editForm.content = res.data.content
  } finally {
    loading.value = false
  }
}

async function saveInfo() {
  if (!editForm.title) return ElMessage.warning('请填写项目名称')
  // categoryId is fixed by the batch and intentionally not sent back.
  await memberApi.updateApplication(id, {
    title: editForm.title,
    projectBrief: editForm.projectBrief,
    content: editForm.content,
  })
  ElMessage.success('保存成功')
  fetch()
}

async function handleUpload(option: any) {
  const res = await memberApi.uploadFile(option.file)
  materials.value.push({
    name: res.data.name, fileUrl: res.data.fileUrl, fileType: res.data.fileType, fileSize: res.data.fileSize,
  })
  await memberApi.saveMaterials(id, materials.value)
  ElMessage.success('材料已上传')
}

function removeMaterial(row: any) {
  materials.value = materials.value.filter((m) => m !== row)
  memberApi.saveMaterials(id, materials.value).then(() => ElMessage.success('材料已删除'))
}

async function submit() {
  if (!materials.value.length) return ElMessage.warning('请先上传申报材料')
  submitting.value = true
  try {
    await memberApi.submitApplication(id)
    ElMessage.success('提交成功，等待初审')
    fetch()
  } finally {
    submitting.value = false
  }
}

async function withdraw() {
  withdrawing.value = true
  try {
    await memberApi.withdrawApplication(id)
    ElMessage.success('已撤回，可继续修改')
    fetch()
  } finally {
    withdrawing.value = false
  }
}

function download(row: any) {
  window.open(fileUrl(row.fileUrl), '_blank')
}

onMounted(async () => {
  await fetch()
  const res = await memberApi.getCategories()
  categories.value = res.data
})
</script>

<style scoped>
.form-hint { font-size: 12px; color: #909399; line-height: 1.6; }
</style>
