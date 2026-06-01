<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="文章标题">
          <el-input v-model="queryForm.title" placeholder="请输入文章标题" clearable />
        </el-form-item>
        <el-form-item label="文章分类">
          <el-select v-model="queryForm.categoryId" placeholder="全部分类" clearable style="width: 140px">
            <el-option label="技术文章" :value="1" />
            <el-option label="行业资讯" :value="2" />
            <el-option label="公司动态" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="发布状态">
          <el-select v-model="queryForm.status" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="已发布" :value="1" />
            <el-option label="草稿" :value="0" />
            <el-option label="已下架" :value="2" />
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
          <span>文章列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增文章
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="title" label="文章标题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="categoryName" label="所属分类" width="120" />
        <el-table-column prop="author" label="作者" width="120" />
        <el-table-column prop="views" label="阅读量" width="100" align="center" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : row.status === 0 ? 'info' : 'danger'">
              {{ row.status === 1 ? '已发布' : row.status === 0 ? '草稿' : '已下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="isTop" label="置顶" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.isTop" type="warning">置顶</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handlePreview(row)">
              <el-icon><View /></el-icon>预览
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
      width="700px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="80px"
      >
        <el-form-item label="文章标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入文章标题" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="所属分类" prop="categoryId">
              <el-select v-model="form.categoryId" placeholder="请选择分类" style="width: 100%">
                <el-option label="技术文章" :value="1" />
                <el-option label="行业资讯" :value="2" />
                <el-option label="公司动态" :value="3" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="文章标签" prop="tagIds">
              <el-select v-model="form.tagIds" multiple placeholder="请选择标签" style="width: 100%">
                <el-option label="Vue" :value="1" />
                <el-option label="React" :value="2" />
                <el-option label="Node.js" :value="3" />
                <el-option label="前端" :value="4" />
                <el-option label="后端" :value="5" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="文章摘要" prop="summary">
          <el-input v-model="form.summary" type="textarea" :rows="2" placeholder="请输入文章摘要" />
        </el-form-item>
        <el-form-item label="文章内容" prop="content">
          <el-input v-model="form.content" type="textarea" :rows="6" placeholder="请输入文章内容" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="发布状态" prop="status">
              <el-radio-group v-model="form.status">
                <el-radio :value="1">已发布</el-radio>
                <el-radio :value="0">草稿</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="是否置顶" prop="isTop">
              <el-radio-group v-model="form.isTop">
                <el-radio :value="1">置顶</el-radio>
                <el-radio :value="0">不置顶</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="封面图" prop="cover">
          <el-input v-model="form.cover" placeholder="请输入封面图URL" />
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
  Delete,
  View
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
  title: '',
  categoryId: undefined as number | undefined,
  status: undefined as number | undefined
})

const form = reactive({
  id: undefined as number | undefined,
  title: '',
  categoryId: undefined as number | undefined,
  tagIds: [] as number[],
  summary: '',
  content: '',
  status: 0,
  isTop: 0,
  cover: ''
})

const formRules = {
  title: [{ required: true, message: '请输入文章标题', trigger: 'blur' }],
  categoryId: [{ required: true, message: '请选择所属分类', trigger: 'change' }],
  content: [{ required: true, message: '请输入文章内容', trigger: 'blur' }]
}

const tableData = ref<any[]>([])

const fetchData = async () => {
  loading.value = true
  try {
    const mockData = [
      { id: 1, title: 'Vue3 组合式函数最佳实践', categoryName: '技术文章', author: '张三', views: 1250, status: 1, isTop: 1, createTime: '2026-05-20 10:00:00' },
      { id: 2, title: '2026年前端技术发展趋势', categoryName: '行业资讯', author: '李四', views: 890, status: 1, isTop: 0, createTime: '2026-05-18 14:30:00' },
      { id: 3, title: 'Element Plus 深度定制指南', categoryName: '技术文章', author: '王五', views: 567, status: 0, isTop: 0, createTime: '2026-05-15 09:00:00' },
      { id: 4, title: '公司新产品发布会回顾', categoryName: '公司动态', author: '赵六', views: 2100, status: 1, isTop: 1, createTime: '2026-05-10 16:00:00' }
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
  queryForm.title = ''
  queryForm.categoryId = undefined
  queryForm.status = undefined
  queryForm.page = 1
  fetchData()
}

const handleAdd = () => {
  dialogTitle.value = '新增文章'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  dialogTitle.value = '编辑文章'
  Object.assign(form, {
    id: row.id,
    title: row.title,
    categoryId: 1,
    tagIds: [1, 4],
    summary: '文章摘要示例...',
    content: '文章内容示例...',
    status: row.status,
    isTop: row.isTop,
    cover: ''
  })
  dialogVisible.value = true
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除文章 "${row.title}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handlePreview = (row: any) => {
  ElMessage.info(`预览文章: ${row.title}`)
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
  form.title = ''
  form.categoryId = undefined
  form.tagIds = []
  form.summary = ''
  form.content = ''
  form.status = 0
  form.isTop = 0
  form.cover = ''
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
