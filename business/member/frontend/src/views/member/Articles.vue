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
        <el-table-column label="操作" width="180">
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
    <el-dialog v-model="showDialog" :title="editingId ? '编辑文章' : '发布文章'" width="720px">
      <el-form :model="articleForm" label-width="80px" size="large">
        <el-form-item label="标题" required><el-input v-model="articleForm.title" /></el-form-item>
        <el-form-item label="分类">
          <el-select v-model="articleForm.categoryId" style="width:100%">
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="摘要"><el-input v-model="articleForm.summary" type="textarea" :rows="2" /></el-form-item>
        <el-form-item label="内容" required>
          <el-input v-model="articleForm.content" type="textarea" :rows="8" placeholder="请输入文章内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button @click="saveArticle(false)">保存草稿</el-button>
        <el-button type="primary" @click="saveArticle(true)">立即提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { articleApi } from '@/api/index'
import { ElMessage, ElMessageBox } from 'element-plus'

const articles = ref<any[]>([])
const categories = ref<any[]>([])
const loading = ref(true)
const showDialog = ref(false)
const editingId = ref<number | null>(null)
const filterStatus = ref('')
const filterCat = ref<number | ''>('')

const articleForm = reactive({ title: '', categoryId: null as number | null, summary: '', content: '' })

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
  Object.assign(articleForm, { title: '', categoryId: null, summary: '', content: '' })
  showDialog.value = true
}

function editArticle(row: any) {
  editingId.value = row.id
  Object.assign(articleForm, {
    title: row.title,
    categoryId: row.category_id,
    summary: row.summary,
    content: row.content
  })
  showDialog.value = true
}

function viewArticle(row: any) {
  ElMessageBox.alert(row.content || '暂无内容', row.title, { confirmButtonText: '关闭' })
}

async function saveArticle(submit: boolean) {
  if (!articleForm.title) { ElMessage.warning('请输入标题'); return }
  try {
    const data = {
      title: articleForm.title,
      category_id: articleForm.categoryId,
      summary: articleForm.summary,
      content: articleForm.content,
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
.articles-page { max-width: 1000px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.filter-bar { margin-bottom: 16px; }
</style>
