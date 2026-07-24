<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="角色名称">
          <el-input v-model="queryForm.name" placeholder="请输入角色名称" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>查询
          </el-button>
          <el-button @click="resetQuery">
            <el-icon><RefreshRight /></el-icon>重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="hover" class="table-card">
      <template #header>
        <div class="card-header">
          <span>流程角色列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增流程角色
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="name" label="角色名称" min-width="140" />
        <el-table-column prop="code" label="角色编码" min-width="140" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleMembers(row)">
              <el-icon><User /></el-icon>成员
            </el-button>
            <el-button link type="primary" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button link type="danger" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="queryForm.page"
          v-model:page-size="queryForm.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="80px"
      >
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色编码" prop="code">
          <el-input v-model="form.code" placeholder="请输入角色编码" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="membersVisible"
      title="设置成员"
      width="600px"
      destroy-on-close
    >
      <el-transfer
        v-model="selectedUserIds"
        :data="userTransferData"
        :titles="['可选成员', '已选成员']"
        :props="{ key: 'id', label: 'label' }"
        filterable
        :filter-method="filterUser"
        filter-placeholder="请输入用户名/账号"
      />
      <template #footer>
        <el-button @click="membersVisible = false">取消</el-button>
        <el-button type="primary" :loading="membersLoading" @click="handleMembersSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  RefreshRight,
  Plus,
  Edit,
  Delete,
  User
} from '@element-plus/icons-vue'
import {
  getWorkflowRoleList,
  createWorkflowRole,
  updateWorkflowRole,
  deleteWorkflowRole,
  getWorkflowRoleUsers,
  updateWorkflowRoleUsers
} from '@/api/workflow-role'
import { getUserList } from '@/api/user'

const loading = ref(false)
const dialogVisible = ref(false)
const membersVisible = ref(false)
const membersLoading = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const total = ref(0)
const formRef = ref()
const currentRoleId = ref<number>(0)

const queryForm = reactive({
  page: 1,
  pageSize: 10,
  name: ''
})

const form = reactive({
  id: undefined as number | undefined,
  name: '',
  code: '',
  description: '',
  status: 1
})

const formRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入角色编码', trigger: 'blur' }]
}

const tableData = ref<any[]>([])
const userList = ref<any[]>([])
const selectedUserIds = ref<number[]>([])

const userTransferData = computed(() => {
  return userList.value.map(user => ({
    id: user.id,
    label: `${user.username || user.nickname || ''}（${user.account || ''}）`
  }))
})

const filterUser = (query: string, item: any) => {
  return item.label.toLowerCase().includes(query.toLowerCase())
}

const formatTime = (dateStr: string) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const handleSearch = () => {
  queryForm.page = 1
  fetchData()
}

const resetQuery = () => {
  queryForm.name = ''
  queryForm.page = 1
  fetchData()
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getWorkflowRoleList(queryForm)
    if (res && res.code === 0) {
      const list = res.data.list || []
      tableData.value = list.map((item: any) => ({
        ...item,
        createTime: formatTime(item.createdAt)
      }))
      total.value = res.data.total || 0
    } else {
      ElMessage.error(res?.message || '获取流程角色列表失败')
    }
  } catch (error) {
    ElMessage.error('获取流程角色列表失败')
  } finally {
    loading.value = false
  }
}

const fetchUsers = async () => {
  try {
    const res: any = await getUserList({ page: 1, pageSize: 9999, status: 1 })
    if (res && res.code === 0) {
      userList.value = res.data.list || []
    } else {
      ElMessage.error(res?.message || '获取用户列表失败')
    }
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  }
}

const handleAdd = () => {
  dialogTitle.value = '新增流程角色'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  dialogTitle.value = '编辑流程角色'
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除流程角色 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res: any = await deleteWorkflowRole(row.id)
      if (res && res.code === 0) {
        ElMessage.success('删除成功')
        fetchData()
      } else {
        ElMessage.error(res?.message || '删除失败')
      }
    } catch {
      ElMessage.error('删除失败')
    }
  })
}

const handleMembers = async (row: any) => {
  currentRoleId.value = row.id
  membersVisible.value = true
  selectedUserIds.value = []
  await fetchUsers()
  try {
    const res: any = await getWorkflowRoleUsers(row.id)
    if (res && res.code === 0) {
      const users = res.data || []
      selectedUserIds.value = users.map((u: any) => u.id)
    } else {
      ElMessage.error(res?.message || '获取成员失败')
    }
  } catch {
    ElMessage.error('获取成员失败')
  }
}

const handleMembersSubmit = async () => {
  if (!currentRoleId.value) return
  membersLoading.value = true
  try {
    const res: any = await updateWorkflowRoleUsers(currentRoleId.value, selectedUserIds.value)
    if (res && res.code === 0) {
      ElMessage.success('成员设置成功')
      membersVisible.value = false
    } else {
      ElMessage.error(res?.message || '成员设置失败')
    }
  } catch {
    ElMessage.error('成员设置失败')
  } finally {
    membersLoading.value = false
  }
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    const payload = {
      name: form.name,
      code: form.code,
      description: form.description,
      status: form.status
    }
    let res: any
    if (form.id) {
      res = await updateWorkflowRole(form.id, payload)
    } else {
      res = await createWorkflowRole(payload)
    }
    if (res && res.code === 0) {
      ElMessage.success(form.id ? '修改成功' : '新增成功')
      dialogVisible.value = false
      fetchData()
    } else {
      ElMessage.error(res?.message || (form.id ? '修改失败' : '新增失败'))
    }
  } catch {
    ElMessage.error(form.id ? '修改失败' : '新增失败')
  } finally {
    submitLoading.value = false
  }
}

const resetForm = () => {
  form.id = undefined
  form.name = ''
  form.code = ''
  form.description = ''
  form.status = 1
}

const handleSizeChange = (val: number) => {
  queryForm.pageSize = val
  fetchData()
}

const handleCurrentChange = (val: number) => {
  queryForm.page = val
  fetchData()
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.page-container {
  .search-card {
    margin-bottom: 20px;
    border-radius: 12px;
    border: 1px solid #e6f2ff;
  }

  .table-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-weight: 600;
      color: #2c3e50;
    }
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  :deep(.el-transfer) {
    display: flex;
    justify-content: center;
    align-items: center;

    .el-transfer-panel {
      width: 220px;
    }
  }
}
</style>
