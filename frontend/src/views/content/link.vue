<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="网站名称">
          <el-input v-model="queryForm.name" placeholder="请输入网站名称" clearable />
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
          <span>友链列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增友链
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="name" label="网站名称" min-width="150" />
        <el-table-column prop="url" label="网站链接" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link :href="row.url" target="_blank" type="primary">{{ row.url }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="logo" label="Logo" width="80" align="center">
          <template #default="{ row }">
            <el-avatar v-if="row.logo" :size="36" :src="row.logo" />
            <el-avatar v-else :size="36">{{ row.name.charAt(0) }}</el-avatar>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="网站描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="(val: number) => handleStatusChange(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
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
        <el-form-item label="网站名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入网站名称" />
        </el-form-item>
        <el-form-item label="网站链接" prop="url">
          <el-input v-model="form.url" placeholder="请输入网站链接" />
        </el-form-item>
        <el-form-item label="Logo" prop="logo">
          <el-input v-model="form.logo" placeholder="请输入Logo URL" />
        </el-form-item>
        <el-form-item label="网站描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入网站描述" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="999" />
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
  Delete
} from '@element-plus/icons-vue'

const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const total = ref(0)
const formRef = ref()

const queryForm = reactive({
  page: 1,
  pageSize: 10,
  name: '',
  status: undefined as number | undefined
})

const form = reactive({
  id: undefined as number | undefined,
  name: '',
  url: '',
  logo: '',
  description: '',
  sort: 0,
  status: 1
})

const formRules = {
  name: [{ required: true, message: '请输入网站名称', trigger: 'blur' }],
  url: [
    { required: true, message: '请输入网站链接', trigger: 'blur' },
    { pattern: /^https?:\/\/.+/, message: '链接格式不正确', trigger: 'blur' }
  ]
}

const tableData = ref<any[]>([])

const fetchData = async () => {
  loading.value = true
  try {
    const mockData = [
      { id: 1, name: 'Vue.js 官方', url: 'https://vuejs.org', logo: 'https://vuejs.org/logo.svg', description: '渐进式 JavaScript 框架', sort: 1, status: 1, createTime: '2026-01-10 09:00:00' },
      { id: 2, name: 'Element Plus', url: 'https://element-plus.org', logo: '', description: '基于 Vue 3 的组件库', sort: 2, status: 1, createTime: '2026-01-15 10:30:00' },
      { id: 3, name: 'Vite', url: 'https://vitejs.dev', logo: '', description: '下一代前端构建工具', sort: 3, status: 1, createTime: '2026-02-01 14:00:00' },
      { id: 4, name: 'MDN Web Docs', url: 'https://developer.mozilla.org', logo: '', description: 'Web 技术文档', sort: 4, status: 0, createTime: '2026-03-10 08:20:00' }
    ]
    tableData.value = mockData
    total.value = mockData.length
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  queryForm.page = 1
  fetchData()
}

const resetQuery = () => {
  queryForm.name = ''
  queryForm.status = undefined
  queryForm.page = 1
  fetchData()
}

const handleAdd = () => {
  dialogTitle.value = '新增友链'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  dialogTitle.value = '编辑友链'
  Object.assign(form, {
    id: row.id,
    name: row.name,
    url: row.url,
    logo: row.logo,
    description: row.description,
    sort: row.sort,
    status: row.status
  })
  dialogVisible.value = true
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除友链 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleStatusChange = async (_row: any, val: number) => {
  ElMessage.success(`友链状态已${val === 1 ? '启用' : '禁用'}`)
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  setTimeout(() => {
    ElMessage.success(form.id ? '修改成功' : '新增成功')
    dialogVisible.value = false
    fetchData()
    submitLoading.value = false
  }, 500)
}

const resetForm = () => {
  form.id = undefined
  form.name = ''
  form.url = ''
  form.logo = ''
  form.description = ''
  form.sort = 0
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
