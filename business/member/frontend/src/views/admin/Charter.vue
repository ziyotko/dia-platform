<template>
  <div class="admin-charter" v-loading="loading">
    <div class="page-header">
      <div class="title-block">
        <h3>协会章程</h3>
        <p class="sub">使用富文本编辑器维护，保存后会员端「协会章程」页面即时生效</p>
      </div>
      <div class="header-actions">
        <el-button :disabled="!content" @click="showPreview = true">
          <el-icon><View /></el-icon> 预览
        </el-button>
        <el-button :disabled="!content" @click="openPublicPage">
          <el-icon><TopRight /></el-icon> 前台查看
        </el-button>
        <el-button type="primary" :loading="saving" @click="save">
          <el-icon><Check /></el-icon> 保存
        </el-button>
      </div>
    </div>

    <!-- 章程 PDF 附件：供会员端「入会章程」下载（入会申请页 / 服务中心） -->
    <el-card class="file-card">
      <div class="file-head">
        <div>
          <div class="file-title">章程 PDF 附件</div>
          <div class="file-desc">会员端「入会章程」下载的就是这个文件；上传新文件会替换旧文件</div>
        </div>
      </div>
      <div class="file-body">
        <div class="file-info">
          <template v-if="fileMeta">
            <el-icon :size="18"><Document /></el-icon>
            <span class="file-name">{{ fileMeta.name }}</span>
            <span class="file-meta">{{ formatSize(fileMeta.size) }} · 上传于 {{ formatTime(fileMeta.uploaded_at) }}</span>
            <el-link type="primary" :underline="false" @click="downloadPdf">下载</el-link>
            <el-button text type="danger" size="small" @click="removePdf">移除</el-button>
          </template>
          <span v-else class="file-empty">暂未上传 PDF 附件</span>
        </div>
        <el-upload
          :show-file-list="false"
          accept=".pdf"
          :before-upload="beforePdfUpload"
          :http-request="uploadPdf"
        >
          <el-button type="primary" plain :loading="uploading">
            <el-icon><Upload /></el-icon> {{ fileMeta ? '替换 PDF' : '上传 PDF' }}
          </el-button>
          <template #tip>
            <div class="upload-tip">仅支持 PDF 格式，大小不超过 20MB</div>
          </template>
        </el-upload>
      </div>
    </el-card>

    <el-card v-if="!loading">
      <div class="editor-wrapper">
        <Toolbar :editor="editorRef" :defaultConfig="toolbarConfig" />
        <Editor v-model="content" :defaultConfig="editorConfig" @onCreated="handleCreated" />
      </div>
      <div class="editor-tip">
        提示：章程内容支持标题、列表、表格、图片等富文本格式；会员端「协会章程」页面与服务中心按此效果展示。
      </div>
    </el-card>

    <!-- 预览（与会员端展示效果一致） -->
    <el-dialog v-model="showPreview" title="预览（会员端展示效果）" width="900px" top="5vh">
      <div class="preview-title">协会章程</div>
      <div class="rich-text-content" v-html="content"></div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import '@wangeditor/editor/dist/css/style.css'
import { adminApi } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()

const content = ref('')
const loading = ref(true)
const saving = ref(false)
const showPreview = ref(false)

// 章程 PDF 附件
const fileMeta = ref<any>(null)
const uploading = ref(false)
const MAX_PDF_SIZE = 20 * 1024 * 1024

/* ---- wangEditor ---- */
const editorRef = shallowRef()
const toolbarConfig = {}
const editorConfig = { placeholder: '请输入协会章程内容...' }

function handleCreated(editor: any) {
  editorRef.value = editor
}

onBeforeUnmount(() => {
  const editor = editorRef.value
  if (editor == null) return
  editor.destroy()
  editorRef.value = null
})

onMounted(async () => {
  // 注意：加载完成后再渲染 Editor（模板上的 v-if="!loading"），
  // 保证编辑器创建时就带上已有正文，避免先空后填导致的内容不同步。
  try {
    const res = await adminApi.getCharter()
    content.value = res.data?.content || ''
    fileMeta.value = res.data?.file || null
  } catch {} finally { loading.value = false }
})

async function save() {
  saving.value = true
  try {
    await adminApi.saveCharter(content.value)
    ElMessage.success('保存成功')
  } catch {} finally { saving.value = false }
}

/* ---- 章程 PDF 附件 ---- */

function beforePdfUpload(file: File) {
  if (!file.name.toLowerCase().endsWith('.pdf')) {
    ElMessage.warning('仅支持 PDF 文件')
    return false
  }
  if (file.size > MAX_PDF_SIZE) {
    ElMessage.warning('文件大小不能超过 20MB')
    return false
  }
  return true
}

async function uploadPdf(options: any) {
  uploading.value = true
  try {
    const res = await adminApi.uploadCharterFile(options.file)
    fileMeta.value = res.data || null
    ElMessage.success('上传成功')
    options.onSuccess?.(res)
  } catch (e) {
    options.onError?.(e as any)
  } finally {
    uploading.value = false
  }
}

async function removePdf() {
  try {
    await ElMessageBox.confirm('确认移除已上传的章程 PDF？会员端将不可再下载。', '警告', { type: 'warning' })
    await adminApi.deleteCharterFile()
    fileMeta.value = null
    ElMessage.success('已移除')
  } catch {}
}

function downloadPdf() {
  const base = import.meta.env.VITE_API_BASE_URL || '/business_member/api'
  window.open(`${base}/charter`, '_blank')
}

function formatSize(n?: number) {
  if (!n) return '-'
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(2)} MB`
}

function formatTime(s?: string) {
  return s ? s.replace('T', ' ').slice(0, 16) : ''
}

function openPublicPage() {
  window.open(router.resolve('/charter').href, '_blank')
}
</script>

<style scoped lang="scss">
.admin-charter {
  width: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 24px;

  h3 {
    font-size: 22px;
    font-weight: 600;
    color: #1d2739;
    margin: 0;
  }
  .sub {
    margin: 6px 0 0;
    font-size: 13px;
    color: var(--app-text-muted);
  }
  .header-actions {
    display: flex;
    gap: 12px;
    flex-shrink: 0;
  }
}

.el-card {
  border-radius: 10px;
}

.editor-wrapper {
  width: 100%;
  border: 1px solid #dcdfe6;
  :deep(.w-e-text-container) { min-height: 460px; }
}

.editor-tip {
  margin-top: 12px;
  font-size: 13px;
  color: var(--app-text-muted);
}

/* ---- 章程 PDF 附件 ---- */
.file-card {
  border-radius: 10px;
  margin-bottom: 20px;

  .file-title {
    font-size: 15px;
    font-weight: 600;
    color: #1d2739;
  }
  .file-desc {
    margin-top: 4px;
    font-size: 13px;
    color: var(--app-text-muted);
  }
  .file-body {
    margin-top: 16px;
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
  }
  .file-info {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 10px;
    min-height: 32px;
    font-size: 13px;
    color: var(--app-text-secondary);
  }
  .file-name {
    font-size: 14px;
    color: #1d2739;
    font-weight: 500;
    word-break: break-all;
  }
  .file-meta { color: var(--app-text-muted); }
  .file-empty { color: var(--app-text-muted); }
  .upload-tip {
    margin-top: 6px;
    font-size: 12px;
    color: var(--app-text-muted);
  }
}

.preview-title {
  font-size: 20px;
  font-weight: 600;
  text-align: center;
  color: var(--app-text-primary);
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--app-border);
}
</style>
