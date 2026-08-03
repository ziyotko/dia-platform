<template>
  <div class="articles-page" v-loading="loading">
    <div class="page-header">
      <h3>我的文章</h3>
      <el-button type="primary" @click="openCreate">发布文章</el-button>
    </div>

    <el-card>
      <div class="filter-bar">
        <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width:140px" @change="fetchData">
          <el-option label="草稿" value="draft" />
          <el-option label="待审核" value="pending" />
          <el-option label="已发布" value="published" />
          <el-option label="已拒绝" value="rejected" />
        </el-select>
        <el-select v-model="filterCat" placeholder="分类筛选" clearable style="width:140px;margin-left:12px" @change="fetchData">
          <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </div>

      <el-table :data="articles" stripe>
        <el-table-column prop="title" label="标题" min-width="200" />
        <el-table-column prop="category.name" label="分类" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240">
          <template #default="{ row }">
            <el-button text type="primary" size="small" @click="editArticle(row)">编辑</el-button>
            <el-button text type="primary" size="small" @click="viewArticle(row)">浏览</el-button>
            <el-button text type="danger" size="small" @click="deleteArticle(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && articles.length === 0" description="暂无文章" />
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog v-model="showDialog" :title="editingId ? '编辑文章' : '发布文章'" width="1000px" top="3vh">
      <el-form :model="articleForm" label-width="80px" size="large">
        <el-form-item label="标题" required><el-input v-model="articleForm.title" /></el-form-item>
        <el-form-item label="分类">
          <el-select v-model="articleForm.categoryId" style="width:100%">
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="封面图">
          <el-upload
            :show-file-list="false"
            :before-upload="uploadCover"
            accept="image/*"
          >
            <img v-if="articleForm.coverImage" :src="articleForm.coverImage" class="cover-preview" />
            <el-button v-else type="default">
              <el-icon><Plus /></el-icon> 上传封面
            </el-button>
          </el-upload>
        </el-form-item>
        <el-form-item label="摘要"><el-input v-model="articleForm.summary" type="textarea" :rows="2" /></el-form-item>
        <el-form-item label="内容" required>
          <div class="editor-wrapper">
            <Toolbar :editor="editorRef" :defaultConfig="toolbarConfig" />
            <Editor v-model="articleForm.content" :defaultConfig="editorConfig" @onCreated="handleCreated" />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button @click="saveArticle(false)">保存草稿</el-button>
        <el-button type="primary" @click="saveArticle(true)">立即提交</el-button>
      </template>
    </el-dialog>

    <!-- Preview Dialog -->
    <el-dialog v-model="showPreview" :title="viewed?.title || '文章预览'" width="800px" top="5vh">
      <div class="preview-body" v-if="viewed">
        <img v-if="viewed.cover_image" :src="viewed.cover_image" class="preview-cover" />
        <div class="preview-summary" v-if="viewed.summary">{{ viewed.summary }}</div>
        <div class="preview-content" v-html="viewed.content || '<p style=&quot;color:#9ca3af&quot;>暂无内容</p>'"></div>
      </div>
      <template #footer>
        <el-button @click="showPreview = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, shallowRef, onMounted, onBeforeUnmount } from 'vue'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import '@wangeditor/editor/dist/css/style.css'
import { articleApi } from '@/api/index'
import { authApi } from '@/api/auth'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

const articles = ref<any[]>([])
const categories = ref<any[]>([])
const loading = ref(true)
const showDialog = ref(false)
const editingId = ref<number | null>(null)
const filterStatus = ref('')
const filterCat = ref<number | ''>('')

const articleForm = reactive({ title: '', categoryId: null as number | null, summary: '', content: '', coverImage: '' })

/* ---- Preview ---- */
const showPreview = ref(false)
const viewed = ref<any>(null)

/* ---- wangEditor ---- */
const editorRef = shallowRef()
const toolbarConfig = {}
const editorConfig = { placeholder: '请输入文章内容...' }

function handleCreated(editor: any) {
  editorRef.value = editor
}

onBeforeUnmount(() => {
  const editor = editorRef.value
  if (editor == null) return
  editor.destroy()
  editorRef.value = null
})

