<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="部门名称">
          <el-input v-model="queryForm.name" placeholder="请输入部门名称" clearable />
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
          <span>部门列表</span>
          <div class="header-actions">
            <el-button type="success" @click="handleImport">
              <el-icon><Upload /></el-icon>导入部门
            </el-button>
            <el-button type="primary" @click="handleAdd">
              <el-icon><Plus /></el-icon>新增部门
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="filteredTableData"
        v-loading="loading"
        row-key="id"
        border
        stripe
        default-expand-all
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      >
        <el-table-column prop="name" label="部门名称" min-width="280">
          <template #default="{ row }">
            <el-icon style="margin-right: 6px; color: #002fa7;"><OfficeBuilding /></el-icon>
            <span>{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="code" label="部门编码" min-width="140" />
        <el-table-column prop="orgName" label="所属机构" min-width="280" />
        <el-table-column prop="leader" label="负责人" min-width="100" />
        <el-table-column prop="userCount" label="人员数量" width="100" align="center" />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column label="操作" width="420" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleAddChild(row)">
              <el-icon><CirclePlus /></el-icon>子部门
            </el-button>
            <el-button link type="primary" @click="handleAssignUsers(row)">
              <el-icon><User /></el-icon>选人
            </el-button>
            <el-button link type="primary" @click="handleViewUsers(row)">
              <el-icon><User /></el-icon>人员查看
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
    </el-card>

    <!-- 新增/编辑部门弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="520px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="90px"
      >
        <el-form-item label="所属机构" prop="orgId">
          <el-tree-select
            v-model="form.orgId"
            :data="orgTreeData"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="请选择所属机构"
            clearable
            check-strictly
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="上级部门">
          <el-tree-select
            v-model="form.parentId"
            :data="deptTreeSelectData"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="请选择上级部门"
            clearable
            check-strictly
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入部门名称" />
        </el-form-item>
        <el-form-item label="部门编码" prop="code">
          <el-input v-model="form.code" placeholder="请输入部门编码" />
        </el-form-item>
        <el-form-item label="负责人" prop="leader">
          <el-select
            v-model="form.leaderCode"
            :placeholder="form.orgId ? '请选择本机构负责人' : '请先选择所属机构'"
            clearable
            filterable
            :disabled="!form.orgId"
            style="width: 100%"
          >
            <el-option
              v-for="user in dialogUserOptions"
              :key="user.id"
              :label="user.username"
              :value="String(user.id)"
            >
              <span style="display: flex; align-items: center; justify-content: space-between;">
                <span>{{ user.username }} ({{ user.account }})</span>
                <el-icon v-if="form.leaderCode === String(user.id)" color="#002fa7"><Check /></el-icon>
              </span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="显示排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 导入部门弹窗 -->
    <el-dialog
      v-model="importDialogVisible"
      title="导入部门"
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
            仅支持 .xlsx / .xls 文件，第一行为标题：上级部门ID（没有就是0）、部门名称、部门编码、机构名称
          </div>
        </template>
      </el-upload>
      <template #footer>
        <el-button @click="importDialogVisible = false">关闭</el-button>
        <el-button type="primary" :loading="importLoading" @click="handleImportSubmit">确定导入</el-button>
      </template>
    </el-dialog>

    <!-- 部门人员查看弹窗 -->
    <el-dialog
      v-model="viewUserDialogVisible"
      title="部门人员"
      width="700px"
      destroy-on-close
    >
      <div class="user-select-header">
        <span>当前部门：{{ viewCurrentDeptName }}</span>
        <el-input
          v-model="viewUserSearch"
          placeholder="搜索用户姓名/账号"
          clearable
          style="width: 220px"
        />
      </div>
      <el-table
        :data="paginatedViewUserOptions"
        v-loading="viewUserLoading"
        border
        stripe
        height="360"
      >
        <el-table-column label="序号" width="60" align="center">
          <template #default="{ $index }">
            {{ (viewUserPage - 1) * viewUserPageSize + $index + 1 }}
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="account" label="账号" min-width="120" />
        <el-table-column prop="phone" label="手机号" min-width="130" />
      </el-table>
      <div class="user-pagination">
        <el-pagination
          v-model:current-page="viewUserPage"
          v-model:page-size="viewUserPageSize"
          :page-sizes="[6]"
          :total="viewUserTotal"
          layout="total, prev, pager, next"
          :pager-count="5"
          small
        />
      </div>
      <template #footer>
        <el-button @click="viewUserDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 部门选人弹窗 -->
    <el-dialog
      v-model="userDialogVisible"
      title="部门选人"
      width="700px"
      destroy-on-close
    >
      <div class="user-select-header">
        <span>当前部门：{{ currentDeptName }}</span>
        <el-input
          v-model="userSearch"
          placeholder="搜索用户姓名/账号"
          clearable
          style="width: 220px"
        />
      </div>
      <el-table
        :data="filteredUserOptions"
        v-loading="userLoading"
        border
        stripe
        height="360"
        @selection-change="handleUserSelectionChange"
        ref="userTableRef"
      >
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="account" label="账号" min-width="120" />
        <el-table-column prop="phone" label="手机号" min-width="130" />
      </el-table>
      <template #footer>
        <el-button @click="userDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="userSubmitLoading" @click="handleUserSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  RefreshRight,
  Plus,
  Edit,
  Delete,
  CirclePlus,
  User,
  OfficeBuilding,
  Check,
  Upload
} from '@element-plus/icons-vue'
import {
  getDepartmentList,
  createDepartment,
  updateDepartment,
  deleteDepartment,
  getDepartmentUsers,
  assignDepartmentUsers,
  importDepartments
} from '@/api/department'
import { getUserList } from '@/api/user'
import { getOrgTree, getOrgUsers, type OrgItem } from '@/api/org'

