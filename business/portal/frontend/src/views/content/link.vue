<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="网站名称">
          <el-input v-model="queryForm.name" placeholder="请输入网站名称" clearable />
        </el-form-item>
        <el-form-item label="友链位置">
          <el-select
            v-model="queryForm.pageId"
            placeholder="选择页面"
            clearable
            style="width: 140px"
            @change="onQueryPageChange"
          >
            <el-option
              v-for="item in pageList"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
          <el-select
            v-model="queryForm.columnId"
            placeholder="选择栏目"
            clearable
            style="width: 140px; margin-left: 8px"
            :disabled="!queryForm.pageId"
          >
            <el-option
              v-for="item in queryColumnList"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
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
          <span>友链列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增友链
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="name" label="网站名称" min-width="150" />
        <el-table-column label="友链位置" min-width="180">
          <template #default="{ row }">
            <div v-if="row.pageName">
              <el-tag size="small" type="info">{{ row.pageName }}</el-tag>
              <el-icon class="position-arrow"><ArrowRight /></el-icon>
              <el-tag size="small">{{ row.columnName || '默认' }}</el-tag>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="url" label="网站链接" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link v-if="row.url" :href="row.url" target="_blank" type="primary">{{ row.url }}</el-link>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="logo" label="Logo" width="80" align="center">
          <template #default="{ row }">
            <el-image
              v-if="row.logo"
              :src="row.logo"
              :preview-src-list="[row.logo]"
              fit="cover"
              style="width: 60px; height: 60px; border-radius: 8px"
            />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="网站描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="author" label="作者" width="100" />
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
        <el-table-column prop="createdAt" label="创建时间" width="170" />
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
      width="560px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="90px"
      >
        <el-form-item label="网站名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入网站名称" />
        </el-form-item>
        <el-form-item label="友链位置" prop="pageId">
          <el-select
            v-model="form.pageId"
            placeholder="选择页面"
            style="width: 100%"
            @change="onFormPageChange"
          >
            <el-option
              v-for="item in pageList"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="栏目" prop="columnId">
          <el-select
            v-model="form.columnId"
            placeholder="请选择栏目"
            style="width: 100%"
            :disabled="!form.pageId"
          >
            <el-option
              v-for="item in formColumnList"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="网站链接" prop="url">
          <el-input v-model="form.url" placeholder="请输入网站链接" />
        </el-form-item>
        <el-form-item label="Logo" prop="logo">
          <div class="link-logo-uploader">
            <el-upload
              v-if="!form.logo"
              class="logo-uploader"
              :http-request="handleLogoUpload"
              :show-file-list="false"
              accept="image/*"
            >
              <el-icon class="uploader-icon"><Plus /></el-icon>
              <div class="uploader-text">点击上传</div>
            </el-upload>
            <div v-else class="logo-preview">
              <el-image
                :src="form.logo"
                fit="cover"
                style="width: 120px; height: 120px; border-radius: 8px"
                :preview-src-list="[form.logo]"
              />
              <div class="logo-actions">
                <el-button type="danger" size="small" @click="handleRemoveLogo">
                  <el-icon><Delete /></el-icon>删除
                </el-button>
              </div>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="网站描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入网站描述" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="排序" prop="sort">
              <el-input-number v-model="form.sort" :min="0" :max="999" style="width: 100%" />
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
  ArrowRight
} from '@element-plus/icons-vue'

import {
  getLinks,
  createLink,
  updateLink,
  deleteLink,
  updateLinkStatus
} from '@/api/link'
import { getPages } from '@/api/page'
import { getColumns } from '@/api/column'
import { uploadFile } from '@/api/upload'

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
  pageId: undefined as number | undefined,
  columnId: undefined as number | undefined,
  status: undefined as number | undefined
})

const form = reactive({
  id: undefined as number | undefined,
  name: '',
  url: '',
  logo: '',
  description: '',
  pageId: undefined as number | undefined,
  columnId: undefined as number | undefined,
  sort: 0,
  status: 1
})

const formRules = {
  name: [{ required: true, message: '请输入网站名称', trigger: 'blur' }],
  pageId: [{ required: true, message: '请选择友链位置', trigger: 'change' }],
  columnId: [{ required: true, message: '请选择栏目', trigger: 'change' }],
  url: [
    { required: true, message: '请输入网站链接', trigger: 'blur' },
    { pattern: /^https?:\/\/.+/, message: '链接格式不正确', trigger: 'blur' }
  ]
}

const tableData = ref<any[]>([])
const pageList = ref<any[]>([])
const queryColumnList = ref<any[]>([])
const formColumnList = ref<any[]>([])

const fetchPages = async () => {
  try {
    const res: any = await getPages({ status: 1 })
    pageList.value = res.data || []
  } catch {
    // ignore
  }
}

