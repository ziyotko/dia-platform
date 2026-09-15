<template>
  <div class="role-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色管理</span>
          <el-button type="primary" @click="handleAdd">新增角色</el-button>
        </div>
      </template>

      <el-form :inline="true" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="角色名称 / 编码" clearable style="width: 200px" @keyup.enter="handleSearch" />
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
        <el-table-column prop="code" label="角色编码" />
        <el-table-column prop="name" label="角色名称" />
        <el-table-column v-if="isSuperAdmin" label="所属租户" min-width="140">
          <template #default="{ row }">{{ tenantName(row.tenantId) }}</template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="primary" @click="handleMenu(row)">分配菜单</el-button>
            <el-button link type="primary" @click="handlePermission(row)">分配权限</el-button>
            <el-button v-if="!isProtectedRole(row)" link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑角色' : '新增角色'" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item v-if="isSuperAdmin" label="所属租户" prop="tenantId">
          <tenant-select v-model="form.tenantId" :disabled="isEdit" :clearable="!isEdit" />
          <div class="form-tip">编辑时不允许变更角色所属租户</div>
        </el-form-item>
        <el-form-item label="角色编码" prop="code">
          <el-input v-model="form.code" placeholder="留空自动生成" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="menuDialogVisible" title="分配菜单" width="500px">
      <el-alert
        title="决定菜单（页面）的可见性；接口访问权限请在「分配权限」中设置"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 12px"
      />
      <el-tree
        ref="menuTreeRef"
        :data="menuTree"
        show-checkbox
        node-key="id"
        :props="{ label: 'name', children: 'children' }"
        default-expand-all
      />
      <template #footer>
        <el-button @click="menuDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveMenus">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="permDialogVisible" title="分配接口权限" width="600px">
      <el-alert
        title="决定可以调用哪些接口；未分配任何权限的普通用户仅能浏览（GET）"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 12px"
      />
      <el-tree
        ref="permTreeRef"
        :data="permTree"
        show-checkbox
        node-key="id"
        :props="{ label: 'name', children: 'children' }"
        default-expand-all
      />
      <template #footer>
        <el-button @click="permDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSavePermissions">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getRoleList,
  getRole,
  createRole,
  updateRole,
  deleteRole,
  assignRoleMenus,
  assignRolePermissions
} from '@/api/role'
import { getMenuTree } from '@/api/menu'
import { getPermissionTree } from '@/api/permission'
import type { Role } from '@/api/role'
import type { Menu } from '@/api/menu'
import type { Permission } from '@/api/permission'
import { tenantName, ensureTenants } from '@/utils/tenantOptions'
import { useUserStore } from '@/stores/user'
import TenantSelect from '@/components/TenantSelect.vue'

const userStore = useUserStore()
// 仅平台超级管理员（tenantId === 0）可以跨租户管理
const isSuperAdmin = computed(() => userStore.userInfo?.tenantId === 0)

const loading = ref(false)
const tableData = ref<Role[]>([])
const total = ref(0)
const query = reactive<{ page: number; size: number; keyword: string; tenantId?: number }>({
  page: 1,
  size: 10,
  keyword: '',
  tenantId: undefined
})
const dialogVisible = ref(false)
const menuDialogVisible = ref(false)
const permDialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<any>(null)
const menuTreeRef = ref<any>(null)
const permTreeRef = ref<any>(null)
const menuTree = ref<Menu[]>([])
const permTree = ref<Permission[]>([])
const currentRoleId = ref(0)
const form = reactive<Role>({
  id: 0,
  tenantId: 0,
  code: '',
  name: '',
  status: 1,
  remark: ''
})

const rules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }]
}

// 平台内置超级管理员角色（tenant_id=0 且 code=super_admin）不允许删除
const isProtectedRole = (row: Role) => !row.tenantId && row.code === 'super_admin'

// el-tree 的 setCheckedKeys 传入父节点会连带勾选其全部子节点，
// 因此回显时只保留末级节点，目录类父级由 el-tree 自行推导半选/全选状态。
const leafIdsOf = (tree: any[], ids: number[]): number[] => {
  const idSet = new Set(ids)
  const leaves: number[] = []
  const walk = (nodes: any[]) => {
    for (const node of nodes) {
      const children = node.children || []
      if (children.length === 0) {
        if (idSet.has(node.id)) leaves.push(node.id)
      } else {
        walk(children)
      }
    }
  }
  walk(tree)
  return leaves
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getRoleList(query)
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

const handleEdit = (row: Role) => {
  isEdit.value = true
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row: Role) => {
  await ElMessageBox.confirm('确认删除该角色？', '提示', { type: 'warning' })
  await deleteRole(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

const handleMenu = async (row: Role) => {
  currentRoleId.value = row.id
  // 列表接口不返回菜单关联，必须取详情才能正确回显（否则保存会清空原有菜单）
  const [menuRes, roleRes]: any[] = await Promise.all([getMenuTree(), getRole(row.id)])
  menuTree.value = menuRes.data || []
  menuDialogVisible.value = true
  await nextTick()
  menuTreeRef.value?.setCheckedKeys(leafIdsOf(menuTree.value, (roleRes.data?.menus || []).map((m: any) => m.id)))
}

const handleSaveMenus = async () => {
  const ids = menuTreeRef.value?.getCheckedKeys(false)
  await assignRoleMenus(currentRoleId.value, ids)
  ElMessage.success('菜单分配成功')
  menuDialogVisible.value = false
}

const handlePermission = async (row: Role) => {
  currentRoleId.value = row.id
  const [permRes, roleRes]: any[] = await Promise.all([getPermissionTree('base'), getRole(row.id)])
  permTree.value = permRes.data || []
  permDialogVisible.value = true
  await nextTick()
  permTreeRef.value?.setCheckedKeys(leafIdsOf(permTree.value, (roleRes.data?.permissions || []).map((p: any) => p.id)))
}

const handleSavePermissions = async () => {
  const ids = permTreeRef.value?.getCheckedKeys(false)
  await assignRolePermissions(currentRoleId.value, ids)
  ElMessage.success('权限分配成功')
  permDialogVisible.value = false
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (isEdit.value) {
    await updateRole(form.id, form)
  } else {
    await createRole({ ...form, tenantId: isSuperAdmin.value ? form.tenantId : undefined })
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetchData()
}

const resetForm = () => {
  form.id = 0
  form.tenantId = userStore.userInfo?.tenantId || 0
  form.code = ''
  form.name = ''
  form.status = 1
  form.remark = ''
}

onMounted(() => {
  fetchData()
  if (isSuperAdmin.value) {
    ensureTenants()
  }
})
</script>

<style scoped lang="scss">
.role-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
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
}
</style>
