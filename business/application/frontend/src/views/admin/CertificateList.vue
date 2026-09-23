<template>
  <div class="page-card">
    <div class="page-toolbar">
      <el-input v-model="keyword" placeholder="搜索证书编号/名称/持有人" clearable style="width:260px" @keyup.enter="fetch" @clear="fetch" />
      <el-button type="primary" @click="openDialog">颁发证书</el-button>
    </div>

    <el-table :data="list" v-loading="loading">
      <el-table-column prop="certNo" label="证书编号" width="170" />
      <el-table-column prop="title" label="证书名称" min-width="200" />
      <el-table-column prop="holder" label="持有人" width="110" />
      <el-table-column label="所属项目" min-width="180">
        <template #default="{ row }">{{ row.application?.title || '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="certStatusType[row.status] || 'info'" size="small">{{ certStatusMap[row.status] || row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.fileUrl" size="small" type="primary" link @click="download(row)">下载</el-button>
          <el-button size="small" type="primary" :disabled="row.status === 'void'" @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" :disabled="row.status === 'void'" @click="voidCert(row)">作废</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      style="margin-top:16px;justify-content:flex-end"
      layout="total, prev, pager, next"
      :total="total" :page-size="pageSize" :current-page="page"
      @current-change="onPage"
    />

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑证书' : '颁发证书'" width="560px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="关联申报" v-if="!form.id">
          <el-select v-model="form.applicationId" style="width:100%" filterable placeholder="选择已公示且通过的申报">
            <el-option v-for="a in passedApps" :key="a.id" :label="`${a.title}（${a.user?.realName}）`" :value="a.id" />
          </el-select>
          <div v-if="!passedApps.length" style="color:#909399;font-size:12px">
            暂无可颁发证书的申报（需已公示且通过）
          </div>
        </el-form-item>
        <el-form-item label="证书编号">
          <el-input v-model="form.certNo" :placeholder="form.id ? '请填写证书编号' : '留空则自动生成（CAAM-年份-申报ID）'" />
        </el-form-item>
        <el-form-item label="证书名称"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="持有人"><el-input v-model="form.holder" /></el-form-item>
        <el-form-item label="证书文件">
          <div style="display:flex;gap:8px;width:100%">
            <el-input v-model="form.fileUrl" placeholder="可上传或粘贴文件地址" />
            <el-upload :show-file-list="false" :http-request="handleUpload" accept=".pdf,.doc,.docx,.xls,.xlsx,.jpg,.jpeg,.png">
              <el-button :loading="uploading">上传</el-button>
            </el-upload>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/admin'
import { certStatusMap, certStatusType } from '@/utils/constants'
import { openFile } from '@/utils/file'

const list = ref<any[]>([])
const passedApps = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const keyword = ref('')
const loading = ref(false)
const dialogVisible = ref(false)
const uploading = ref(false)
const form = reactive<any>({ id: 0, applicationId: '', certNo: '', title: '', holder: '', fileUrl: '' })

async function fetch() {
  loading.value = true
  try {
    const res = await adminApi.getCertificates({ page: page.value, pageSize, keyword: keyword.value })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }

async function openDialog() {
  Object.assign(form, { id: 0, applicationId: '', certNo: '', title: '', holder: '', fileUrl: '' })
  const res = await adminApi.getApplications({ page: 1, pageSize: 200, status: 'published' })
  passedApps.value = res.data.list
  dialogVisible.value = true
}

function openEdit(row: any) {
  Object.assign(form, { ...row })
  dialogVisible.value = true
}

async function save() {
  if (!form.title) return ElMessage.warning('请填写证书名称')
  if (form.id) {
    await adminApi.updateCertificate(form.id, { certNo: form.certNo, title: form.title, holder: form.holder, fileUrl: form.fileUrl })
  } else {
    if (!form.applicationId) return ElMessage.warning('请选择关联申报')
    await adminApi.issueCertificate({
      applicationId: Number(form.applicationId), certNo: form.certNo, title: form.title,
      holder: form.holder, fileUrl: form.fileUrl,
    })
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetch()
}

// 作废后申报会退回「已公示」，可以重新颁发正确的证书
async function voidCert(row: any) {
  const { value } = await ElMessageBox.prompt('作废后将通知持有人，且该申报可重新颁发证书。请填写作废原因（可选）：', '作废证书', {
    type: 'warning',
    inputPlaceholder: '如：信息填写错误',
    inputValidator: () => true,
  })
  await adminApi.voidCertificate(row.id, { reason: value || '' })
  ElMessage.success('证书已作废')
  fetch()
}

function download(row: any) {
  openFile(row.fileUrl, 'admin', row.certNo || 'certificate')
}

// Reuses the shared upload endpoint so a certificate file is picked from disk
// like every other attachment instead of being pasted in by hand.
async function handleUpload(option: any) {
  uploading.value = true
  try {
    const res = await adminApi.uploadFile(option.file)
    form.fileUrl = res.data.fileUrl
    ElMessage.success('文件已上传')
  } finally {
    uploading.value = false
  }
}

onMounted(fetch)
</script>
