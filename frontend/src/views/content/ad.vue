<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="广告名称">
          <el-input v-model="queryForm.name" placeholder="请输入广告名称" clearable />
        </el-form-item>
        <el-form-item label="广告位置">
          <el-select v-model="queryForm.position" placeholder="全部位置" clearable style="width: 140px">
            <el-option label="首页轮播" value="home_banner" />
            <el-option label="侧边栏" value="sidebar" />
            <el-option label="文章底部" value="article_bottom" />
            <el-option label="弹窗广告" value="popup" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryForm.status" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="上架" :value="1" />
            <el-option label="下架" :value="0" />
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
          <span>广告列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增广告
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="name" label="广告名称" min-width="160" />
        <el-table-column prop="position" label="广告位置" width="120">
          <template #default="{ row }">
            <el-tag>{{ positionMap[row.position] || row.position }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="image" label="广告图片" width="120" align="center">
          <template #default="{ row }">
            <el-image
              v-if="row.image"
              :src="row.image"
              :preview-src-list="[row.image]"
              fit="cover"
              style="width: 80px; height: 50px; border-radius: 4px"
            />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="link" label="跳转链接" min-width="200" show-overflow-tooltip />
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
        <el-table-column prop="startTime" label="开始时间" width="170" />
        <el-table-column prop="endTime" label="结束时间" width="170" />
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
      width="600px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="90px"
      >
        <el-form-item label="广告名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入广告名称" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="广告位置" prop="position">
              <el-select v-model="form.position" placeholder="请选择位置" style="width: 100%">
                <el-option label="首页轮播" value="home_banner" />
                <el-option label="侧边栏" value="sidebar" />
                <el-option label="文章底部" value="article_bottom" />
                <el-option label="弹窗广告" value="popup" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="排序" prop="sort">
              <el-input-number v-model="form.sort" :min="0" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="跳转链接" prop="link">
          <el-input v-model="form.link" placeholder="请输入跳转链接" />
        </el-form-item>
        <el-form-item label="广告图片" prop="image">
          <el-input v-model="form.image" placeholder="请输入广告图片URL" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="开始时间" prop="startTime">
              <el-date-picker v-model="form.startTime" type="datetime" placeholder="选择开始时间" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="结束时间" prop="endTime">
              <el-date-picker v-model="form.endTime" type="datetime" placeholder="选择结束时间" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">上架</el-radio>
            <el-radio :value="0">下架</el-radio>
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

const positionMap: Record<string, string> = {
  home_banner: '首页轮播',
  sidebar: '侧边栏',
  article_bottom: '文章底部',
  popup: '弹窗广告'
}

const queryForm = reactive({
  page: 1,
  pageSize: 10,
  name: '',
  position: '',
  status: undefined as number | undefined
})

const form = reactive({
  id: undefined as number | undefined,
  name: '',
  position: '',
  image: '',
  link: '',
  sort: 0,
  status: 1,
  startTime: '',
  endTime: ''
})

const formRules = {
  name: [{ required: true, message: '请输入广告名称', trigger: 'blur' }],
  position: [{ required: true, message: '请选择广告位置', trigger: 'change' }],
  link: [{ required: true, message: '请输入跳转链接', trigger: 'blur' }]
}

const tableData = ref<any[]>([])

const fetchData = async () => {
  loading.value = true
  try {
    const mockData = [
      { id: 1, name: '618大促活动', position: 'home_banner', image: 'https://picsum.photos/400/200?random=1', link: 'https://example.com/promo', sort: 1, status: 1, startTime: '2026-05-01 00:00:00', endTime: '2026-06-20 23:59:59' },
      { id: 2, name: '新品上线', position: 'sidebar', image: 'https://picsum.photos/400/200?random=2', link: 'https://example.com/new', sort: 2, status: 1, startTime: '2026-05-15 00:00:00', endTime: '2026-07-15 23:59:59' },
      { id: 3, name: '会员招募', position: 'article_bottom', image: 'https://picsum.photos/400/200?random=3', link: 'https://example.com/vip', sort: 1, status: 0, startTime: '2026-04-01 00:00:00', endTime: '2026-05-31 23:59:59' },
      { id: 4, name: '问卷调查', position: 'popup', image: '', link: 'https://example.com/survey', sort: 1, status: 1, startTime: '2026-05-20 00:00:00', endTime: '2026-06-20 23:59:59' }
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
  queryForm.position = ''
  queryForm.status = undefined
  queryForm.page = 1
  fetchData()
}

const handleAdd = () => {
  dialogTitle.value = '新增广告'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  dialogTitle.value = '编辑广告'
  Object.assign(form, {
    id: row.id,
    name: row.name,
    position: row.position,
    image: row.image,
    link: row.link,
    sort: row.sort,
    status: row.status,
    startTime: row.startTime,
    endTime: row.endTime
  })
  dialogVisible.value = true
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除广告 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleStatusChange = async (_row: any, val: number) => {
  ElMessage.success(`广告已${val === 1 ? '上架' : '下架'}`)
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
  form.position = ''
  form.image = ''
  form.link = ''
  form.sort = 0
  form.status = 1
  form.startTime = ''
  form.endTime = ''
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
