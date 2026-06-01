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
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handlePermission(row)">
              <el-icon><Key /></el-icon>权限
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
      title="分配权限"
      width="500px"
    >
      <el-tree
        ref="treeRef"
        :data="permissionData"
        show-checkbox
        node-key="id"
        :props="{ label: 'name', children: 'children' }"
        :default-expand-all="true"
      />
      <template #footer>
        <el-button @click="permissionVisible = false">取消</el-button>
        <el-button type="primary" @click="handlePermissionSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  RefreshRight,
  Plus,
  Edit,
  Delete,
  Key
} from '@element-plus/icons-vue'

const loading = ref(false)
const dialogVisible = ref(false)
const permissionVisible = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const total = ref(50)
const formRef = ref()
const treeRef = ref()

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

const tableData = ref([
  { id: 1, name: '超级管理员', code: 'super_admin', description: '系统最高权限，可操作所有功能', status: 1, createTime: '2026-01-10 09:00:00' },
  { id: 2, name: '内容管理员', code: 'content_admin', description: '负责内容发布、编辑、审核', status: 1, createTime: '2026-01-15 10:30:00' },
  { id: 3, name: '普通用户', code: 'common_user', description: '普通注册用户权限', status: 1, createTime: '2026-02-01 14:00:00' },
  { id: 4, name: '访客', code: 'visitor', description: '仅浏览权限', status: 0, createTime: '2026-03-10 08:20:00' }
])

const permissionData = ref([
  {
    id: 1,
    name: '系统管理',
    children: [
      { id: 11, name: '用户管理' },
      { id: 12, name: '角色管理' },
      { id: 13, name: '菜单管理' },
      { id: 14, name: '操作日志' }
    ]
  },
  {
    id: 2,
    name: '内容管理',
    children: [
      { id: 21, name: '文章管理' },
      { id: 22, name: '分类管理' },
      { id: 23, name: '标签管理' },
      { id: 24, name: '评论管理' }
    ]
  },
  {
    id: 3,
    name: '个人中心',
    children: [
      { id: 31, name: '个人信息' },
      { id: 32, name: '系统设置' }
    ]
  }
])

const handleSearch = () => {
  queryForm.page = 1
  fetchData()
}

const resetQuery = () => {
  queryForm.name = ''
  queryForm.page = 1
  fetchData()
}

const fetchData = () => {
  loading.value = true
  setTimeout(() => {
    loading.value = false
  }, 500)
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
  }).then(() => {
    ElMessage.success('删除成功')
  })
}

const handlePermission = (_row: any) => {
  permissionVisible.value = true
}

const handlePermissionSubmit = () => {
  treeRef.value?.getCheckedKeys()
  ElMessage.success('权限分配成功')
  permissionVisible.value = false
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  setTimeout(() => {
    submitLoading.value = false
    dialogVisible.value = false
    ElMessage.success(form.id ? '修改成功' : '新增成功')
    fetchData()
  }, 500)
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
}
</style>