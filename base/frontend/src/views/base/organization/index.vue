<template>
  <div class="organization-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>机构管理</span>
          <el-button v-if="can('base:org:create')" type="primary" @click="handleAdd">新增机构</el-button>
        </div>
      </template>
      <el-table :data="tableData" v-loading="loading" row-key="id" border default-expand-all empty-text="暂无数据">
        <el-table-column prop="name" label="机构名称" width="220" />
        <el-table-column prop="code" label="机构编码" />
        <el-table-column prop="leader" label="负责人" />
        <el-table-column prop="phone" label="联系电话" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-if="can('base:org:update')" link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button v-if="can('base:org:delete')" link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑机构' : '新增机构'" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="上级机构">
          <el-tree-select
            v-model="form.parentId"
            :data="orgOptions"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            check-strictly
            clearable
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="机构编码" prop="code">
          <el-input v-model="form.code" />
        </el-form-item>
        <el-form-item label="机构名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="负责人">
          <el-input v-model="form.leader" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
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
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrganizationTree, createOrganization, updateOrganization, deleteOrganization } from '@/api/organization'
import type { Organization } from '@/api/organization'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
// 按钮级权限：与后端 base:org:* 权限点对齐
const can = (code: string) => userStore.can(code)

const loading = ref(false)
const tableData = ref<Organization[]>([])
const orgOptions = ref<Organization[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<any>(null)
const form = reactive<Organization>({
  id: 0,
  parentId: 0,
  code: '',
  name: '',
  leader: '',
  phone: '',
  email: '',
  sort: 0,
  status: 1,
  description: ''
})

const rules = {
  code: [{ required: true, message: '请输入机构编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入机构名称', trigger: 'blur' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getOrganizationTree()
    const data = res.data || []
    tableData.value = data
    orgOptions.value = [{ id: 0, parentId: 0, code: '', name: '根机构', sort: 0, status: 1 }, ...data]
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: Organization) => {
  isEdit.value = true
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row: Organization) => {
  await ElMessageBox.confirm('确认删除该机构？', '提示', { type: 'warning' })
  await deleteOrganization(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (isEdit.value) {
    await updateOrganization(form.id, form)
  } else {
    await createOrganization(form)
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetchData()
}

const resetForm = () => {
  form.id = 0
  form.parentId = 0
  form.code = ''
  form.name = ''
  form.leader = ''
  form.phone = ''
  form.email = ''
  form.sort = 0
  form.status = 1
  form.description = ''
}

onMounted(fetchData)
</script>

<style scoped lang="scss">
.organization-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>
