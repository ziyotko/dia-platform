<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="机构名称">
          <el-input v-model="queryForm.name" placeholder="请输入机构名称" clearable />
        </el-form-item>
        <el-form-item label="机构类型">
          <el-select v-model="queryForm.orgType" placeholder="全部类型" clearable style="width: 140px">
            <el-option label="机构" :value="1" />
            <el-option label="分支机构" :value="2" />
            <el-option label="内设机构" :value="3" />
          </el-select>
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
          <span>机构列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增机构
          </el-button>
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
        <el-table-column prop="name" label="机构名称" min-width="200">
          <template #default="{ row }">
            <el-icon style="margin-right: 6px; color: #409eff;"><OfficeBuilding /></el-icon>
            <span>{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="code" label="机构编码" min-width="140" />
        <el-table-column prop="orgType" label="机构类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="orgTypeTagType(row.orgType)" size="small">
              {{ row.orgTypeText || orgTypeText(row.orgType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="orgLevel" label="层级" width="80" align="center" />
        <el-table-column prop="category" label="机构分类" min-width="120" />
        <el-table-column prop="region" label="所属区域" min-width="140" show-overflow-tooltip />
        <el-table-column prop="manager" label="负责人" min-width="120" />
        <el-table-column prop="userCount" label="人员数量" width="100" align="center" />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column label="操作" width="300" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleAddChild(row)">
              <el-icon><CirclePlus /></el-icon>子机构
            </el-button>
            <el-button link type="primary" @click="handleAssignUsers(row)">
              <el-icon><User /></el-icon>选人
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

    <!-- 新增/编辑机构弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="620px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="100px"
      >
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="上级机构">
              <el-tree-select
                v-model="form.parentId"
                :data="orgTreeSelectData"
                :props="{ label: 'name', value: 'id', children: 'children' }"
                placeholder="请选择上级机构"
                clearable
                check-strictly
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="机构类型" prop="orgType">
              <el-select v-model="form.orgType" placeholder="请选择机构类型" style="width: 100%">
                <el-option label="机构" :value="1" />
                <el-option label="分支机构" :value="2" />
                <el-option label="内设机构" :value="3" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="机构名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入机构名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="机构编码" prop="code">
              <el-input v-model="form.code" placeholder="请输入机构编码" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="机构层级" prop="orgLevel">
              <el-input-number v-model="form.orgLevel" :min="1" :max="10" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="机构分类" prop="category">
              <el-input v-model="form.category" placeholder="如：职能/业务/区域" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="所属区域">
              <el-input v-model="form.region" placeholder="请输入所属区域" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="负责人">
              <el-select
                v-model="form.managerCode"
                placeholder="请选择负责人"
                clearable
                filterable
                style="width: 100%"
              >
                <el-option
                  v-for="user in dialogUserOptions"
                  :key="user.id"
                  :label="user.username"
                  :value="user.account"
                >
                  <span style="display: flex; align-items: center; justify-content: space-between;">
                    <span>{{ user.username }} ({{ user.account }})</span>
                    <el-icon v-if="form.managerCode === user.account" color="#409eff"><Check /></el-icon>
                  </span>
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="省份">
              <el-input v-model="form.province" placeholder="请输入省份" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="城市">
              <el-input v-model="form.city" placeholder="请输入城市" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="详细地址">
          <el-input v-model="form.address" placeholder="请输入详细地址" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="显示排序" prop="sort">
              <el-input-number v-model="form.sort" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态" prop="status">
              <el-radio-group v-model="form.status">
                <el-radio :value="1">启用</el-radio>
                <el-radio :value="0">禁用</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 机构选人弹窗 -->
    <el-dialog
      v-model="userDialogVisible"
      title="机构选人"
      width="700px"
      destroy-on-close
    >
      <div class="user-select-header">
        <span>当前机构：{{ currentOrgName }}</span>
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
        <el-table-column prop="nickname" label="昵称" min-width="120" />
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
  Check
} from '@element-plus/icons-vue'
import {
  getOrgList,
  createOrg,
  updateOrg,
  deleteOrg,
  getOrgUsers,
  assignOrgUsers,
  type OrgItem,
  type OrgForm
} from '@/api/org'
import { getUserList } from '@/api/user'

interface UserItem {
  id: number
  username: string
  account: string
  nickname: string
  phone: string
}

const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const formRef = ref()
const tableData = ref<OrgItem[]>([])

const userDialogVisible = ref(false)
const userLoading = ref(false)
const userSubmitLoading = ref(false)
const userSearch = ref('')
const userTableRef = ref()
const currentOrgId = ref<number>(0)
const currentOrgName = ref('')
const selectedUserIds = ref<number[]>([])

const queryForm = reactive({
  name: '',
  orgType: undefined as number | undefined,
  status: undefined as number | undefined
})

const form = reactive<OrgForm>({
  id: undefined,
  parentId: undefined,
  name: '',
  code: '',
  orgType: 3,
  orgLevel: 1,
  category: '',
  region: '',
  province: '',
  city: '',
  address: '',
  manager: '',
  managerCode: '',
  sort: 0,
  status: 1,
  description: ''
})

const formRules = {
  name: [{ required: true, message: '请输入机构名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入机构编码', trigger: 'blur' }],
  orgType: [{ required: true, message: '请选择机构类型', trigger: 'change' }],
  orgLevel: [{ required: true, message: '请输入机构层级', trigger: 'blur' }],
  sort: [{ required: true, message: '请输入排序', trigger: 'blur' }]
}

const userOptions = ref<UserItem[]>([])
const dialogUserOptions = ref<UserItem[]>([])

const orgTypeOptions = [
  { value: 1, label: '集团', tagType: 'danger' },
  { value: 2, label: '公司', tagType: 'warning' },
  { value: 3, label: '部门', tagType: 'primary' },
  { value: 4, label: '团队', tagType: 'success' }
]

const orgTypeText = (value: number) => {
  return orgTypeOptions.find((item) => item.value === value)?.label || '其他'
}

const orgTypeTagType = (value: number) => {
  return orgTypeOptions.find((item) => item.value === value)?.tagType || ''
}

const filteredUserOptions = computed(() => {
  if (!userSearch.value) return userOptions.value
  const keyword = userSearch.value.toLowerCase()
  return userOptions.value.filter(
    (u) =>
      u.username.toLowerCase().includes(keyword) ||
      u.account.toLowerCase().includes(keyword) ||
      u.nickname.toLowerCase().includes(keyword)
  )
})

const orgTreeSelectData = computed(() => {
  return [{ id: 0, parentId: 0, name: '顶级机构', children: [], hasChildren: false }, ...tableData.value]
})

const filteredTableData = computed(() => {
  let data = tableData.value
  if (queryForm.name) {
    const keyword = queryForm.name.toLowerCase()
    data = filterTree(data, (item) => item.name.toLowerCase().includes(keyword))
  }
  if (queryForm.orgType !== undefined) {
    data = filterTree(data, (item) => item.orgType === queryForm.orgType)
  }
  if (queryForm.status !== undefined) {
    data = filterTree(data, (item) => item.status === queryForm.status)
  }
  return data
})

function filterTree(nodes: OrgItem[], predicate: (item: OrgItem) => boolean): OrgItem[] {
  const result: OrgItem[] = []
  for (const node of nodes) {
    const match = predicate(node)
    let children: OrgItem[] | undefined
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
  queryForm.orgType = undefined
  queryForm.status = undefined
  fetchData()
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getOrgList({
      name: queryForm.name || undefined,
      orgType: queryForm.orgType,
      status: queryForm.status
    })
    if (res && res.code === 0) {
      tableData.value = res.data.list || []
    } else {
      ElMessage.error(res?.message || '获取机构列表失败')
    }
  } catch (error) {
    console.error('获取机构列表失败', error)
  } finally {
    loading.value = false
  }
}

const handleAdd = async () => {
  dialogTitle.value = '新增机构'
  resetForm()
  await fetchDialogUsers()
  dialogVisible.value = true
}

const handleAddChild = async (row: OrgItem) => {
  dialogTitle.value = '新增子机构'
  resetForm()
  form.parentId = row.id
  form.orgLevel = (row.orgLevel || 1) + 1
  await fetchDialogUsers()
  dialogVisible.value = true
}

const handleEdit = async (row: OrgItem) => {
  dialogTitle.value = '编辑机构'
  Object.assign(form, {
    id: row.id,
    parentId: row.parentId === 0 ? undefined : row.parentId,
    name: row.name,
    code: row.code,
    orgType: row.orgType,
    orgLevel: row.orgLevel,
    category: row.category,
    region: row.region,
    province: row.province,
    city: row.city,
    address: row.address,
    manager: row.manager,
    managerCode: row.managerCode,
    sort: row.sort,
    status: row.status,
    description: row.description
  })
  await fetchDialogUsers()
  dialogVisible.value = true
}

const handleDelete = (row: OrgItem) => {
  if (row.children && row.children.length > 0) {
    ElMessage.warning('请先删除子机构')
    return
  }
  ElMessageBox.confirm(`确定要删除机构 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res: any = await deleteOrg(row.id)
      if (res && res.code === 0) {
        ElMessage.success('删除成功')
        fetchData()
      } else {
        ElMessage.error(res?.message || '删除失败')
      }
    } catch (error) {
      console.error('删除机构失败', error)
    }
  })
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    const payload: OrgForm = {
      parentId: form.parentId || 0,
      name: form.name,
      code: form.code,
      orgType: form.orgType,
      orgLevel: form.orgLevel,
      category: form.category,
      region: form.region,
      province: form.province,
      city: form.city,
      address: form.address,
      manager: form.manager,
      managerCode: form.managerCode,
      sort: form.sort,
      status: form.status,
      description: form.description
    }
    let res: any
    if (form.id) {
      res = await updateOrg(form.id, payload)
    } else {
      res = await createOrg(payload)
    }
    if (res && res.code === 0) {
      ElMessage.success(form.id ? '修改成功' : '新增成功')
      dialogVisible.value = false
      fetchData()
    } else {
      ElMessage.error(res?.message || (form.id ? '修改失败' : '新增失败'))
    }
  } catch (error) {
    console.error('提交失败', error)
  } finally {
    submitLoading.value = false
  }
}

const resetForm = () => {
  form.id = undefined
  form.parentId = undefined
  form.name = ''
  form.code = ''
  form.orgType = 3
  form.orgLevel = 1
  form.category = ''
  form.region = ''
  form.province = ''
  form.city = ''
  form.address = ''
  form.manager = ''
  form.managerCode = ''
  form.sort = 0
  form.status = 1
  form.description = ''
}

const syncManagerFromCode = () => {
  if (!form.managerCode) {
    form.manager = ''
    return
  }
  const user = dialogUserOptions.value.find((u) => u.account === form.managerCode)
  if (user) {
    form.manager = user.username
  } else {
    form.manager = ''
  }
}

watch(() => form.managerCode, syncManagerFromCode)
watch(() => dialogUserOptions.value, syncManagerFromCode)

const fetchDialogUsers = async () => {
  try {
    const res: any = await getUserList({ page: 1, pageSize: 1000, status: 1 })
    if (res && res.code === 0) {
      dialogUserOptions.value = (res.data.list || []).map((u: any) => ({
        id: u.id,
        username: u.username,
        account: u.account,
        nickname: u.nickname,
        phone: u.phone || u.mobile || ''
      }))
    }
  } catch (error) {
    console.error('获取用户列表失败', error)
  }
}

const fetchUsers = async () => {
  try {
    const res: any = await getUserList({ page: 1, pageSize: 1000 })
    if (res && res.code === 0) {
      userOptions.value = (res.data.list || []).map((u: any) => ({
        id: u.id,
        username: u.username,
        account: u.account,
        nickname: u.nickname,
        phone: u.phone || u.mobile || ''
      }))
    }
  } catch (error) {
    console.error('获取用户列表失败', error)
  }
}

const handleAssignUsers = async (row: OrgItem) => {
  currentOrgId.value = row.id
  currentOrgName.value = row.name
  userDialogVisible.value = true
  userLoading.value = true
  selectedUserIds.value = []
  try {
    await fetchUsers()
    const res: any = await getOrgUsers(row.id)
    if (res && res.code === 0) {
      selectedUserIds.value = res.data || []
    }
    nextTick(() => {
      const rows = userOptions.value.filter((u) => selectedUserIds.value.includes(u.id))
      rows.forEach((r) => {
        userTableRef.value?.toggleRowSelection(r, true)
      })
    })
  } catch (error) {
    console.error('获取机构用户失败', error)
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
    const res: any = await assignOrgUsers(currentOrgId.value, selectedUserIds.value)
    if (res && res.code === 0) {
      ElMessage.success('人员分配成功')
      userDialogVisible.value = false
      fetchData()
    } else {
      ElMessage.error(res?.message || '人员分配失败')
    }
  } catch (error) {
    console.error('人员分配失败', error)
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
}
</style>
