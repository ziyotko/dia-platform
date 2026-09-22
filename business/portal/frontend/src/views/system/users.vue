<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="用户名">
          <el-input v-model="queryForm.username" placeholder="请输入用户名" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="用户账号">
          <el-input v-model="queryForm.account" placeholder="请输入用户账号" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryForm.status" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
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
          <span>用户列表</span>
          <div class="header-actions">
            <el-button type="success" @click="handleImport">
              <el-icon><Upload /></el-icon>导入用户
            </el-button>
            <el-button type="primary" @click="handleAdd">
              <el-icon><Plus /></el-icon>新增用户
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="account" label="用户账号" min-width="120" />
        <el-table-column label="所属机构" min-width="200">
          <template #default="{ row }">
            <span>{{ row.orgNames?.join('、') || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column prop="phone" label="手机号" min-width="130" />
        <el-table-column label="性别" width="80" align="center">
          <template #default="{ row }">
            <span>{{ row.sex === 1 ? '男' : row.sex === 2 ? '女' : '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              :disabled="row.id === 1"
              @change="(val: number) => handleStatusChange(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" />
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button v-if="row.id !== 1" link type="danger" @click="handleDelete(row)">
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
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" :disabled="isReadonly" />
        </el-form-item>
        <el-form-item label="用户账号" prop="account">
          <el-input v-model="form.account" placeholder="请输入用户账号" :disabled="isReadonly || isEditing" />
        </el-form-item>
        <el-form-item label="所属机构" prop="orgIds">
          <el-tree-select
            v-model="form.orgIds"
            :data="orgOptions"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="请选择所属机构"
            multiple
            clearable
            check-strictly
            style="width: 100%"
            :disabled="isReadonly"
          />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" :disabled="isReadonly" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入手机号" :disabled="isReadonly" />
        </el-form-item>
        <el-form-item label="性别" prop="sex">
          <el-select v-model="form.sex" placeholder="请选择性别" clearable style="width: 100%" :disabled="isReadonly">
            <el-option label="男" :value="1" />
            <el-option label="女" :value="2" />
            <el-option label="未知" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="角色" prop="roleIds">
          <el-select v-model="form.roleIds" multiple placeholder="请选择角色" style="width: 100%" :disabled="isReadonly">
            <el-option
              v-for="role in assignableRoles"
              :key="role.id"
              :label="role.name"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status" :disabled="isReadonly">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="isEditing && !isReadonly" label="重置密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            clearable
            :placeholder="`留空则不修改，至少 ${minPasswordLength} 位`"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">关闭</el-button>
        <el-button v-if="!isReadonly" type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="importDialogVisible"
      title="导入用户"
      width="500px"
      destroy-on-close
    >
      <el-upload
        ref="uploadRef"
        drag
        action=""
        :auto-upload="false"
        :limit="1"
        accept=".xlsx,.xls"
        :on-change="handleImportFileChange"
        :on-remove="handleImportFileRemove"
      >
        <el-icon class="el-icon--upload"><Upload /></el-icon>
        <div class="el-upload__text">
          拖拽文件到此处或 <em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            仅支持 .xlsx / .xls 文件，第一行为标题：姓名、登陆账号、邮箱、手机号
          </div>
        </template>
      </el-upload>
      <template #footer>
        <el-button @click="importDialogVisible = false">关闭</el-button>
        <el-button type="primary" :loading="importLoading" @click="handleImportSubmit">确定导入</el-button>
      </template>
    </el-dialog>

    <!-- 初始密码 / 导入结果 -->
    <el-dialog v-model="credentialDialogVisible" :title="credentialTitle" width="560px" destroy-on-close>
      <template v-if="credentialList.length > 0">
        <el-alert
          type="warning"
          :closable="false"
          show-icon
          title="初始密码仅显示一次，请及时告知用户并妥善保存"
          style="margin-bottom: 12px"
        />
        <el-table :data="credentialList" border stripe max-height="300">
          <el-table-column prop="account" label="账号" min-width="140" />
          <el-table-column prop="password" label="初始密码" min-width="160" />
        </el-table>
      </template>
      <template v-if="importFailDetails.length > 0">
        <div style="margin-top: 12px; font-weight: 600; color: var(--app-text-heading)">
          失败详情（{{ importFailDetails.length }} 条）
        </div>
        <div
          style="max-height: 240px; overflow: auto; margin-top: 8px; color: #909399; font-size: 13px; line-height: 1.8"
        >
          <div v-for="(item, idx) in importFailDetails" :key="idx">{{ item }}</div>
        </div>
      </template>
      <template #footer>
        <el-button v-if="credentialList.length > 0" @click="copyCredentials">复制账号密码</el-button>
        <el-button type="primary" @click="credentialDialogVisible = false">知道了</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  RefreshRight,
  Plus,
  Edit,
  Delete,
  Upload
} from '@element-plus/icons-vue'
import { getUserList, createUser, updateUser, deleteUser, updateUserStatus, importUsers, checkUserUnique } from '@/api/user'
import { getAllRoles } from '@/api/role'
import { getOrgList, type OrgItem } from '@/api/org'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

// 密码最小长度来自系统设置（与后端 minPasswordLength 校验保持一致）
const minPasswordLength = computed(() => appStore.minPasswordLength || 8)

const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const isReadonly = ref(false)
const isEditing = ref(false)
const total = ref(0)
const formRef = ref()
const importDialogVisible = ref(false)
const importLoading = ref(false)
const importFile = ref<File | null>(null)

// 新建/导入用户的一次性密码展示
const credentialDialogVisible = ref(false)
const credentialTitle = ref('初始密码')
const credentialList = ref<{ account: string; password: string }[]>([])
const importFailDetails = ref<string[]>([])

const showCredentials = (
  list: { account: string; password: string }[],
  failDetails: string[] = [],
  title = '初始密码'
) => {
  credentialList.value = list
  importFailDetails.value = failDetails
  credentialTitle.value = title
  credentialDialogVisible.value = true
}

const copyCredentials = async () => {
  const text = credentialList.value.map((i) => `${i.account}\t${i.password}`).join('\n')
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.warning('复制失败，请手动选择复制')
  }
}

const queryForm = reactive({
  page: 1,
  pageSize: 10,
  username: '',
  account: '',
  status: undefined as number | undefined
})

const form = reactive({
  id: undefined as number | undefined,
  username: '',
  account: '',
  orgIds: [] as number[],
  email: '',
  phone: '',
  status: 1,
  roleIds: [] as number[],
  sex: undefined as number | undefined,
  // 重置密码：仅编辑时可用，留空表示不修改
  password: ''
})

const checkUnique = (field: string, label: string) => {
  return async (_rule: any, value: any, callback: any) => {
    if (!value) {
      callback()
      return
    }
    try {
      const res: any = await checkUserUnique(field, value, form.id)
      if (res.code === 0 && !res.data?.unique) {
        callback(new Error(`${label}已存在，请重新输入`))
      } else {
        callback()
      }
    } catch {
      callback()
    }
  }
}

const formRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  account: [
    { required: true, message: '请输入用户账号', trigger: 'blur' },
    { validator: checkUnique('account', '用户账号'), trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
    { validator: checkUnique('email', '邮箱'), trigger: 'blur' }
  ],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' },
    { validator: checkUnique('phone', '手机号'), trigger: 'blur' }
  ],
  roleIds: [{ required: true, message: '请选择角色', trigger: 'change', type: 'array' }],
  password: [
    {
      // 留空 = 不修改密码；填写时校验最小长度（与后端 minPasswordLength 设置一致）
      validator: (_rule: any, value: any, callback: any) => {
        if (!value) {
          callback()
          return
        }
        if (String(value).length < minPasswordLength.value) {
          callback(new Error(`密码长度不能少于 ${minPasswordLength.value} 位`))
          return
        }
        callback()
      },
      trigger: 'blur'
    }
  ]
}

const tableData = ref<any[]>([])
const roleOptions = ref<any[]>([])
const orgOptions = ref<OrgItem[]>([])

const userStore = useUserStore()

// 当前登录用户可分配的最高角色序号（角色序号越小角色越高，取当前用户角色中序号最小的）
const currentMinRoleId = computed(() => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const raw: any = userStore.userInfo?.roleIds
  const roleIds = Array.isArray(raw)
    ? raw.map(Number).filter((n) => !Number.isNaN(n))
    : typeof raw === 'string' && raw.trim()
      ? raw.split(',').map((s) => Number(s.trim())).filter((n) => !Number.isNaN(n))
      : []
  if (roleIds.length === 0) return 0
  return Math.min(...roleIds)
})

