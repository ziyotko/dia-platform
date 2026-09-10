<template>
  <div v-loading="loading" class="page-card" style="max-width:900px">
    <h3 style="margin-bottom:16px">专家评审</h3>

    <el-descriptions :column="2" border>
      <el-descriptions-item label="项目名称" :span="2">{{ app.application?.title || '-' }}</el-descriptions-item>
      <el-descriptions-item label="所属批次">{{ app.application?.batch?.title || '-' }}</el-descriptions-item>
      <el-descriptions-item label="项目类别">{{ app.application?.category?.name || '-' }}</el-descriptions-item>
      <el-descriptions-item label="项目简介" :span="2">{{ app.application?.projectBrief || '-' }}</el-descriptions-item>
      <el-descriptions-item label="申报内容" :span="2">{{ app.application?.content || '-' }}</el-descriptions-item>
    </el-descriptions>

    <el-divider content-position="left">申报材料</el-divider>
    <el-table :data="app.application?.materials || []" size="small">
      <el-table-column prop="name" label="材料名称" min-width="220" />
      <el-table-column label="操作" width="120">
        <template #default="{ row }"><el-button size="small" type="primary" link @click="download(row)">下载</el-button></template>
      </el-table-column>
    </el-table>

    <el-divider content-position="left">评审打分</el-divider>
    <el-form label-width="90px">
      <el-form-item label="评分">
        <el-input-number v-model="score" :min="0" :max="100" :precision="1" />
        <span style="margin-left:8px;color:#999">0-100 分</span>
      </el-form-item>
      <el-form-item label="评审意见">
        <el-input v-model="comment" type="textarea" :rows="5" placeholder="请填写评审意见" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="submit" :loading="submitting" :disabled="app.status === 'scored'">提交评审</el-button>
        <el-button @click="$router.back()">返回</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'
import { fileUrl } from '@/utils/constants'

const route = useRoute()
const id = Number(route.params.id)
const loading = ref(false)
const submitting = ref(false)
const app = ref<any>({})
const score = ref(0)
const comment = ref('')

async function fetch() {
  loading.value = true
  try {
    const res = await adminApi.getReview(id)
    app.value = res.data
    score.value = res.data.score || 0
    comment.value = res.data.comment || ''
  } finally {
    loading.value = false
  }
}

async function submit() {
  submitting.value = true
  try {
    await adminApi.submitReview(id, { score: score.value, comment: comment.value })
    ElMessage.success('评审提交成功')
    fetch()
  } finally {
    submitting.value = false
  }
}

function download(row: any) {
  window.open(fileUrl(row.fileUrl), '_blank')
}

onMounted(fetch)
</script>