const statusMap: Record<string, { label: string; tag: string }> = {
  draft: { label: '草稿', tag: 'info' },
  pending: { label: '待审核', tag: 'warning' },
  published: { label: '已发布', tag: 'success' },
  rejected: { label: '已拒绝', tag: 'danger' }
}
function statusLabel(s: string) { return statusMap[s]?.label || s }
function statusTag(s: string) { return statusMap[s]?.tag || 'info' as any }

onMounted(async () => {
  try {
    const [artRes, catRes] = await Promise.all([
      articleApi.getMyArticles({ page: 1, size: 100 }),
      articleApi.getCategories()
    ])
    articles.value = artRes.data?.list || []
    categories.value = catRes.data || []
  } catch {} finally { loading.value = false }
})

async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: 1, size: 100 }
    if (filterStatus.value) params.status = filterStatus.value
    if (filterCat.value) params.category_id = filterCat.value
    const res = await articleApi.getMyArticles(params)
    articles.value = res.data?.list || []
  } catch {} finally { loading.value = false }
}

function openCreate() {
  editingId.value = null
  Object.assign(articleForm, { title: '', categoryId: null, summary: '', content: '', coverImage: '' })
  showDialog.value = true
}

function editArticle(row: any) {
  editingId.value = row.id
  Object.assign(articleForm, {
    title: row.title,
    categoryId: row.category_id,
    summary: row.summary,
    content: row.content,
    coverImage: row.cover_image || ''
  })
  showDialog.value = true
}

async function uploadCover(file: File) {
  const fd = new FormData()
  fd.append('file', file)
  fd.append('dir', 'covers')
  try {
    const res = await authApi.upload(fd)
    articleForm.coverImage = res.data?.url || ''
    ElMessage.success('封面上传成功')
  } catch {} finally { return false }
}

function viewArticle(row: any) {
  viewed.value = row
  showPreview.value = true
}

async function saveArticle(submit: boolean) {
  if (!articleForm.title) { ElMessage.warning('请输入标题'); return }
  try {
    const data = {
      title: articleForm.title,
      category_id: articleForm.categoryId,
      summary: articleForm.summary,
      content: articleForm.content,
      cover_image: articleForm.coverImage,
      submit
    }
    if (editingId.value) {
      await articleApi.updateArticle(editingId.value, data)
      ElMessage.success('更新成功')
    } else {
      await articleApi.createArticle(data)
      ElMessage.success(submit ? '提交成功' : '草稿已保存')
    }
    showDialog.value = false
    fetchData()
  } catch {}
}

async function deleteArticle(row: any) {
  try {
    await ElMessageBox.confirm('确认删除该文章？', '提示', { type: 'warning' })
    await articleApi.deleteArticle(row.id)
    ElMessage.success('删除成功')
    fetchData()
  } catch {}
}

function formatDate(d: string) { return d ? d.slice(0, 16) : '' }
</script>

<style scoped lang="scss">
.articles-page {  width: 100%;}
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.filter-bar { margin-bottom: 16px; }
.cover-preview {
  width: 200px;
  height: 120px;
  object-fit: cover;
  border-radius: 6px;
  border: 1px solid #dcdfe6;
  cursor: pointer;
}
.editor-wrapper {
  width: 100%;
  border: 1px solid #dcdfe6;
  :deep(.w-e-text-container) { min-height: 350px; }
}
.preview-body {
  max-height: 70vh;
  overflow-y: auto;
}
.preview-cover {
  width: 100%;
  max-height: 300px;
  object-fit: cover;
  border-radius: 8px;
  margin-bottom: 16px;
}
.preview-summary {
  color: #6b7280;
  font-size: 14px;
  line-height: 1.7;
  background: #f8fafc;
  border-left: 3px solid #409eff;
  padding: 10px 14px;
  border-radius: 4px;
  margin-bottom: 16px;
}
.preview-content {
  line-height: 1.9;
  font-size: 15px;
  color: #303133;
  :deep(img) { max-width: 100%; border-radius: 6px; }
  :deep(p) { margin: 0 0 12px; }
  :deep(blockquote) { border-left: 4px solid #dcdfe6; margin: 12px 0; padding: 8px 14px; color: #6b7280; background: #fafafa; }
}
</style>