const loadColumnsByPage = async (pageId: number | undefined, target: 'query' | 'form') => {
  const list = target === 'query' ? queryColumnList : formColumnList
  list.value = []
  if (!pageId) return
  try {
    // 友链位置只展示栏目类型为「友链展示」的栏目
    const res: any = await getColumns({ pageId, displayType: 5 })
    list.value = res.data || []
  } catch {
    // ignore
  }
}

const onQueryPageChange = (val: number | undefined) => {
  queryForm.columnId = undefined
  loadColumnsByPage(val, 'query')
}

const onFormPageChange = (val: number | undefined) => {
  form.columnId = undefined
  loadColumnsByPage(val, 'form')
}

const fetchData = async () => {
  loading.value = true
  try {
    const params: any = {
      page: queryForm.page,
      pageSize: queryForm.pageSize,
      name: queryForm.name || undefined,
      pageId: queryForm.pageId || undefined,
      columnId: queryForm.columnId || undefined,
      status: queryForm.status !== undefined ? queryForm.status : undefined
    }
    const res: any = await getLinks(params)
    const data = res.data || {}
    tableData.value = data.list || []
    total.value = data.total || 0
  } catch (error: any) {
    ElMessage.error(error?.message || '获取友链列表失败')
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
  queryForm.pageId = undefined
  queryForm.columnId = undefined
  queryForm.status = undefined
  queryForm.page = 1
  queryColumnList.value = []
  fetchData()
}

const handleAdd = () => {
  dialogTitle.value = '新增友链'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = async (row: any) => {
  dialogTitle.value = '编辑友链'
  resetForm()
  if (row.pageId) {
    await loadColumnsByPage(row.pageId, 'form')
  }
  Object.assign(form, {
    id: row.id,
    name: row.name,
    url: row.url,
    logo: row.logo,
    description: row.description,
    pageId: row.pageId,
    columnId: row.columnId || undefined,
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
  }).then(async () => {
    try {
      await deleteLink(row.id)
      ElMessage.success('删除成功')
      fetchData()
    } catch (error: any) {
      ElMessage.error(error?.message || '删除失败')
    }
  })
}

const handleStatusChange = async (row: any, val: number) => {
  try {
    await updateLinkStatus(row.id, val)
    ElMessage.success(`友链已${val === 1 ? '启用' : '禁用'}`)
  } catch (error: any) {
    ElMessage.error(error?.message || '状态更新失败')
    row.status = val === 1 ? 0 : 1
  }
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    const payload = {
      name: form.name,
      url: form.url,
      logo: form.logo,
      description: form.description,
      pageId: form.pageId as number,
      columnId: form.columnId,
      sort: form.sort,
      status: form.status
    }
    if (form.id) {
      await updateLink(form.id, payload)
      ElMessage.success('修改成功')
    } else {
      await createLink(payload)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (error: any) {
    ElMessage.error(error?.message || '提交失败')
  } finally {
    submitLoading.value = false
  }
}

const resetForm = () => {
  form.id = undefined
  form.name = ''
  form.url = ''
  form.logo = ''
  form.description = ''
  form.pageId = undefined
  form.columnId = undefined
  form.sort = 0
  form.status = 1
  formColumnList.value = []
  formRef.value?.resetFields()
}

const handleLogoUpload = async (options: any) => {
  try {
    const res: any = await uploadFile(options.file, 'link')
    if (res.code === 0) {
      form.logo = res.data.url
      ElMessage.success('上传成功')
    } else {
      ElMessage.error(res.message || '上传失败')
    }
  } catch {
    ElMessage.error('上传失败')
  }
}

const handleRemoveLogo = () => {
  form.logo = ''
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
  fetchPages()
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

  .position-arrow {
    margin: 0 4px;
    color: #909399;
    vertical-align: middle;
  }

  .link-logo-uploader {
    .logo-uploader {
      :deep(.el-upload) {
        width: 120px;
        height: 120px;
        border: 2px dashed #d9d9d9;
        border-radius: 8px;
        cursor: pointer;
        position: relative;
        overflow: hidden;
        transition: border-color 0.3s;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        color: #8c939d;
        &:hover {
          border-color: #002fa7;
        }
        .uploader-icon {
          font-size: 28px;
          margin-bottom: 8px;
        }
        .uploader-text {
          font-size: 13px;
        }
      }
    }
    .logo-preview {
      display: flex;
      flex-direction: column;
      gap: 8px;
      .logo-actions {
        display: flex;
        gap: 8px;
      }
    }
  }
}

:global(.el-image-viewer__wrapper) {
  z-index: 9999 !important;
  .el-image-viewer__canvas {
    width: 600px !important;
    height: 500px !important;
    left: 50% !important;
    top: 50% !important;
    transform: translate(-50%, -50%) !important;
  }
  .el-image-viewer__img {
    max-width: 600px !important;
    max-height: 500px !important;
    width: auto !important;
    height: auto !important;
    object-fit: contain !important;
  }
}
</style>