interface DeptItem {
  id: number
  parentId: number
  orgId: number
  orgName: string
  name: string
  code: string
  leader: string
  leaderCode: string
  sort: number
  status: number
  description: string
  userCount: number
  createTime: string
  children?: DeptItem[]
  hasChildren?: boolean
}

interface UserItem {
  id: number
  username: string
  account: string
  phone: string
}

const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const formRef = ref()
const tableData = ref<DeptItem[]>([])
const importDialogVisible = ref(false)
const importLoading = ref(false)
const importFile = ref<File | null>(null)

const userDialogVisible = ref(false)
const userLoading = ref(false)
const userSubmitLoading = ref(false)
const userSearch = ref('')
const userTableRef = ref()
const currentDeptId = ref<number>(0)
const currentDeptName = ref('')
const currentOrgId = ref<number>(0)
const selectedUserIds = ref<number[]>([])

const viewUserDialogVisible = ref(false)
const viewUserLoading = ref(false)
const viewUserSearch = ref('')
const viewUserPage = ref(1)
const viewUserPageSize = ref(6)
const viewCurrentDeptId = ref<number>(0)
const viewCurrentDeptName = ref('')
const viewUserOptions = ref<UserItem[]>([])

const queryForm = reactive({
  name: '',
  status: undefined as number | undefined
})

const form = reactive({
  id: undefined as number | undefined,
  parentId: undefined as number | undefined,
  orgId: undefined as number | undefined,
  name: '',
  code: '',
  leader: '',
  leaderCode: '',
  sort: 0,
  status: 1,
  description: ''
})

const formRules = {
  name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入部门编码', trigger: 'blur' }],
  orgId: [{ required: true, message: '请选择所属机构', trigger: 'change' }],
  sort: [{ required: true, message: '请输入排序', trigger: 'blur' }]
}

const userOptions = ref<UserItem[]>([])
const allDialogUsers = ref<UserItem[]>([])
const dialogUserOptions = ref<UserItem[]>([])
const orgTreeData = ref<OrgItem[]>([])

const filteredUserOptions = computed(() => {
  if (!userSearch.value) return userOptions.value
  const keyword = userSearch.value.toLowerCase()
  return userOptions.value.filter(
    (u) =>
      u.username.toLowerCase().includes(keyword) ||
      u.account.toLowerCase().includes(keyword)
  )
})

const filteredViewUserOptions = computed(() => {
  if (!viewUserSearch.value) return viewUserOptions.value
  const keyword = viewUserSearch.value.toLowerCase()
  return viewUserOptions.value.filter(
    (u) =>
      u.username.toLowerCase().includes(keyword) ||
      u.account.toLowerCase().includes(keyword)
  )
})

const viewUserTotal = computed(() => filteredViewUserOptions.value.length)

const paginatedViewUserOptions = computed(() => {
  const start = (viewUserPage.value - 1) * viewUserPageSize.value
  return filteredViewUserOptions.value.slice(start, start + viewUserPageSize.value)
})

const deptTreeSelectData = computed(() => {
  return [...tableData.value]
})

const filteredTableData = computed(() => {
  let data = tableData.value
  if (queryForm.name) {
    const keyword = queryForm.name.toLowerCase()
    data = filterTree(data, (item) => item.name.toLowerCase().includes(keyword))
  }
  if (queryForm.status !== undefined) {
    data = filterTree(data, (item) => item.status === queryForm.status)
  }
  return data
})

