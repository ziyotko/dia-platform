<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="角色名称">
          <el-input v-model="queryForm.name" placeholder="请输入角色名称" clearable @keyup.enter="handleSearch" />
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
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" />
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
      width="600px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="90px"
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
      width="880px"
      top="8vh"
      destroy-on-close
      class="members-dialog"
      :close-on-click-modal="false"
    >
      <div class="members-header">
        <div class="members-role">
          <el-icon class="role-icon"><User /></el-icon>
          <span class="role-name">{{ currentRoleName }}</span>
        </div>
        <el-tag type="primary" size="large" effect="light" round>
          已选 {{ selectedUserIds.length }} 人
        </el-tag>
      </div>
      <el-transfer
        v-model="selectedUserIds"
        :data="userTransferData"
        :titles="['可选成员', '已选成员']"
        :props="{ key: 'id', label: 'label' }"
        filterable
        :filter-method="filterUser"
        filter-placeholder="请输入用户名 / 账号搜索"
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
import { formatDateTime } from '@/utils/format'
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
import { getAllUsers } from '@/api/user'

const loading = ref(false)
const dialogVisible = ref(false)
const membersVisible = ref(false)
const membersLoading = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const total = ref(0)
const formRef = ref()
const currentRoleId = ref<number>(0)
const currentRoleName = ref('')

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
    label: `${user.username || ''}（${user.account || ''}）`
  }))
})

const filterUser = (query: string, item: any) => {
  return item.label.toLowerCase().includes(query.toLowerCase())
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
        createdAt: formatDateTime(item.createdAt)
      }))
      total.value = res.data.total || 0
    }
  } catch {} finally {
    loading.value = false
  }
}

const fetchUsers = async () => {
  try {
    const res: any = await getAllUsers(1)
    if (res && res.code === 0) {
      userList.value = res.data.list || []
    }
  } catch {}
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
  ElMessageBox.confirm(`确定要删除流程角色 「${row.name}」 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res: any = await deleteWorkflowRole(row.id)
      if (res && res.code === 0) {
        ElMessage.success('删除成功')
        fetchData()
      }
    } catch {}
  }).catch(() => {})
}

const handleMembers = async (row: any) => {
  currentRoleId.value = row.id
  currentRoleName.value = row.name || ''
  membersVisible.value = true
  selectedUserIds.value = []
  await fetchUsers()
  try {
    const res: any = await getWorkflowRoleUsers(row.id)
    if (res && res.code === 0) {
      const users = res.data || []
      selectedUserIds.value = users.map((u: any) => u.id)
    }
  } catch {}
}

const handleMembersSubmit = async () => {
  if (!currentRoleId.value) return
  membersLoading.value = true
  try {
    const res: any = await updateWorkflowRoleUsers(currentRoleId.value, selectedUserIds.value)
    if (res && res.code === 0) {
      ElMessage.success('成员设置成功')
      membersVisible.value = false
    }
  } catch {} finally {
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
    }
  } catch {} finally {
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
  /* 列表页骨架（.search-card / .table-card / .card-header / .pagination）已统一到
     styles/global.scss 的 .page-container 规则，此处不再重复定义（2026-09-22） */

  :deep(.members-dialog) {
    .el-dialog__body {
      padding-top: 12px;
    }
  }

  .members-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding: 14px 18px;
    background: linear-gradient(135deg, #f0f7ff 0%, var(--app-brand-soft) 100%);
    border-radius: var(--app-card-radius);
    border: 1px solid #d0e6ff;

    .members-role {
      display: flex;
      align-items: center;
      gap: 10px;

      .role-icon {
        width: 36px;
        height: 36px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--el-color-primary);
        color: #fff;
        border-radius: 50%;
        font-size: 18px;
      }

      .role-name {
        font-size: 16px;
        font-weight: 600;
        color: var(--app-text-heading);
      }
    }
  }

  :deep(.el-transfer) {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 10px 0;

    .el-transfer-panel {
      width: 340px;
      height: 460px;
      border-radius: 12px;
      box-shadow: 0 2px 12px rgba(0, 47, 167, 0.08);
      border: 1px solid #e4e7ed;

      .el-transfer-panel__header {
        border-radius: 12px 12px 0 0;
        background: var(--el-fill-color-light);
        padding: 12px 16px;

        .el-checkbox__label {
          font-size: 15px;
          font-weight: 600;
          color: #303133;
        }
      }

      .el-transfer-panel__body {
        height: calc(100% - 50px);
      }

      .el-transfer-panel__list {
        height: calc(100% - 42px);
      }

      .el-transfer-panel__filter {
        margin: 12px;
        width: auto;

        .el-input__inner {
          border-radius: 8px;
        }
      }

      .el-transfer-panel__item {
        padding-left: 12px;
        margin-right: 0;
        height: 40px;
        line-height: 40px;

        .el-checkbox__label {
          font-size: 14px;
        }
      }
    }

    .el-transfer__buttons {
      padding: 0 24px;

      .el-button {
        width: 44px;
        height: 44px;
        border-radius: 50%;
        font-size: 18px;
      }
    }
  }
}
</style>
