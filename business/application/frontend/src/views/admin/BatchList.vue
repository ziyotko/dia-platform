<template>
  <div class="page-card">
    <div class="page-toolbar">
      <div style="display:flex;gap:12px">
        <el-input v-model="keyword" placeholder="搜索批次名称" clearable style="width:220px" @keyup.enter="fetch" @clear="fetch" />
        <el-select v-model="status" placeholder="全部状态" clearable style="width:140px" @change="fetch">
          <el-option v-for="(label, key) in batchStatusMap" :key="key" :label="label" :value="key" />
        </el-select>
        <el-button type="primary" @click="fetch">查询</el-button>
      </div>
      <el-button type="primary" @click="openDialog()">新增批次</el-button>
    </div>

    <el-table :data="list" v-loading="loading">
      <el-table-column prop="title" label="批次名称" min-width="200" />
      <el-table-column label="类别" width="130">
        <template #default="{ row }">{{ row.category?.name || '-' }}</template>
      </el-table-column>
      <el-table-column label="申报起止" min-width="240">
        <template #default="{ row }">{{ fmt(row.applyStart) }} ~ {{ fmt(row.applyEnd) }}</template>
      </el-table-column>
      <el-table-column label="评审截止" width="160">
        <template #default="{ row }">{{ fmt(row.reviewDeadline) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }"><el-tag :type="batchStatusType[row.status]">{{ batchStatusMap[row.status] }}</el-tag></template>
      </el-table-column>
      <el-table-column label="操作" width="300" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="openDialog(row)">编辑</el-button>
          <el-button v-if="row.status === 'draft'" size="small" type="success" @click="publish(row)">发布</el-button>
          <el-button v-if="row.status === 'open'" size="small" type="warning" @click="startReview(row)">开始评审</el-button>
          <el-button v-if="row.status === 'open' || row.status === 'reviewing'" size="small" type="info" @click="close(row)">结束</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      style="margin-top:16px;justify-content:flex-end"
      layout="total, prev, pager, next"
      :total="total" :page-size="pageSize" :current-page="page"
      @current-change="onPage"
    />

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑批次' : '新增批次'" width="640px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="批次名称"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="项目类别">
          <el-select v-model="form.categoryId" style="width:100%" :disabled="locked">
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="申报开始">
          <el-date-picker v-model="form.applyStart" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width:100%" :disabled="locked" />
        </el-form-item>
        <el-form-item label="申报截止">
          <el-date-picker v-model="form.applyEnd" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width:100%" :disabled="locked" />
        </el-form-item>
        <div v-if="locked" style="color:#909399;font-size:12px;margin:-8px 0 12px 100px">
          批次已发布，项目类别和申报时间不可修改
        </div>
        <el-form-item label="评审截止">
          <el-date-picker v-model="form.reviewDeadline" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width:100%" />
        </el-form-item>
        <el-form-item label="申报要求">
          <el-input v-model="form.requirements" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="批次说明">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/admin'
import { batchStatusMap, batchStatusType, fmt } from '@/utils/constants'

const list = ref<any[]>([])
const categories = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const keyword = ref('')
const status = ref('')
const loading = ref(false)
const dialogVisible = ref(false)
const form = reactive<any>({ id: 0, title: '', categoryId: '', applyStart: '', applyEnd: '', reviewDeadline: '', requirements: '', description: '' })

// The backend freezes category + application window once a batch is published;
// the same fields are locked here so the form never sends a rejected change.
const locked = computed(() => !!form.id && form.status !== 'draft')

async function fetch() {
  loading.value = true
  try {
    const res = await adminApi.getBatches({ page: page.value, pageSize, keyword: keyword.value, status: status.value })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }

// Converts an API datetime (ISO) into the value-format expected by the picker.
function toFormDate(v?: string) {
  return v ? v.replace('T', ' ').slice(0, 19) : ''
}

function openDialog(row?: any) {
  if (row) {
    Object.assign(form, {
      ...row,
      applyStart: toFormDate(row.applyStart),
      applyEnd: toFormDate(row.applyEnd),
      reviewDeadline: toFormDate(row.reviewDeadline),
    })
  } else {
    Object.assign(form, { id: 0, title: '', categoryId: '', applyStart: '', applyEnd: '', reviewDeadline: '', requirements: '', description: '' })
  }
  dialogVisible.value = true
}

async function save() {
  if (!form.title) return ElMessage.warning('请填写批次名称')
  if (!locked.value && !form.categoryId) return ElMessage.warning('请选择项目类别')
  const payload: any = {
    title: form.title,
    reviewDeadline: form.reviewDeadline || null,
    requirements: form.requirements,
    description: form.description,
  }
  // Only a draft batch may change its category / application window.
  if (!locked.value) {
    payload.categoryId = Number(form.categoryId)
    payload.applyStart = form.applyStart || null
    payload.applyEnd = form.applyEnd || null
  }
  if (form.id) await adminApi.updateBatch(form.id, payload)
  else await adminApi.createBatch(payload)
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetch()
}

// A batch can only be published with a complete application window; the backend
// enforces the same rule.
async function publish(row: any) {
  if (!row.applyStart || !row.applyEnd) {
    return ElMessage.warning('请先设置申报开始和截止时间')
  }
  await adminApi.publishBatch(row.id)
  ElMessage.success('已发布')
  fetch()
}
async function startReview(row: any) { await adminApi.startReview(row.id); ElMessage.success('已进入评审阶段'); fetch() }
async function close(row: any) { await adminApi.closeBatch(row.id); ElMessage.success('已结束'); fetch() }

async function remove(row: any) {
  await ElMessageBox.confirm('确认删除该批次？', '提示', { type: 'warning' })
  await adminApi.deleteBatch(row.id)
  ElMessage.success('删除成功')
  fetch()
}

onMounted(async () => {
  fetch()
  const res = await adminApi.getCategories()
  categories.value = res.data
})
</script>