function filterTree(nodes: DeptItem[], predicate: (item: DeptItem) => boolean): DeptItem[] {
  const result: DeptItem[] = []
  for (const node of nodes) {
    const match = predicate(node)
    let children: DeptItem[] | undefined
    if (node.children && node.children.length > 0) {
      children = filterTree(node.children, predicate)
    }
    if (match || (children && children.length > 0)) {
      result.push({ ...node, children })
    }
  }
  return result
}

const handleSearch = () => {
  fetchData()
}

const resetQuery = () => {
  queryForm.name = ''
  queryForm.status = undefined
  fetchData()
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getDepartmentList({
      name: queryForm.name || undefined,
      status: queryForm.status
    })
    if (res && res.code === 0) {
      tableData.value = res.data.list || []
    } else {
      ElMessage.error(res?.message || '获取部门列表失败')
    }
  } catch (error) {
    ElMessage.error('获取部门列表失败')
  } finally {
    loading.value = false
  }
}

const fetchOrgTreeData = async () => {
  try {
    const res: any = await getOrgTree()
    if (res && res.code === 0) {
      orgTreeData.value = res.data || []
    }
  } catch (error) {
    ElMessage.error('获取机构树失败')
  }
}

const handleAdd = async () => {
  dialogTitle.value = '新增部门'
  resetForm()
  await Promise.all([fetchOrgTreeData(), fetchDialogUsers()])
  dialogVisible.value = true
}

const handleAddChild = async (row: DeptItem) => {
  dialogTitle.value = '新增子部门'
  resetForm()
  form.parentId = row.id
  form.orgId = row.orgId
  await Promise.all([fetchOrgTreeData(), fetchDialogUsers()])
  await refreshDialogUsers()
  dialogVisible.value = true
}

const handleEdit = async (row: DeptItem) => {
  dialogTitle.value = '编辑部门'
  Object.assign(form, {
    id: row.id,
    parentId: row.parentId === 0 ? undefined : row.parentId,
    orgId: row.orgId,
    name: row.name,
    code: row.code,
    leader: row.leader,
    leaderCode: row.leaderCode,
    sort: row.sort,
    status: row.status,
    description: row.description
  })
  await Promise.all([fetchOrgTreeData(), fetchDialogUsers()])
  await refreshDialogUsers()
  dialogVisible.value = true
}

