<template>
  <div class="page-container">
    <el-card shadow="hover" class="table-card">
      <template #header>
        <div class="card-header">
          <span>菜单管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增菜单
          </el-button>
        </div>
      </template>

      <el-table
        :data="tableData"
        v-loading="loading"
        row-key="id"
        border
        stripe
        default-expand-all
      >
        <el-table-column prop="name" label="菜单名称" min-width="160">
          <template #default="{ row }">
            <el-icon v-if="row.icon" style="margin-right: 6px;">
              <component :is="row.icon" />
            </el-icon>
            <span>{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路由路径" min-width="140" />
        <el-table-column prop="component" label="组件路径" min-width="180" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="type" label="类型" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.type === 'directory' ? 'primary' : 'success'" size="small">
              {{ row.type === 'directory' ? '目录' : '菜单' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '显示' : '隐藏' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleAddChild(row)">
              <el-icon><CirclePlus /></el-icon>子菜单
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
        label-width="90px"
      >
        <el-form-item label="上级菜单">
          <el-tree-select
            v-model="form.parentId"
            :data="menuTreeData"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="请选择上级菜单"
            clearable
            check-strictly
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="菜单类型" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio value="directory">目录</el-radio>
            <el-radio value="menu">菜单</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="菜单名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入菜单名称" />
        </el-form-item>
        <el-form-item label="路由路径" prop="path">
          <el-input v-model="form.path" placeholder="请输入路由路径" />
        </el-form-item>
        <el-form-item label="组件路径" v-if="form.type === 'menu'">
          <el-input v-model="form.component" placeholder="请输入组件路径" />
        </el-form-item>
        <el-form-item label="菜单图标">
          <el-input v-model="form.icon" placeholder="请输入图标名称" />
        </el-form-item>
        <el-form-item label="显示排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="显示状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">显示</el-radio>
            <el-radio :value="0">隐藏</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete, CirclePlus } from '@element-plus/icons-vue'

const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const formRef = ref()

const form = reactive({
  id: undefined as number | undefined,
  parentId: undefined as number | undefined,
  name: '',
  path: '',
  component: '',
  icon: '',
  type: 'directory',
  sort: 0,
  status: 1
})

const formRules = {
  name: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
  path: [{ required: true, message: '请输入路由路径', trigger: 'blur' }],
  sort: [{ required: true, message: '请输入排序', trigger: 'blur' }]
}

const tableData = ref([
  {
    id: 1,
    name: '系统管理',
    path: '/system',
    component: '',
    icon: 'Tools',
    type: 'directory',
    sort: 1,
    status: 1,
    children: [
      { id: 11, name: '用户管理', path: '/users', component: 'views/system/users.vue', icon: 'UserFilled', type: 'menu', sort: 1, status: 1 },
      { id: 12, name: '角色管理', path: '/roles', component: 'views/system/roles.vue', icon: 'Avatar', type: 'menu', sort: 2, status: 1 },
      { id: 13, name: '菜单管理', path: '/menus', component: 'views/system/menus.vue', icon: 'Menu', type: 'menu', sort: 3, status: 1 },
      { id: 14, name: '操作日志', path: '/logs', component: 'views/system/logs.vue', icon: 'List', type: 'menu', sort: 4, status: 1 }
    ]
  },
  {
    id: 2,
    name: '个人中心',
    path: '/personal',
    component: '',
    icon: 'User',
    type: 'directory',
    sort: 2,
    status: 1,
    children: [
      { id: 21, name: '个人信息', path: '/profile', component: 'views/profile/index.vue', icon: 'Document', type: 'menu', sort: 1, status: 1 },
      { id: 22, name: '系统设置', path: '/settings', component: 'views/settings/index.vue', icon: 'Setting', type: 'menu', sort: 2, status: 1 }
    ]
  }
])

const menuTreeData = ref([
  { id: 0, name: '顶级菜单' },
  ...tableData.value
])

const handleAdd = () => {
  dialogTitle.value = '新增菜单'
  resetForm()
  dialogVisible.value = true
}

const handleAddChild = (row: any) => {
  dialogTitle.value = '新增子菜单'
  resetForm()
  form.parentId = row.id
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  dialogTitle.value = '编辑菜单'
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除菜单 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success('删除成功')
  })
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  setTimeout(() => {
    submitLoading.value = false
    dialogVisible.value = false
    ElMessage.success(form.id ? '修改成功' : '新增成功')
  }, 500)
}

const resetForm = () => {
  form.id = undefined
  form.parentId = undefined
  form.name = ''
  form.path = ''
  form.component = ''
  form.icon = ''
  form.type = 'directory'
  form.sort = 0
  form.status = 1
}

onMounted(() => {
  loading.value = true
  setTimeout(() => {
    loading.value = false
  }, 300)
})
</script>

<style scoped lang="scss">
.page-container {
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
}
</style>