// 仅显示不高于当前登录用户最高角色的角色（序号 >= 当前用户最高角色序号）
const assignableRoles = computed(() => {
  const min = currentMinRoleId.value
  return roleOptions.value.filter((role: any) => Number(role.id) >= min)
})

const fetchOrgs = async () => {
  try {
    const res: any = await getOrgList()
    if (res && res.code === 0) {
      orgOptions.value = res.data.list || []
    }
  } catch (error) {
    ElMessage.error('获取机构列表失败')
  }
}

const fetchRoles = async () => {
  try {
    const res: any = await getAllRoles()
    if (res && res.code === 0) {
      roleOptions.value = res.data || []
    }
  } catch (error) {
    ElMessage.error('获取角色列表失败')
  }
}

const handleSearch = () => {
  queryForm.page = 1
  fetchData()
}

const resetQuery = () => {
  queryForm.username = ''
  queryForm.account = ''
  queryForm.status = undefined
  queryForm.page = 1
  fetchData()
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getUserList({
      page: queryForm.page,
      pageSize: queryForm.pageSize,
      username: queryForm.username || undefined,
      account: queryForm.account || undefined,
      status: queryForm.status
    })
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  isReadonly.value = false
  isEditing.value = false
  dialogTitle.value = '新增用户'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  isReadonly.value = row.id === 1
  isEditing.value = true
  dialogTitle.value = isReadonly.value ? '查看用户' : '编辑用户'
  Object.assign(form, {
    id: row.id,
    username: row.username,
    account: row.account,
    orgIds: row.orgIds || [],
    email: row.email,
    phone: row.phone,
    status: row.status,
    roleIds: row.roleIds || [],
    sex: row.sex,
    password: ''
  })
  dialogVisible.value = true
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除用户 "${row.username}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deleteUser(row.id)
      ElMessage.success('删除成功')
      fetchData()
    } catch (error) {
      ElMessage.error('删除用户失败')
    }
  })
}

