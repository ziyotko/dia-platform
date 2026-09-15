<template>
  <div class="page-card" style="max-width:900px">
    <h3 style="margin-bottom:16px">项目申报</h3>
    <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
      <el-form-item label="申报批次" prop="batchId">
        <el-select v-model="form.batchId" placeholder="请选择申报批次" style="width:100%" @change="onBatchChange">
          <el-option v-for="b in batches" :key="b.id" :label="b.title" :value="b.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="项目类别" prop="categoryId">
        <el-select v-model="form.categoryId" placeholder="请先选择申报批次" style="width:100%" disabled>
          <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
        <div class="form-hint">项目类别由申报批次决定，不可修改</div>
      </el-form-item>
      <el-form-item label="项目名称" prop="title">
        <el-input v-model="form.title" placeholder="请输入项目名称" maxlength="100" show-word-limit />
      </el-form-item>
      <el-form-item label="项目简介">
        <el-input v-model="form.projectBrief" type="textarea" :rows="3" placeholder="项目简介（200字以内）" />
      </el-form-item>
      <el-form-item label="申报内容">
        <el-input v-model="form.content" type="textarea" :rows="6" placeholder="项目背景、目标、实施方案、预期成果等" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="save" :loading="loading">保存草稿</el-button>
        <el-button @click="$router.back()">返回</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { memberApi } from '@/api/member'

const route = useRoute()
const router = useRouter()
const formRef = ref()
const loading = ref(false)
const batches = ref<any[]>([])
const categories = ref<any[]>([])

const form = reactive({
  batchId: route.query.batchId ? Number(route.query.batchId) : ('' as any),
  categoryId: '' as any,
  title: '',
  projectBrief: '',
  content: '',
})

const rules = {
  batchId: [{ required: true, message: '请选择申报批次', trigger: 'change' }],
  title: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
}

function onBatchChange(batchId: number) {
  const batch = batches.value.find((b) => b.id === batchId)
  form.categoryId = batch?.categoryId || ''
}

async function save() {
  await formRef.value?.validate()
  if (!form.categoryId) return ElMessage.warning('该批次未设置项目类别，请联系管理员')
  loading.value = true
  try {
    const res = await memberApi.createApplication({
      batchId: Number(form.batchId),
      categoryId: Number(form.categoryId),
      title: form.title,
      projectBrief: form.projectBrief,
      content: form.content,
    })
    ElMessage.success('草稿已保存，请上传材料后提交')
    router.push(`/member/applications/${res.data.id}`)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  const [b, c] = await Promise.all([memberApi.getBatches({ page: 1, pageSize: 100 }), memberApi.getCategories()])
  // 只能选当前仍在申报期内的批次（列表中同时存在评审中/已结束的批次）
  batches.value = (b.data.list || []).filter((item: any) => item.canApply)
  categories.value = c.data
  // Arriving from 「立即申报」: derive the locked category right away.
  if (form.batchId) onBatchChange(Number(form.batchId))
})
</script>

<style scoped>
.form-hint { font-size: 12px; color: #909399; line-height: 1.6; }
</style>
