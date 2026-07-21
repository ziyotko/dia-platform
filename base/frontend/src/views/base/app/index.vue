<template>
  <div class="app-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>应用管理</span>
          <el-button type="primary" @click="handleAdd">新增应用</el-button>
        </div>
      </template>
      <el-table :data="tableData" v-loading="loading" border>
        <el-table-column prop="code" label="应用编码" width="140" />
        <el-table-column prop="name" label="应用名称" />
        <el-table-column prop="type" label="接入方式" width="120">
          <template #default="{ row }">
            <el-tag>{{ typeMap[row.type] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="frontendUrl" label="前端入口" show-overflow-tooltip />
        <el-table-column prop="backendUrl" label="后端入口" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.size"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑应用' : '新增应用'" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="应用编码" prop="code">
          <el-input v-model="form.code" :disabled="isEdit" placeholder="唯一编码，如 caam-portal" />
        </el-form-item>
        <el-form-item label="应用名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="接入方式" prop="type">
          <el-select v-model="form.type" placeholder="请选择" style="width: 100%">
            <el-option label="IFrame" value="iframe" />
            <el-option label="API代理" value="proxy" />
            <el-option label="微应用" value="micro" />
          </el-select>
        </el-form-item>
        <el-form-item label="前端入口">
          <el-input v-model="form.frontendUrl" placeholder="http://localhost:5173" />
        </el-form-item>
        <el-form-item label="后端入口">
          <el-input v-model="form.backendUrl" placeholder="http://localhost:8080" />
        </el-form-item>
        <el-form-item label="API前缀">
          <el-input v-model="form.apiPrefix" placeholder="/caamm/api" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAppList, createApp, updateApp, deleteApp } from '@/api/app'
import type { App } from '@/api/app'

const loading = ref(false)
const tableData = ref<App[]>([])
const total = ref(0)
const query = reactive({ page: 1, size: 10, keyword: '' })
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<any>(null)
const form = reactive<App>({
  id: 0,
  code: '',
  name: '',
  type: 'iframe',
  frontendUrl: '',
  backendUrl: '',
  apiPrefix: '',
  status: 1,
  sort: 0,
  description: ''
})

const typeMap: Record<string, string> = {
  iframe: 'IFrame',
  proxy: 'API代理',
  micro: '微应用'
}

const rules = {
  code: [{ required: true, message: '请输入应用编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入应用名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择接入方式', trigger: 'change' }]
}

const fetchData = async () => {
  loading.value = true
  const res: any = await getAppList(query)
  tableData.value = res.data.list
  total.value = res.data.total
  loading.value = false
}

const handleAdd = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: App) => {
  isEdit.value = true
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row: App) => {
  await ElMessageBox.confirm('确认删除该应用？', '提示', { type: 'warning' })
  await deleteApp(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (isEdit.value) {
    await updateApp(form.id, form)
  } else {
    await createApp(form)
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetchData()
}

const resetForm = () => {
  form.id = 0
  form.code = ''
  form.name = ''
  form.type = 'iframe'
  form.frontendUrl = ''
  form.backendUrl = ''
  form.apiPrefix = ''
  form.status = 1
  form.sort = 0
  form.description = ''
}

onMounted(fetchData)
</script>

<style scoped lang="scss">
.app-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
