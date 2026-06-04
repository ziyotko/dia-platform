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
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增部门
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
        <el-table-column prop="name" label="部门名称" min-width="180">
          <template #default="{ row }">
            <el-icon style="margin-right: 6px; color: #409eff;"><OfficeBuilding /></el-icon>
            <span>{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="code" label="部门编码" min-width="140" />
        <el-table-column prop="leader" label="负责人" min-width="120" />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column label="操作" width="260" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleAddChild(row)">
              <el-icon><CirclePlus /></el-icon>子部门
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
          <el-input v-model="form.leader" placeholder="请输入负责人姓名" />
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
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  RefreshRight,
  Plus,
  Edit,
  Delete,
  CirclePlus,
  User,
  OfficeBuilding
} from '@element-plus/icons-vue'

interface DeptItem {
  id: number
  parentId: number
  name: string
  code: string
  leader: string
  sort: number
  status: number
  description: string
  createTime: string
  children?: DeptItem[]
  hasChildren?: boolean
}

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
const tableData = ref<DeptItem[]>([])

const userDialogVisible = ref(false)
const userLoading = ref(false)
const userSubmitLoading = ref(false)
const userSearch = ref('')
const userTableRef = ref()
const currentDeptId = ref<number>(0)
const currentDeptName = ref('')
const selectedUserIds = ref<number[]>([])
const deptUserMap = ref<Record<number, number[]>>({})

const queryForm = reactive({
  name: '',
  status: undefined as number | undefined
})

const form = reactive({
  id: undefined as number | undefined,
  parentId: undefined as number | undefined,
  name: '',
  code: '',
  leader: '',
  sort: 0,
  status: 1,
  description: ''
})

const formRules = {
  name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入部门编码', trigger: 'blur' }],
  leader: [{ required: true, message: '请输入负责人', trigger: 'blur' }],
  sort: [{ required: true, message: '请输入排序', trigger: 'blur' }]
}

// 模拟用户数据
const userOptions = ref<UserItem[]>([
  { id: 1, username: '张三', account: 'zhangsan', nickname: '张三', phone: '13800138001' },
  { id: 2, username: '李四', account: 'lisi', nickname: '李四', phone: '13800138002' },
  { id: 3, username: '王五', account: 'wangwu', nickname: '王五', phone: '13800138003' },
  { id: 4, username: '赵六', account: 'zhaoliu', nickname: '赵六', phone: '13800138004' },
  { id: 5, username: '孙七', account: 'sunqi', nickname: '孙七', phone: '13800138005' },
  { id: 6, username: '周八', account: 'zhouba', nickname: '周八', phone: '13800138006' },
  { id: 7, username: '吴九', account: 'wujiu', nickname: '吴九', phone: '13800138007' },
  { id: 8, username: '郑十', account: 'zhengshi', nickname: '郑十', phone: '13800138008' }
])

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

// 生成模拟部门数据
const generateMockData = (): DeptItem[] => {
  return [
    {
      id: 1,
      parentId: 0,
      name: '总公司',
      code: 'HQ',
      leader: '张三',
      sort: 1,
      status: 1,
      description: '公司总部',
      createTime: '2024-01-15 09:30:00',
      children: [
        {
          id: 2,
          parentId: 1,
          name: '技术研发部',
          code: 'RD',
          leader: '李四',
          sort: 1,
          status: 1,
          description: '负责产品研发',
          createTime: '2024-01-16 10:00:00',
          children: [
            {
              id: 5,
              parentId: 2,
              name: '前端组',
              code: 'RD-FE',
              leader: '王五',
              sort: 1,
              status: 1,
              description: '前端开发',
              createTime: '2024-02-01 14:00:00'
            },
            {
              id: 6,
              parentId: 2,
              name: '后端组',
              code: 'RD-BE',
              leader: '赵六',
              sort: 2,
              status: 1,
              description: '后端开发',
              createTime: '2024-02-01 14:30:00'
            }
          ]
        },
        {
          id: 3,
          parentId: 1,
          name: '产品部',
          code: 'PD',
          leader: '孙七',
          sort: 2,
          status: 1,
          description: '产品设计与规划',
          createTime: '2024-01-17 11:00:00'
        },
        {
          id: 4,
          parentId: 1,
          name: '市场部',
          code: 'MK',
          leader: '周八',
          sort: 3,
          status: 0,
          description: '市场推广与运营',
          createTime: '2024-01-18 09:00:00',
          children: [
            {
              id: 7,
              parentId: 4,
              name: '品牌组',
              code: 'MK-BR',
              leader: '吴九',
              sort: 1,
              status: 1,
              description: '品牌建设',
              createTime: '2024-03-01 10:00:00'
            }
          ]
        }
      ]
    }
  ]
}