const handleStatusChange = async (row: any, val: number) => {
  try {
    await updateUserStatus(row.id, val)
    ElMessage.success(`用户状态已${val === 1 ? '启用' : '禁用'}`)
  } catch (error) {
    row.status = val === 1 ? 0 : 1
    ElMessage.error('更新状态失败')
  }
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (form.roleIds.some((id) => id < currentMinRoleId.value)) {
    ElMessage.warning('不能设置高于当前登录用户最高角色的角色')
    return
  }
  submitLoading.value = true
  try {
    if (form.id) {
      await updateUser(form.id, {
        username: form.username,
        account: form.account,
        email: form.email,
        phone: form.phone,
        status: form.status,
        roleIds: form.roleIds,
        orgIds: form.orgIds,
        sex: form.sex,
        // 留空表示不修改密码（后端仅在非空时写入）
        password: form.password || undefined
      })
      ElMessage.success('修改成功')
    } else {
      const res: any = await createUser({
        username: form.username,
        account: form.account,
        email: form.email,
        phone: form.phone,
        status: form.status,
        roleIds: form.roleIds,
        orgIds: form.orgIds,
        sex: form.sex
      })
      ElMessage.success('新增成功')
      // 后端随机生成初始密码，仅在创建响应中返回一次，需展示给管理员
      const pwd = res?.data?.password
      if (pwd) {
        showCredentials([{ account: form.account, password: pwd }], [], '新增用户成功')
      }
    }
    dialogVisible.value = false
    fetchData()
  } catch (error) {
    ElMessage.error('提交失败')
  } finally {
    submitLoading.value = false
  }
}

const resetForm = () => {
  form.id = undefined
  form.username = ''
  form.account = ''
  form.orgIds = []
  form.email = ''
  form.phone = ''
  form.status = 1
  form.roleIds = []
  form.sex = undefined
  form.password = ''
}

const handleImport = () => {
  importFile.value = null
  importDialogVisible.value = true
}

const handleImportFileChange = (uploadFile: any) => {
  importFile.value = uploadFile.raw
}

const handleImportFileRemove = () => {
  importFile.value = null
}

const handleImportSubmit = async () => {
  if (!importFile.value) {
    ElMessage.warning('请选择要导入的 Excel 文件')
    return
  }
  importLoading.value = true
  try {
    const res: any = await importUsers(importFile.value)
    if (res && res.code === 0) {
      const data = res.data || {}
      ElMessage.success(`导入完成，成功 ${data.successCount || 0} 条，失败 ${data.failCount || 0} 条`)
      // 导入时后端为每个成功账号生成随机初始密码（仅本次返回），与失败明细一并展示
      const generated: Record<string, string> = data.generatedPasswords || {}
      const credentials = Object.entries(generated).map(([account, password]) => ({
        account,
        password: String(password)
      }))
      const failDetails: string[] = data.failDetails || []
      if (credentials.length > 0 || failDetails.length > 0) {
        showCredentials(credentials, failDetails, '导入结果')
      }
      importDialogVisible.value = false
      fetchData()
    }
  } catch (error) {
    ElMessage.error('导入失败')
  } finally {
    importLoading.value = false
  }
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
  fetchRoles()
  fetchOrgs()
})
</script>

<style scoped lang="scss">
.page-container {
  /* 列表页骨架（.search-card / .table-card / .card-header / .pagination）已统一到
     styles/global.scss 的 .page-container 规则（2026-09-22），此处仅保留页面特有部分 */
  .table-card {
    .card-header {
      .header-actions {
        display: flex;
        gap: 10px;
      }
    }
  }
}
</style>