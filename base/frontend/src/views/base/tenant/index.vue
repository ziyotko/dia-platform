<template>
  <div class="tenant-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>租户管理</span>
          <el-button v-if="isSuperAdmin" type="primary" @click="handleAdd">新增租户</el-button>
        </div>
      </template>

      <el-alert
        v-if="!isSuperAdmin"
        title="租户管理为平台级功能，仅平台超级管理员可查看与维护"
        type="warning"
        :closable="false"
        show-icon
        style="margin-bottom: 16px"
      />

      <el-form :inline="true" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="租户名称 / 编码" clearable style="width: 220px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" border>
        <el-table-column prop="code" label="租户编码" width="140" />
        <el-table-column prop="name" label="租户名称" />
        <el-table-column prop="contactName" label="联系人" />
        <el-table-column prop="contactPhone" label="联系电话" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-if="isSuperAdmin" link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button v-if="isSuperAdmin" link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑租户' : '新增租户'" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="租户编码" prop="code">
          <el-input v-model="form.code" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="租户名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="form.contactName" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.contactPhone" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
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
import { reactive, ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTenantList, createTenant, updateTenant, deleteTenant } from '@/api/tenant'
import type { Tenant } from '@/api/tenant'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
// /tenants 整组接口仅平台超管可用，非超管页面上不应出现必然 403 的写入口
const isSuperAdmin = computed(() => userStore.userInfo?.tenantId === 0)

const loading = ref(false)
const tableData = ref<Tenant[]>([])
const total = ref(0)
const query = reactive({ page: 1, size: 10, keyword: '' })
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<any>(null)
const form = reactive<Tenant>({
  id: 0,
  code: '',
  name: '',
  status: 1,
  contactName: '',
  contactPhone: '',
  description: ''
})

const rules = {
  code: [{ required: true, message: '请输入租户编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入租户名称', trigger: 'blur' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getTenantList(query)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    // 必须放 finally：请求失败（如无权限 403）时否则 loading 永远为 true
    loading.value = false
  }
}

const handleSearch = () => {
  query.page = 1
  fetchData()
}

const handleReset = () => {
  query.keyword = ''
  query.page = 1
  fetchData()
}

const handleAdd = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: Tenant) => {
  isEdit.value = true
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row: Tenant) => {
  await ElMessageBox.confirm('确认删除该租户？', '提示', { type: 'warning' })
  await deleteTenant(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (isEdit.value) {
    await updateTenant(form.id, form)
  } else {
    await createTenant(form)
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetchData()
}

const resetForm = () => {
  form.id = 0
  form.code = ''
  form.name = ''
  form.status = 1
  form.contactName = ''
  form.contactPhone = ''
  form.description = ''
}

onMounted(fetchData)
</script>

<style scoped lang="scss">
.tenant-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .search-form {
    margin-bottom: 16px;
  }
  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
