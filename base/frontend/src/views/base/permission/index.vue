<template>
  <div class="permission-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>权限管理</span>
          <el-button v-if="can('base:permission:create')" type="primary" @click="handleAdd">新增权限</el-button>
        </div>
      </template>

      <el-alert
        title="接口权限点用于控制「哪些接口可以被调用」；在「角色管理 → 分配权限」中把权限授予角色。"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 16px"
      />

      <el-form :inline="true" class="search-form">
        <el-form-item label="应用编码">
          <el-input v-model="appCode" placeholder="如 base" clearable style="width: 180px" @keyup.enter="fetchData" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchData">查询</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" row-key="id" border default-expand-all>
        <el-table-column prop="name" label="权限名称" min-width="180" />
        <el-table-column prop="code" label="权限编码" min-width="200" />
        <el-table-column prop="type" label="类型" width="90">
          <template #default="{ row }">
            <el-tag :type="typeTag[row.type] || 'info'">{{ typeMap[row.type] || row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="method" label="方法" width="90" />
        <el-table-column prop="path" label="接口路径" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button v-if="can('base:permission:update')" link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button v-if="can('base:permission:delete')" link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑权限' : '新增权限'" width="620px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="应用编码" prop="appCode">
          <el-input v-model="form.appCode" placeholder="base 或子应用编码" />
        </el-form-item>
        <el-form-item label="上级权限">
          <el-tree-select
            v-model="form.parentId"
            :data="parentOptions"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            check-strictly
            clearable
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="权限名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="权限编码" prop="code">
          <el-input v-model="form.code" placeholder="如 base:user:create" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio-button label="menu">分组</el-radio-button>
            <el-radio-button label="api">接口</el-radio-button>
            <el-radio-button label="button">按钮</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <template v-if="form.type === 'api'">
          <el-form-item label="请求方法" prop="method">
            <el-select v-model="form.method" placeholder="请选择" style="width: 100%">
              <el-option v-for="m in methods" :key="m" :label="m" :value="m" />
            </el-select>
          </el-form-item>
          <el-form-item label="接口路径" prop="path">
            <el-input v-model="form.path" placeholder="相对路径，如 /users/:id" />
          </el-form-item>
        </template>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
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
import { getPermissionTree, createPermission, updatePermission, deletePermission } from '@/api/permission'
import type { Permission } from '@/api/permission'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
// 按钮级权限：与后端 base:permission:* 权限点对齐
const can = (code: string) => userStore.can(code)

const methods = ['GET', 'POST', 'PUT', 'DELETE']
const typeMap: Record<string, string> = { menu: '分组', api: '接口', button: '按钮' }
const typeTag: Record<string, string> = { menu: 'info', api: 'success', button: 'warning' }

const loading = ref(false)
const appCode = ref('base')
const tableData = ref<Permission[]>([])
const parentOptions = ref<any[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<any>(null)
const form = reactive<Permission>({
  id: 0,
  appCode: 'base',
  code: '',
  name: '',
  type: 'api',
  parentId: 0,
  path: '',
  method: 'GET',
  status: 1
})

const rules = {
  appCode: [{ required: true, message: '请输入应用编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入权限编码', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getPermissionTree(appCode.value || undefined)
    tableData.value = res.data || []
    parentOptions.value = [{ id: 0, name: '顶级权限', children: [] }, ...tableData.value]
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  isEdit.value = false
  Object.assign(form, {
    id: 0,
    appCode: appCode.value || 'base',
    code: '',
    name: '',
    type: 'api',
    parentId: 0,
    path: '',
    method: 'GET',
    status: 1
  })
  dialogVisible.value = true
}

const handleEdit = (row: Permission) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    appCode: row.appCode,
    code: row.code,
    name: row.name,
    type: row.type,
    parentId: row.parentId,
    path: row.path || '',
    method: row.method || 'GET',
    status: row.status
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  const payload: Permission = {
    ...form,
    // 分组/按钮类权限不参与接口匹配，方法与前缀清空避免误命中
    method: form.type === 'api' ? form.method || 'GET' : '',
    path: form.type === 'api' ? form.path || '' : ''
  }
  if (isEdit.value) {
    await updatePermission(form.id, payload)
  } else {
    await createPermission(payload)
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetchData()
}

const handleDelete = async (row: Permission) => {
  await ElMessageBox.confirm(
    row.children?.length
      ? '该权限下存在子权限，删除后子权限将失去归属，确认删除？'
      : '确认删除该权限？',
    '提示',
    { type: 'warning' }
  )
  await deletePermission(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(fetchData)
</script>

<style scoped lang="scss">
.permission-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .search-form {
    margin-bottom: 16px;
  }
}
</style>
