<template>
  <div class="page-card">
    <div class="page-toolbar">
      <el-input v-model="keyword" placeholder="搜索公示标题" clearable style="width:220px" @keyup.enter="fetch" @clear="fetch" />
      <el-button type="primary" @click="openDialog()">新增公示</el-button>
    </div>

    <el-table :data="list" v-loading="loading">
      <el-table-column prop="title" label="公示标题" min-width="220" />
      <el-table-column label="所属批次" min-width="160">
        <template #default="{ row }">{{ row.batch?.title || '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'published' ? 'success' : 'info'" size="small">{{ row.status === 'published' ? '已发布' : '草稿' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="openDialog(row)">编辑</el-button>
          <el-button v-if="row.status !== 'published'" size="small" type="success" @click="publish(row)">发布</el-button>
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

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑公示' : '新增公示'" width="640px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="公示标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="所属批次">
          <el-select v-model="form.batchId" style="width:100%" filterable>
            <el-option v-for="b in batches" :key="b.id" :label="b.title" :value="b.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="公示内容"><el-input v-model="form.content" type="textarea" :rows="6" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([])
const batches = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const keyword = ref('')
const loading = ref(false)
const dialogVisible = ref(false)
const form = reactive<any>({ id: 0, title: '', batchId: '', content: '' })

async function fetch() {
  loading.value = true
  try {
    const res = await adminApi.getAnnouncements({ page: page.value, pageSize, keyword: keyword.value })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }

function openDialog(row?: any) {
  Object.assign(form, row ? { ...row } : { id: 0, title: '', batchId: '', content: '' })
  dialogVisible.value = true
}

async function save() {
  const payload = { title: form.title, batchId: Number(form.batchId), content: form.content }
  if (form.id) await adminApi.updateAnnouncement(form.id, payload)
  else await adminApi.createAnnouncement(payload)
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetch()
}

async function publish(row: any) {
  await adminApi.publishAnnouncement(row.id)
  ElMessage.success('已发布')
  fetch()
}

async function remove(row: any) {
  await ElMessageBox.confirm('确认删除该公示？', '提示', { type: 'warning' })
  await adminApi.deleteAnnouncement(row.id)
  ElMessage.success('删除成功')
  fetch()
}

onMounted(async () => {
  fetch()
  const res = await adminApi.getBatches({ page: 1, pageSize: 100 })
  batches.value = res.data.list
})
</script>
