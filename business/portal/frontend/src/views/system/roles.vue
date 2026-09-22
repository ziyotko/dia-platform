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
          <span>角色列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增角色
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
            <el-button link type="primary" @click="handlePermission(row)">
              <el-icon><Key /></el-icon>{{ row.id === 1 || Number(row.id) < currentMinRoleId ? '查看' : '权限' }}
            </el-button>
            <el-button v-if="!protectedRoleIds.includes(row.id)" link type="primary" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button v-if="!protectedRoleIds.includes(row.id)" link type="danger" @click="handleDelete(row)">
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
      v-model="permissionVisible"
      :title="isPermissionReadonly ? '查看权限' : '分配权限'"
      width="500px"
      destroy-on-close
    >
      <el-tree
        ref="treeRef"
        :data="permissionData"
        show-checkbox
        node-key="id"
        :props="{ label: 'name', children: 'children', disabled: () => isPermissionReadonly }"
        :default-expand-all="true"
      />
      <template #footer>
        <el-button @click="permissionVisible = false">关闭</el-button>
        <el-button v-if="!isPermissionReadonly" type="primary" :loading="permissionLoading" @click="handlePermissionSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { formatDateTime } from '@/utils/format'
import {
  Search,
  RefreshRight,
  Plus,
  Edit,
  Delete,
  Key
} from '@element-plus/icons-vue'
import {
  getRoleList,
  createRole,
  updateRole,
  deleteRole,
  getMenuTree,
  getRolePermissions,
  updateRolePermissions
} from '@/api/role'
import { useUserStore } from '@/stores/user'

// 系统内置默认角色 ID（管理员/内容审核/内容作者），禁止编辑、删除（权限分配按当前用户角色动态限制）
// 注：原 ID 2「普通管理员（admin）」已下线移除；历史库中残留的该角色按普通角色处理（可编辑/删除）
const protectedRoleIds = [1, 3, 4]

const userStore = useUserStore()

// 当前登录用户角色中序号最小的角色 ID（角色序号越小角色越高，取最小者作为自身"权限下限"）
const currentMinRoleId = computed(() => {
  const raw: any = userStore.userInfo?.roleIds
  const roleIds = Array.isArray(raw)
    ? raw.map(Number).filter((n) => !Number.isNaN(n))
    : typeof raw === 'string' && raw.trim()
      ? raw.split(',').map((s) => Number(s.trim())).filter((n) => !Number.isNaN(n))
      : []
  if (roleIds.length === 0) return 0
  return Math.min(...roleIds)
})

const loading = ref(false)
const dialogVisible = ref(false)
const permissionVisible = ref(false)
const permissionLoading = ref(false)
const isPermissionReadonly = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const total = ref(0)
const formRef = ref()
const treeRef = ref()
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
const permissionData = ref<any[]>([])

const getLeafNodeIds = (nodes: any[]): number[] => {
  const result: number[] = []
  const traverse = (data: any[]) => {
    data.forEach((node: any) => {
      if (!node.children || node.children.length === 0) {
        result.push(Number(node.id))
      } else {
        traverse(node.children)
      }
    })
  }
  traverse(nodes)
  return result
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
    const res: any = await getRoleList(queryForm)
    if (res && res.code === 0) {
      const list = res.data.list || []
      tableData.value = list.map((item: any) => ({
        ...item,
        createdAt: formatDateTime(item.createdAt)
      }))
      total.value = res.data.total || 0
    } else {
      ElMessage.error(res?.message || '获取角色列表失败')
    }
  } catch (error) {
    ElMessage.error('获取角色列表失败')
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  dialogTitle.value = '新增角色'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  dialogTitle.value = '编辑角色'
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除角色 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res: any = await deleteRole(row.id)
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

const handlePermission = async (row: any) => {
  currentRoleId.value = row.id
  // 角色1（管理员）权限永远不可改；只能修改角色序号 >= 当前用户最小角色序号的角色的权限，否则只读查看
  isPermissionReadonly.value = row.id === 1 || Number(row.id) < currentMinRoleId.value
  permissionVisible.value = true
  try {
    const [menuRes, permRes]: any[] = await Promise.all([
      getMenuTree(),
      getRolePermissions(row.id)
    ])
    if (menuRes && menuRes.code === 0) {
      permissionData.value = menuRes.data || []
    }
    if (permRes && permRes.code === 0) {
      const perms = permRes.data || []
      const leafIds = getLeafNodeIds(permissionData.value)
      const validPerms = perms.filter((id: number) => leafIds.includes(Number(id)))
      nextTick(() => {
        treeRef.value?.setCheckedKeys(validPerms)
      })
    }
  } catch {
    ElMessage.error('获取权限数据失败')
  }
}

const handlePermissionSubmit = async () => {
  if (!currentRoleId.value) return
  const checkedKeys = treeRef.value?.getCheckedKeys() || []
  const halfCheckedKeys = treeRef.value?.getHalfCheckedKeys() || []
  const allKeys = [...checkedKeys, ...halfCheckedKeys]
  const leafIds = getLeafNodeIds(permissionData.value)
  const finalKeys = allKeys.filter((key: number) => leafIds.includes(Number(key)))
  permissionLoading.value = true
  try {
    const res: any = await updateRolePermissions(currentRoleId.value, finalKeys)
    if (res && res.code === 0) {
      ElMessage.success('权限分配成功')
      permissionVisible.value = false
    } else {
      ElMessage.error(res?.message || '权限分配失败')
    }
  } catch {
    ElMessage.error('权限分配失败')
  } finally {
    permissionLoading.value = false
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
      res = await updateRole(form.id, payload)
    } else {
      res = await createRole(payload)
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
  /* 列表页骨架（.search-card / .table-card / .card-header / .pagination）已统一到
     styles/global.scss 的 .page-container 规则，此处不再重复定义（2026-09-22） */
}
</style>
