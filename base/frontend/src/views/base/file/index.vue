<template>
  <div class="file-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>文件管理</span>
          <el-upload
            action="#"
            :http-request="handleUpload"
            :show-file-list="false"
            accept=".jpg,.jpeg,.png,.gif,.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.zip,.rar,.mp4"
          >
            <el-button type="primary">上传文件</el-button>
          </el-upload>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="fileName" label="文件名" min-width="200" show-overflow-tooltip />
        <el-table-column prop="fileType" label="类型" width="100" />
        <el-table-column prop="fileSize" label="大小" width="120">
          <template #default="{ row }">
            {{ formatSize(row.fileSize) }}
          </template>
        </el-table-column>
        <el-table-column prop="storage" label="存储" width="100" />
        <el-table-column prop="createdAt" label="上传时间" width="180" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handlePreview(row)">预览</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="previewVisible" title="文件预览" width="800px">
      <img v-if="isImage(currentFile?.fileType)" :src="currentFile?.url" style="max-width: 100%" />
      <div v-else>暂不支持该文件类型预览，请下载查看</div>
      <div style="margin-top: 16px; text-align: right">
        <el-button type="primary" @click="handleDownload">下载</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getFileList, deleteFile, uploadFile } from '@/api/file'
import type { UploadedFile, FileQuery } from '@/api/file'

const loading = ref(false)
const tableData = ref<UploadedFile[]>([])
const total = ref(0)
const previewVisible = ref(false)
const currentFile = ref<UploadedFile | null>(null)

const query = reactive<FileQuery>({ page: 1, size: 10 })

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getFileList(query)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    loading.value = false
  }
}

const handleUpload = async (options: any) => {
  try {
    await uploadFile(options.file)
    ElMessage.success('上传成功')
    fetchData()
  } catch (error) {
    ElMessage.error('上传失败')
  }
}

const handlePreview = (row: UploadedFile) => {
  currentFile.value = row
  previewVisible.value = true
}

const handleDownload = () => {
  if (!currentFile.value) return
  const a = document.createElement('a')
  a.href = currentFile.value.url
  a.download = currentFile.value.fileName
  a.click()
}

const handleDelete = (row: UploadedFile) => {
  ElMessageBox.confirm('确认删除该文件？', '提示', { type: 'warning' }).then(async () => {
    await deleteFile(row.id)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const isImage = (type?: string) => {
  return !!type && ['.jpg', '.jpeg', '.png', '.gif'].includes(type.toLowerCase())
}

const formatSize = (size?: number) => {
  if (!size) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let index = 0
  let s = size
  while (s >= 1024 && index < units.length - 1) {
    s /= 1024
    index++
  }
  return `${s.toFixed(2)} ${units[index]}`
}

fetchData()
</script>

<style scoped lang="scss">
.file-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
