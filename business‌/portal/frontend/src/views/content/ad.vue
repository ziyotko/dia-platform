<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="广告名称">
          <el-input v-model="queryForm.name" placeholder="请输入广告名称" clearable />
        </el-form-item>
        <el-form-item label="广告位置">
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
        <el-table-column label="广告位置" min-width="180">
          <template #default="{ row }">
            <div v-if="row.pageName">
              <el-tag size="small" type="info">{{ row.pageName }}</el-tag>
              <el-icon class="position-arrow"><ArrowRight /></el-icon>
              <el-tag size="small">{{ row.columnName || '默认' }}</el-tag>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="image" label="广告图片" width="80" align="center">
          <template #default="{ row }">
            <el-image
              v-if="row.image"
              :src="row.image"
              :preview-src-list="[row.image]"
              fit="cover"
              style="width: 60px; height: 60px; border-radius: 8px"
            />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="link" label="跳转链接" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link v-if="row.link" :href="row.link" target="_blank" type="primary">{{ row.link }}</el-link>
            <span v-else>-</span>
          </template>
        </el-table-column>
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
      width="640px"
      destroy-on-close
      :close-on-click-modal="false"
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
        <el-form-item label="广告位置" prop="pageId">
          <el-row :gutter="8" style="width: 100%">
            <el-col :span="12">
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
            </el-col>
            <el-col :span="12">
              <el-select
                v-model="form.columnId"
                placeholder="选择栏目"
                style="width: 100%"
                :disabled="!form.pageId"
                clearable
              >
                <el-option
                  v-for="item in formColumnList"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </el-select>
            </el-col>
          </el-row>
        </el-form-item>
        <el-form-item label="跳转链接" prop="link">
          <el-input v-model="form.link" placeholder="请输入跳转链接" />
        </el-form-item>
        <el-form-item label="广告图片" prop="image">
          <div class="ad-image-uploader">
            <el-upload
              v-if="!form.image"
              class="image-uploader"
              :http-request="handleImageUpload"
              :show-file-list="false"
              accept="image/*"
            >
              <el-icon class="uploader-icon"><Plus /></el-icon>
              <div class="uploader-text">点击上传</div>
            </el-upload>
            <div v-else class="image-preview">
              <el-image
                :src="form.image"
                fit="cover"
                style="width: 200px; height: 120px; border-radius: 8px"
                :preview-src-list="[form.image]"
              />
              <div class="image-actions">
                <el-button type="danger" size="small" @click="handleRemoveImage">
                  <el-icon><Delete /></el-icon>删除
                </el-button>
              </div>
            </div>
          </div>
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
                <el-radio :value="1">上架</el-radio>
                <el-radio :value="0">下架</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="开始时间" prop="startTime">
              <el-date-picker
                v-model="form.startTime"
                type="datetime"
                placeholder="选择开始时间"
                style="width: 100%"
                value-format="YYYY-MM-DD HH:mm:ss"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="结束时间" prop="endTime">
              <el-date-picker
                v-model="form.endTime"
                type="datetime"
                placeholder="选择结束时间"
                style="width: 100%"
                value-format="YYYY-MM-DD HH:mm:ss"
              />
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
  getAds,
  createAd,
  updateAd,
  deleteAd,
  updateAdStatus
} from '@/api/ad'
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
  pageId: undefined as number | undefined,
  columnId: undefined as number | undefined,
  image: '',
  link: '',
  sort: 0,
  status: 1,
  startTime: '',
  endTime: ''
})

const formRules = {
  name: [{ required: true, message: '请输入广告名称', trigger: 'blur' }],
  pageId: [{ required: true, message: '请选择广告位置', trigger: 'change' }],
  link: [{ required: true, message: '请输入跳转链接', trigger: 'blur' }]
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
    const res: any = await getColumns({ pageId })
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
    const res: any = await getAds(params)
    const data = res.data || {}
    tableData.value = data.list || []
    total.value = data.total || 0
  } catch (error: any) {
    ElMessage.error(error?.message || '获取广告列表失败')
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
  dialogTitle.value = '新增广告'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = async (row: any) => {
  dialogTitle.value = '编辑广告'
  resetForm()
  if (row.pageId) {
    await loadColumnsByPage(row.pageId, 'form')
  }
  Object.assign(form, {
    id: row.id,
    name: row.name,
    pageId: row.pageId,
    columnId: row.columnId || undefined,
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
  }).then(async () => {
    try {
      await deleteAd(row.id)
      ElMessage.success('删除成功')
      fetchData()
    } catch (error: any) {
      ElMessage.error(error?.message || '删除失败')
    }
  })
}

const handleStatusChange = async (row: any, val: number) => {
  try {
    await updateAdStatus(row.id, val)
    ElMessage.success(`广告已${val === 1 ? '上架' : '下架'}`)
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
      pageId: form.pageId as number,
      columnId: form.columnId,
      image: form.image,
      link: form.link,
      sort: form.sort,
      status: form.status,
      startTime: form.startTime || undefined,
      endTime: form.endTime || undefined
    }
    if (form.id) {
      await updateAd(form.id, payload)
      ElMessage.success('修改成功')
    } else {
      await createAd(payload)
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
  form.pageId = undefined
  form.columnId = undefined
  form.image = ''
  form.link = ''
  form.sort = 0
  form.status = 1
  form.startTime = ''
  form.endTime = ''
  formColumnList.value = []
  formRef.value?.resetFields()
}

const handleImageUpload = async (options: any) => {
  try {
    const res: any = await uploadFile(options.file, 'ad')
    if (res.code === 0) {
      form.image = res.data.url
      ElMessage.success('上传成功')
    } else {
      ElMessage.error(res.message || '上传失败')
    }
  } catch {
    ElMessage.error('上传失败')
  }
}

const handleRemoveImage = () => {
  form.image = ''
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

  .ad-image-uploader {
    .image-uploader {
      :deep(.el-upload) {
        width: 200px;
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
          border-color: #409eff;
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
    .image-preview {
      display: flex;
      flex-direction: column;
      gap: 8px;
      .image-actions {
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
