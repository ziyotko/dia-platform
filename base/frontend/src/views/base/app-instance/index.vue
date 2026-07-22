<template>
  <div class="app-instance-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>应用实例</span>
          <el-button type="primary" @click="handleAdd">开通应用</el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="app.name" label="应用名称" min-width="150" />
        <el-table-column prop="app.code" label="应用编码" min-width="120" />
        <el-table-column prop="tenantId" label="租户ID" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="config" label="配置" min-width="200" show-overflow-tooltip />
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑应用实例' : '开通应用'" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="应用" prop="appId">
          <el-select v-model="form.appId" placeholder="请选择应用" :disabled="isEdit" style="width: 100%">
            <el-option v-for="app in appList" :key="app.id" :label="app.name" :value="app.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="配置">
          <el-input v-model="form.config" type="textarea" :rows="5" placeholder="JSON 配置" />
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
import { getAppInstances, createAppInstance, updateAppInstance, deleteAppInstance } from '@/api/app'
import { getAppList } from '@/api/app'
import type { App } from '@/api/app'

interface AppInstance {
  id: number
  tenantId: number
  appId: number
  status: number
  config: string
  app?: App
}

const loading = ref(false)
const tableData = ref<AppInstance[]>([])
const appList = ref<App[]>([])
const total = ref(0)
const query = reactive({ page: 1, size: 10 })
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref()
const form = reactive<{
  id?: number
  tenantId: number
  appId: number | undefined
  status: number
  config: string
}>({
  tenantId: 0,
  appId: undefined,
  status: 1,
  config: ''
})

const rules = {
  appId: [{ required: true, message: '请选择应用', trigger: 'change' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getAppInstances(query)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    loading.value = false
  }
}

const fetchApps = async () => {
  try {
    const res: any = await getAppList({ page: 1, size: 1000 })
    appList.value = res.data.list || []
  } catch (error) {
    ElMessage.error('加载应用列表失败')
  }
}

const handleAdd = () => {
  isEdit.value = false
  Object.assign(form, {
    id: undefined,
    tenantId: 0,
    appId: undefined,
    status: 1,
    config: ''
  })
  dialogVisible.value = true
}

const handleEdit = (row: AppInstance) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    tenantId: row.tenantId,
    appId: row.appId,
    status: row.status,
    config: row.config
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  try {
    if (isEdit.value && form.id) {
      await updateAppInstance(form.id, {
        status: form.status,
        config: form.config
      })
    } else {
      await createAppInstance({
        tenantId: form.tenantId,
        appId: form.appId as number,
        status: form.status,
        config: form.config
      })
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    fetchData()
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const handleDelete = (row: AppInstance) => {
  ElMessageBox.confirm('确认删除该应用实例？', '提示', { type: 'warning' }).then(async () => {
    await deleteAppInstance(row.id)
    ElMessage.success('删除成功')
    fetchData()
  })
}

onMounted(() => {
  fetchData()
  fetchApps()
})
</script>

<style scoped lang="scss">
.app-instance-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
