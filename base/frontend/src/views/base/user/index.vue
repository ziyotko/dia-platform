<template>
  <div class="user-page">
    <div class="page-header">
      <div class="page-title">
        <div class="page-icon">
          <el-icon :size="24" color="#fff"><UserIcon /></el-icon>
        </div>
        <div class="page-title-text">
          <h1>用户管理</h1>
          <p>管理系统用户账号与权限分配</p>
        </div>
      </div>
      <el-button v-if="can('base:user:create')" type="primary" :icon="Plus" @click="handleAdd">新增用户</el-button>
    </div>

    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>用户列表</span>
        </div>
      </template>

      <el-form :inline="true" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="用户名 / 姓名 / 手机号" clearable style="width: 220px" @keyup.enter="handleSearch" />
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
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="realName" label="真实姓名" />
        <el-table-column prop="phone" label="手机号" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column v-if="isSuperAdmin" label="所属租户" min-width="140">
          <template #default="{ row }">{{ tenantName(row.tenantId) }}</template>
        </el-table-column>
        <el-table-column label="角色" min-width="160">
          <template #default="{ row }">
            <span v-if="row.roles?.length">{{ row.roles.map((r: any) => r.name).join('、') }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button v-if="can('base:user:update')" link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button v-if="can('base:user:assign-role')" link type="primary" @click="handleAssignRole(row)">分配角色</el-button>
            <el-button v-if="can('base:user:reset-password')" link type="primary" @click="handleResetPwd(row)">重置密码</el-button>
            <el-button v-if="can('base:user:delete')" link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '新增用户'" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item v-if="isSuperAdmin" label="所属租户" prop="tenantId">
          <tenant-select v-model="form.tenantId" :disabled="isEdit" :clearable="!isEdit" />
          <div class="form-tip">编辑时不允许变更用户所属租户</div>
        </el-form-item>
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="真实姓名">
          <el-input v-model="form.realName" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="密码" v-if="!isEdit" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            placeholder="登录初始密码"
          />
          <div class="form-tip">长度需满足「系统设置 → 安全策略」的密码最小长度（默认 8 位）</div>
        </el-form-item>
        <el-form-item label="管理员">
          <el-switch v-model="form.isAdmin" />
          <div class="form-tip">开启后在本租户内拥有全部接口权限与菜单</div>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="roleDialogVisible" title="分配角色" width="520px">
      <el-alert
        :title="`用户：${currentUser?.username || ''}（仅可选择同租户的角色）`"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 12px"
      />
      <el-select v-model="selectedRoleIds" multiple placeholder="请选择角色" style="width: 100%">
        <el-option v-for="r in roleOptions" :key="r.id" :label="`${r.name}（${r.code}）`" :value="r.id" />
      </el-select>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveRoles">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { User as UserIcon, Plus } from '@element-plus/icons-vue'
import {
  getUserList,
  createUser,
  updateUser,
  deleteUser,
  resetUserPassword,
  assignUserRoles
} from '@/api/user'
import type { User } from '@/api/user'
import { getRoleList } from '@/api/role'
import type { Role } from '@/api/role'
import { tenantName, ensureTenants } from '@/utils/tenantOptions'
import { useUserStore } from '@/stores/user'
import TenantSelect from '@/components/TenantSelect.vue'

const userStore = useUserStore()
// 仅平台超级管理员（tenantId === 0）可以跨租户管理
const isSuperAdmin = computed(() => userStore.userInfo?.tenantId === 0)
// 按钮级权限：与后端 base:user:* 权限点对齐
const can = (code: string) => userStore.can(code)

const loading = ref(false)
const tableData = ref<User[]>([])
const total = ref(0)
const query = reactive<{ page: number; size: number; keyword: string; tenantId?: number }>({
  page: 1,
  size: 10,
  keyword: '',
  tenantId: undefined
})
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<any>(null)
const form = reactive<any>({
  id: 0,
  tenantId: 0,
  username: '',
  realName: '',
  phone: '',
  email: '',
  password: '',
  isAdmin: false,
  status: 1
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  // 初始密码必填：后端不再隐式默认 123456（6 位无法通过安全策略的密码最小长度校验）
  password: [{ required: true, message: '请输入初始密码', trigger: 'blur' }],
  email: [{ type: 'email', message: '邮箱格式不正确', trigger: 'blur' }]
}

// 分配角色
const roleDialogVisible = ref(false)
const currentUser = ref<User | null>(null)
const roleOptions = ref<Role[]>([])
const selectedRoleIds = ref<number[]>([])

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getUserList(query)
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
  query.tenantId = undefined
  query.page = 1
  fetchData()
}

const handleAdd = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: User) => {
  isEdit.value = true
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row: User) => {
  await ElMessageBox.confirm('确认删除该用户？', '提示', { type: 'warning' })
  await deleteUser(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

// 重置密码：由管理员显式输入新密码（与新增用户同一套长度校验由后端执行）
const handleResetPwd = async (row: User) => {
  const { value } = await ElMessageBox.prompt(
    `请输入 ${row.username} 的新密码（需满足安全策略的密码最小长度）：`,
    '重置密码',
    {
      inputType: 'password',
      inputPlaceholder: '新密码',
      inputValidator: (val: string) => (val ? true : '请输入新密码')
    }
  )
  await resetUserPassword(row.id, value)
  ElMessage.success('密码已重置')
}

const handleAssignRole = async (row: User) => {
  currentUser.value = row
  selectedRoleIds.value = (row.roles || []).map((r) => r.id)
  const res: any = await getRoleList({ page: 1, size: 1000, tenantId: row.tenantId })
  roleOptions.value = res.data.list || []
  roleDialogVisible.value = true
}

const handleSaveRoles = async () => {
  if (!currentUser.value) return
  await assignUserRoles(currentUser.value.id, selectedRoleIds.value)
  ElMessage.success('角色分配成功')
  roleDialogVisible.value = false
  fetchData()
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (isEdit.value) {
    await updateUser(form.id, {
      realName: form.realName,
      phone: form.phone,
      email: form.email,
      isAdmin: form.isAdmin,
      status: form.status
    })
  } else {
    await createUser({
      username: form.username,
      password: form.password,
      realName: form.realName,
      phone: form.phone,
      email: form.email,
      isAdmin: form.isAdmin,
      status: form.status,
      tenantId: isSuperAdmin.value ? form.tenantId : undefined
    })
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetchData()
}

const resetForm = () => {
  form.id = 0
  form.tenantId = userStore.userInfo?.tenantId || 0
  form.username = ''
  form.realName = ''
  form.phone = ''
  form.email = ''
  form.password = ''
  form.isAdmin = false
  form.status = 1
}

onMounted(() => {
  fetchData()
  if (isSuperAdmin.value) {
    ensureTenants()
  }
})
</script>

<style scoped lang="scss">
.user-page {
  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;
  }

  .page-title {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .page-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    background: linear-gradient(135deg, #2563eb 0%, #4f46e5 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 8px 20px rgba(37, 99, 235, 0.3);
  }

  .page-title-text {
    h1 {
      margin: 0;
      font-size: 20px;
      font-weight: 700;
      color: #1e293b;
    }

    p {
      margin: 4px 0 0;
      font-size: 13px;
      color: #64748b;
    }
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-weight: 600;
    color: #1e293b;
  }

  .search-form {
    margin-bottom: 16px;
  }

  .form-tip {
    width: 100%;
    font-size: 12px;
    color: #94a3b8;
    line-height: 1.6;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