const handleDelete = (row: DeptItem) => {
  if (row.children && row.children.length > 0) {
    ElMessage.warning('请先删除子部门')
    return
  }
  ElMessageBox.confirm(`确定要删除部门 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res: any = await deleteDepartment(row.id)
      if (res && res.code === 0) {
        ElMessage.success('删除成功')
        fetchData()
      } else {
        ElMessage.error(res?.message || '删除失败')
      }
    } catch (error) {
      ElMessage.error('删除部门失败')
    }
  })
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    const payload = {
      parentId: form.parentId || 0,
      orgId: form.orgId,
      name: form.name,
      code: form.code,
      leader: form.leader,
      leaderCode: form.leaderCode,
      sort: form.sort,
      status: form.status,
      description: form.description
    }
    let res: any
    if (form.id) {
      res = await updateDepartment(form.id, payload)
    } else {
      res = await createDepartment(payload)
    }
    if (res && res.code === 0) {
      ElMessage.success(form.id ? '修改成功' : '新增成功')
      dialogVisible.value = false
      fetchData()
    } else {
      ElMessage.error(res?.message || (form.id ? '修改失败' : '新增失败'))
    }
  } catch (error) {
    ElMessage.error('提交失败')
  } finally {
    submitLoading.value = false
  }
}

const resetForm = () => {
  form.id = undefined
  form.parentId = undefined
  form.orgId = undefined
  form.name = ''
  form.code = ''
  form.leader = ''
  form.leaderCode = ''
  form.sort = 0
  form.status = 1
  form.description = ''
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
    const res: any = await importDepartments(importFile.value)
    if (res && res.code === 0) {
      const data = res.data || {}
      ElMessage.success(`导入完成，成功 ${data.successCount || 0} 条，失败 ${data.failCount || 0} 条`)
      if (data.failDetails && data.failDetails.length > 0) {
        console.warn('导入失败详情:', data.failDetails)
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

const syncLeaderFromCode = () => {
  if (!form.leaderCode) {
    form.leader = ''
    return
  }
  const user = dialogUserOptions.value.find((u) => String(u.id) === form.leaderCode)
  if (user) {
    form.leader = user.username
  } else {
    form.leader = ''
  }
}

watch(() => form.leaderCode, syncLeaderFromCode)
watch(() => dialogUserOptions.value, syncLeaderFromCode)
watch(() => form.orgId, () => {
  refreshDialogUsers()
})
watch(viewUserSearch, () => {
  viewUserPage.value = 1
})

const handleViewUsers = async (row: DeptItem) => {
  viewCurrentDeptId.value = row.id
  viewCurrentDeptName.value = row.name
  viewUserDialogVisible.value = true
  viewUserLoading.value = true
  viewUserSearch.value = ''
  viewUserPage.value = 1
  viewUserOptions.value = []
  try {
    const [usersRes, deptUsersRes]: any[] = await Promise.all([
      getUserList({ page: 1, pageSize: 1000 }),
      getDepartmentUsers(row.id)
    ])
    const allUsers = usersRes?.data?.list || []
    const deptUserIds = new Set((deptUsersRes?.data || []).map((id: any) => Number(id)))
    const userMap = new Map(allUsers.map((u: any) => [Number(u.id), u]))
    viewUserOptions.value = Array.from(deptUserIds)
      .map((id) => userMap.get(id))
      .filter((u): u is any => !!u)
      .map((u: any) => ({
        id: u.id,
        username: u.username,
        account: u.account,
        phone: u.phone || u.mobile || ''
      }))
  } catch (error) {
    ElMessage.error('获取部门人员失败')
  } finally {
    viewUserLoading.value = false
  }
}

const fetchDialogUsers = async () => {
  try {
    const res: any = await getUserList({ page: 1, pageSize: 1000, status: 1 })
    if (res && res.code === 0) {
      allDialogUsers.value = (res.data.list || []).map((u: any) => ({
        id: u.id,
        username: u.username,
        account: u.account,
        phone: u.phone || u.mobile || ''
      }))
    }
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  }
}

// 部门负责人必须是该机构的人：按所选机构过滤负责人候选
const refreshDialogUsers = async () => {
  let orgUserIds = new Set<number>()
  if (form.orgId) {
    try {
      const res: any = await getOrgUsers(form.orgId)
      orgUserIds = new Set((res?.data || []).map((id: any) => Number(id)))
    } catch (error) {
      orgUserIds = new Set()
    }
  }
  dialogUserOptions.value = form.orgId
    ? allDialogUsers.value.filter((u) => orgUserIds.has(u.id))
    : []
  // 若当前负责人不属于所选机构，则清空
  if (form.leaderCode) {
    const user = dialogUserOptions.value.find((u) => String(u.id) === form.leaderCode)
    if (!user) {
      form.leaderCode = ''
      form.leader = ''
    }
  }
}

const handleAssignUsers = async (row: DeptItem) => {
  currentDeptId.value = row.id
  currentDeptName.value = row.name
  currentOrgId.value = row.orgId
  userDialogVisible.value = true
  userLoading.value = true
  selectedUserIds.value = []
  try {
    const [usersRes, orgUsersRes, deptUsersRes]: any[] = await Promise.all([
      getUserList({ page: 1, pageSize: 1000 }),
      getOrgUsers(row.orgId),
      getDepartmentUsers(row.id)
    ])
    const allUsers = usersRes?.data?.list || []
    const orgUserIds = new Set((orgUsersRes?.data || []).map((id: any) => Number(id)))
    userOptions.value = allUsers
      .filter((u: any) => orgUserIds.has(Number(u.id)))
      .map((u: any) => ({
        id: u.id,
        username: u.username,
        account: u.account,
        phone: u.phone || u.mobile || ''
      }))
    selectedUserIds.value = deptUsersRes?.data || []
    nextTick(() => {
      const rows = userOptions.value.filter((u) => selectedUserIds.value.includes(u.id))
      rows.forEach((r) => {
        userTableRef.value?.toggleRowSelection(r, true)
      })
    })
  } catch (error) {
    ElMessage.error('获取部门用户失败')
  } finally {
    userLoading.value = false
  }
}

const handleUserSelectionChange = (selection: UserItem[]) => {
  selectedUserIds.value = selection.map((item) => item.id)
}

const handleUserSubmit = async () => {
  userSubmitLoading.value = true
  try {
    const res: any = await assignDepartmentUsers(currentDeptId.value, selectedUserIds.value)
    if (res && res.code === 0) {
      ElMessage.success('人员分配成功')
      userDialogVisible.value = false
      fetchData()
    } else {
      ElMessage.error(res?.message || '人员分配失败')
    }
  } catch (error) {
    ElMessage.error('人员分配失败')
  } finally {
    userSubmitLoading.value = false
  }
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

      .header-actions {
        display: flex;
        gap: 10px;
      }
    }
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  .user-select-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    font-weight: 600;
    color: #2c3e50;
  }

  .user-pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
