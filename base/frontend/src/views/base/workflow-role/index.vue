<template>
  <div class="workflow-role-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>流程角色</span>
          <el-button v-if="can('base:workflow-role:create')" type="primary" @click="handleAdd">新增流程角色</el-button>
        </div>
      </template>

      <el-alert
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 16px"
        title="流程角色用于给审批流程的节点指定「谁有资格审批」，与系统角色（决定菜单与接口权限）互相独立"
      />

      <el-form :inline="true" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="角色名称 / 编码" clearable style="width: 200px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" placeholder="全部" clearable style="width: 140px">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isSuperAdmin" label="所属租户">
          <tenant-select v-model="query.tenantId" placeholder="全部租户" clearable style="width: 220px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" border>
        <el-table-column prop="code" label="角色编码" min-width="140" />
        <el-table-column prop="name" label="角色名称" min-width="140" />
        <el-table-column v-if="isSuperAdmin" label="所属租户" min-width="140">
          <template #default="{ row }">{{ tenantName(row.tenantId) }}</template>
        </el-table-column>
        <el-table-column label="成员数" width="100">
          <template #default="{ row }">
            <el-tag :type="row.userCount ? 'info' : 'warning'">{{ row.userCount || 0 }} 人</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="角色说明" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button v-if="can('base:workflow-role:update')" link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button v-if="can('base:workflow-role:assign-user')" link type="primary" @click="handleMembers(row)">成员配置</el-button>
            <el-button v-if="can('base:workflow-role:delete')" link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑流程角色' : '新增流程角色'" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item v-if="isSuperAdmin" label="所属租户" prop="tenantId">
          <tenant-select v-model="form.tenantId" :disabled="isEdit" :clearable="!isEdit" />
          <div class="form-tip">编辑时不允许变更所属租户</div>
        </el-form-item>
        <el-form-item label="角色编码" prop="code">
          <el-input v-model="form.code" placeholder="留空自动生成" :disabled="isEdit" />
          <div class="form-tip">租户内唯一，供审批流程节点引用</div>
        </el-form-item>
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="如：审核组 / 复审组" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="角色说明">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="说明该角色的审批职责" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="memberDialogVisible" title="成员配置" width="620px">
      <el-alert
        title="角色成员即为该角色的审批人候选；被停用的用户仍会保留在角色中，但不应再承担审批"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 12px"
      />
      <div class="member-header">
        <span class="member-role">{{ currentRole.name }}</span>
        <el-input
          v-model="memberKeyword"
          placeholder="搜索用户名 / 姓名"
          clearable
          style="width: 220px"
          @keyup.enter="loadUserOptions"
        />
      </div>
      <el-select
        v-model="selectedUserIds"
        multiple
        filterable
        collapse-tags
        collapse-tags-tooltip
        placeholder="请选择成员"
        style="width: 100%"
        v-loading="memberLoading"
      >
        <el-option
          v-for="user in userOptions"
          :key="user.id"
          :label="userLabel(user)"
          :value="user.id"
          :disabled="user.status !== 1"
        />
      </el-select>
      <div class="form-tip">已选 {{ selectedUserIds.length }} 人；停用用户不可选，如需使用请先在「用户管理」中启用</div>
      <template #footer>
        <el-button @click="memberDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveMembers">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getWorkflowRoleList,
  getWorkflowRole,
  createWorkflowRole,
  updateWorkflowRole,
  deleteWorkflowRole,
  assignWorkflowRoleUsers,
  getWorkflowRoleUserOptions
} from '@/api/workflowRole'
import type { UserOption, WorkflowRole } from '@/api/workflowRole'
import { tenantName, ensureTenants } from '@/utils/tenantOptions'
import { useUserStore } from '@/stores/user'
import TenantSelect from '@/components/TenantSelect.vue'

const userStore = useUserStore()
// 仅平台超级管理员（tenantId === 0）可以跨租户管理
const isSuperAdmin = computed(() => userStore.userInfo?.tenantId === 0)
// 按钮级权限：与后端 base:workflow-role:* 权限点对齐
const can = (code: string) => userStore.can(code)

const loading = ref(false)
const tableData = ref<WorkflowRole[]>([])
const total = ref(0)
const query = reactive<{ page: number; size: number; keyword: string; status?: number; tenantId?: number }>({
  page: 1,
  size: 10,
  keyword: '',
  status: undefined,
  tenantId: undefined
})

const dialogVisible = ref(false)
const memberDialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<any>(null)
const currentRole = ref<WorkflowRole>({ id: 0, code: '', name: '', status: 1 })
const selectedUserIds = ref<number[]>([])
const userOptions = ref<UserOption[]>([])
const memberKeyword = ref('')
const memberLoading = ref(false)

const form = reactive<WorkflowRole>({
  id: 0,
  tenantId: 0,
  code: '',
  name: '',
  description: '',
  status: 1
})

const rules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }]
}

const userLabel = (user: UserOption) =>
  `${user.realName ? user.realName + '（' + user.username + '）' : user.username}${user.status !== 1 ? ' - 已停用' : ''}`

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getWorkflowRoleList(query)
    tableData.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  query.page = 1
  fetchData()
}

const handleReset = () => {
  query.keyword = ''
  query.status = undefined
  query.tenantId = undefined
  query.page = 1
  fetchData()
}

const resetForm = () => {
  form.id = 0
  form.tenantId = userStore.userInfo?.tenantId || 0
  form.code = ''
  form.name = ''
  form.description = ''
  form.status = 1
}

const handleAdd = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: WorkflowRole) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    tenantId: row.tenantId,
    code: row.code,
    name: row.name,
    description: row.description || '',
    status: row.status
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (isEdit.value) {
    await updateWorkflowRole(form.id, form)
  } else {
    await createWorkflowRole({ ...form, tenantId: isSuperAdmin.value ? form.tenantId : undefined })
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetchData()
}

const handleDelete = async (row: WorkflowRole) => {
  await ElMessageBox.confirm(`确认删除流程角色「${row.name}」？`, '提示', { type: 'warning' })
  await deleteWorkflowRole(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

// 可选成员：租户用户由后端强制本租户；超管传角色所属租户，保证候选人与角色同租户
const loadUserOptions = async () => {
  memberLoading.value = true
  try {
    const res: any = await getWorkflowRoleUserOptions({
      keyword: memberKeyword.value,
      tenantId: isSuperAdmin.value ? currentRole.value.tenantId : undefined,
      limit: 500
    })
    userOptions.value = res.data || []
  } finally {
    memberLoading.value = false
  }
}

const handleMembers = async (row: WorkflowRole) => {
  currentRole.value = row
  memberKeyword.value = ''
  memberDialogVisible.value = true
  await Promise.all([loadUserOptions(), loadMembers(row.id)])
}

const loadMembers = async (id: number) => {
  const res: any = await getWorkflowRole(id)
  currentRole.value = res.data
  selectedUserIds.value = (res.data?.users || []).map((u: any) => u.id)
}

const handleSaveMembers = async () => {
  await assignWorkflowRoleUsers(currentRole.value.id, selectedUserIds.value)
  ElMessage.success('成员保存成功')
  memberDialogVisible.value = false
  fetchData()
}

onMounted(() => {
  fetchData()
  if (isSuperAdmin.value) {
    ensureTenants()
  }
})
</script>

<style scoped lang="scss">
.workflow-role-page {
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
  .member-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }
  .member-role {
    font-weight: 600;
    color: #1e293b;
  }
  .form-tip {
    width: 100%;
    font-size: 12px;
    color: #94a3b8;
    line-height: 1.6;
  }
}
</style>