const deptTreeSelectData = computed(() => {
  return [{ id: 0, parentId: 0, name: '顶级部门', children: [], hasChildren: false }, ...tableData.value]
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

function flattenTree(nodes: DeptItem[]): DeptItem[] {
  const result: DeptItem[] = []
  const traverse = (list: DeptItem[]) => {
    for (const node of list) {
      result.push(node)
      if (node.children && node.children.length > 0) {
        traverse(node.children)
      }
    }
  }
  traverse(nodes)
  return result
}

const handleSearch = () => {
  // 前端过滤，无需重置页码
}

const resetQuery = () => {
  queryForm.name = ''
  queryForm.status = undefined
}

const fetchData = async () => {
  loading.value = true
  try {
    // 模拟接口延迟
    await new Promise((resolve) => setTimeout(resolve, 300))
    tableData.value = generateMockData()
  } catch (error) {
    console.error('获取部门列表失败', error)
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  dialogTitle.value = '新增部门'
  resetForm()
  dialogVisible.value = true
}

const handleAddChild = (row: DeptItem) => {
  dialogTitle.value = '新增子部门'
  resetForm()
  form.parentId = row.id
  dialogVisible.value = true
}

const handleEdit = (row: DeptItem) => {
  dialogTitle.value = '编辑部门'
  Object.assign(form, {
    id: row.id,
    parentId: row.parentId === 0 ? undefined : row.parentId,
    name: row.name,
    code: row.code,
    leader: row.leader,
    sort: row.sort,
    status: row.status,
    description: row.description
  })
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
  }).then(() => {
    removeDeptById(tableData.value, row.id)
    ElMessage.success('删除成功')
  })
}

function removeDeptById(nodes: DeptItem[], id: number): boolean {
  for (let i = 0; i < nodes.length; i++) {
    if (nodes[i].id === id) {
      nodes.splice(i, 1)
      return true
    }
    if (nodes[i].children && nodes[i].children!.length > 0) {
      const removed = removeDeptById(nodes[i].children!, id)
      if (removed) return true
    }
  }
  return false
}

function findDeptById(nodes: DeptItem[], id: number): DeptItem | undefined {
  for (const node of nodes) {
    if (node.id === id) return node
    if (node.children && node.children.length > 0) {
      const found = findDeptById(node.children, id)
      if (found) return found
    }
  }
  return undefined
}

function addOrUpdateDept(nodes: DeptItem[], item: DeptItem, parentId?: number): boolean {
  if (item.id) {
    // 更新
    for (const node of nodes) {
      if (node.id === item.id) {
        Object.assign(node, item)
        return true
      }
      if (node.children && node.children.length > 0) {
        const updated = addOrUpdateDept(node.children, item, parentId)
        if (updated) return true
      }
    }
  } else {
    // 新增
    const targetParentId = parentId || 0
    if (targetParentId === 0) {
      nodes.push(item)
      return true
    }
    for (const node of nodes) {
      if (node.id === targetParentId) {
        if (!node.children) node.children = []
        node.children.push(item)
        node.hasChildren = true
        return true
      }
      if (node.children && node.children.length > 0) {
        const added = addOrUpdateDept(node.children, item, parentId)
        if (added) return true
      }
    }
  }
  return false
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    await new Promise((resolve) => setTimeout(resolve, 300))
    const payload: DeptItem = {
      id: form.id || Date.now(),
      parentId: form.parentId || 0,
      name: form.name,
      code: form.code,
      leader: form.leader,
      sort: form.sort,
      status: form.status,
      description: form.description,
      createTime: form.id ? '' : new Date().toLocaleString().replace(/\//g, '-'),
      children: []
    }
    if (form.id) {
      addOrUpdateDept(tableData.value, payload)
      ElMessage.success('修改成功')
    } else {
      addOrUpdateDept(tableData.value, payload, payload.parentId)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
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
  form.leader = ''
  form.sort = 0
  form.status = 1
  form.description = ''
}

// 部门选人
const handleAssignUsers = async (row: DeptItem) => {
  currentDeptId.value = row.id
  currentDeptName.value = row.name
  userDialogVisible.value = true
  userLoading.value = true
  selectedUserIds.value = deptUserMap.value[row.id] || []
  try {
    await new Promise((resolve) => setTimeout(resolve, 200))
    nextTick(() => {
      const rows = userOptions.value.filter((u) => selectedUserIds.value.includes(u.id))
      rows.forEach((r) => {
        userTableRef.value?.toggleRowSelection(r, true)
      })
    })
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
    await new Promise((resolve) => setTimeout(resolve, 300))
    deptUserMap.value[currentDeptId.value] = [...selectedUserIds.value]
    ElMessage.success('人员分配成功')
    userDialogVisible.value = false
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
