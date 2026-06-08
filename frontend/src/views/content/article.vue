<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="文章标题">
          <el-input v-model="queryForm.title" placeholder="请输入文章标题" clearable />
        </el-form-item>
        <el-form-item label="文章分类">
          <el-select v-model="queryForm.categoryId" placeholder="全部分类" clearable style="width: 140px">
            <el-option
              v-for="item in categoryList"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="发布状态">
          <el-select v-model="queryForm.status" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="已发布" :value="1" />
            <el-option label="草稿" :value="0" />
            <el-option label="已下线" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="queryForm.auditStatus" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="待审核" :value="0" />
            <el-option label="审核中" :value="1" />
            <el-option label="已审核" :value="2" />
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
        <el-table-column label="标签" width="120">
          <template #default="{ row }">
            <el-tag
              v-for="tag in getRowTags(row)"
              :key="tag.id"
              size="small"
              :color="tag.color"
              style="margin-right: 4px; color: #fff; border: none;"
            >
              {{ tag.name }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="author" label="作者" width="100" />
        <el-table-column prop="source" label="来源" width="140" show-overflow-tooltip />
        <el-table-column prop="views" label="阅读量" width="100" align="center" />
        <el-table-column prop="status" label="发布状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : row.status === 0 ? 'info' : 'danger'">
              {{ row.status === 1 ? '已发布' : row.status === 0 ? '草稿' : '已下线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="auditStatus" label="审核状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.auditStatus === 2 ? 'success' : row.auditStatus === 1 ? 'warning' : 'info'">
              {{ row.auditStatus === 2 ? '已审核' : row.auditStatus === 1 ? '审核中' : '待审核' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="isTop" label="置顶" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.isTop" type="warning">置顶</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="isBold" label="加粗" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.isBold" type="danger">加粗</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="defaultColor" label="颜色" width="80" align="center">
          <template #default="{ row }">
            <span v-if="row.defaultColor" :style="{ display: 'inline-block', width: '20px', height: '20px', backgroundColor: row.defaultColor, borderRadius: '4px', border: '1px solid #dcdfe6' }" />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column prop="staticTime" label="静态化时间" width="170" />
        <el-table-column label="操作" width="280" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handlePreview(row)">
              <el-icon><View /></el-icon>预览
            </el-button>
            <el-button link type="primary" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button v-if="row.status === 0" link type="warning" @click="handleAudit(row)">
              <el-icon><CircleCheck /></el-icon>审核
            </el-button>
            <el-button v-if="row.status === 1" link type="danger" @click="handleOffShelf(row)">
              <el-icon><CircleClose /></el-icon>下线
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
      fullscreen
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="80px"
        class="article-form"
      >
        <el-form-item label="文章标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入文章标题" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="所属分类" prop="categoryId">
              <el-select v-model="form.categoryId" placeholder="请选择分类" style="width: 100%">
                <el-option
                  v-for="item in categoryList"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="文章标签" prop="tagIds">
              <el-select v-model="form.tagIds" multiple placeholder="请选择标签" style="width: 100%">
                <el-option
                  v-for="item in tagList"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="发布状态">
              <el-tag :type="form.status === 1 ? 'success' : form.status === 0 ? 'info' : 'danger'">
                {{ form.status === 1 ? '已发布' : form.status === 0 ? '草稿' : '已下线' }}
              </el-tag>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="封面图" prop="cover">
              <el-input v-model="form.cover" placeholder="请输入封面图URL" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="来源" prop="source">
              <el-input v-model="form.source" placeholder="请输入文章来源" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="是否置顶" prop="isTop">
              <el-radio-group v-model="form.isTop">
                <el-radio :value="1">置顶</el-radio>
                <el-radio :value="0">不置顶</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="是否加粗" prop="isBold">
              <el-radio-group v-model="form.isBold">
                <el-radio :value="1">加粗</el-radio>
                <el-radio :value="0">不加粗</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="默认颜色" prop="defaultColor">
              <el-color-picker v-model="form.defaultColor" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="文章摘要" prop="summary">
          <el-input v-model="form.summary" type="textarea" :rows="2" placeholder="请输入文章摘要" />
        </el-form-item>
        <el-form-item label="文章内容" prop="content" class="editor-form-item">
          <div class="editor-wrapper">
            <Toolbar
              style="border-bottom: 1px solid #ccc"
              :editor="editorRef"
              :defaultConfig="toolbarConfig"
              mode="default"
            />
            <Editor
              v-model="form.content"
              :defaultConfig="editorConfig"
              mode="default"
              @onCreated="handleCreated"
              @onChange="handleEditorChange"
              @customPaste="handleCustomPaste"
            />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="previewVisible" title="文章预览" width="800px" destroy-on-close>
      <div class="preview-content">
        <h2>{{ previewData.title }}</h2>
        <div class="preview-meta">
          <span>作者：{{ previewData.author }}</span>
          <span>来源：{{ previewData.source }}</span>
          <span>分类：{{ previewData.categoryName }}</span>
          <span>时间：{{ previewData.createTime }}</span>
        </div>
        <div class="preview-summary" v-if="previewData.summary">
          <strong>摘要：</strong>{{ previewData.summary }}
        </div>
        <div class="preview-body" v-html="previewData.content" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, shallowRef, onBeforeUnmount, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  RefreshRight,
  Plus,
  Edit,
  Delete,
  View,
  CircleCheck,
  CircleClose
} from '@element-plus/icons-vue'
import '@wangeditor/editor/dist/css/style.css'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import type { IDomEditor, IEditorConfig, IToolbarConfig } from '@wangeditor/editor'

import request from '@/utils/request'
import {
  getArticles,
  createArticle,
  updateArticle,
  deleteArticle,
  auditArticle,
  updateArticleStatus
} from '@/api/article'
import { getAllCategories } from '@/api/category'
import { getAllTags } from '@/api/tag'

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
  status: undefined as number | undefined,
  auditStatus: undefined as number | undefined
})

const form = reactive({
  id: undefined as number | undefined,
  title: '',
  categoryId: undefined as number | undefined,
  tagIds: [] as number[],
  summary: '',
  content: '',
  status: 0,
  auditStatus: 0,
  isTop: 0,
  isBold: 0,
  defaultColor: '',
  cover: '',
  source: ''
})

const formRules = {
  title: [{ required: true, message: '请输入文章标题', trigger: 'blur' }],
  categoryId: [{ required: true, message: '请选择所属分类', trigger: 'change' }],
  content: [{ required: true, message: '请输入文章内容', trigger: 'change' }]
}

const tableData = ref<any[]>([])
const categoryList = ref<any[]>([])
const tagList = ref<any[]>([])

const tagMap = computed(() => {
  const map: Record<number, any> = {}
  tagList.value.forEach(tag => {
    map[tag.id] = tag
  })
  return map
})

const getRowTags = (row: any) => {
  const ids = row.tagIds || []
  return ids.map((id: number) => tagMap.value[id]).filter(Boolean)
}

// 编辑器
const editorRef = shallowRef<IDomEditor>()
const toolbarConfig: Partial<IToolbarConfig> = {}
const base64ToBlob = (base64: string): Blob => {
  const parts = base64.split(',')
  const mime = parts[0].match(/:(.*?);/)?.[1] || 'image/png'
  const bstr = atob(parts[1])
  let n = bstr.length
  const u8arr = new Uint8Array(n)
  while (n--) {
    u8arr[n] = bstr.charCodeAt(n)
  }
  return new Blob([u8arr], { type: mime })
}

const editorConfig: Partial<IEditorConfig> = {
  placeholder: '请输入文章内容...',
  MENU_CONF: {
    uploadImage: {
      async customUpload(file: File, insertFn: (url: string, alt: string, href: string) => void) {
        const formData = new FormData()
        formData.append('file', file)
        try {
          const res: any = await request.post('/upload', formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
          })
          const url = res.data?.url || ''
          if (url) {
            insertFn(url, '', '')
          } else {
            ElMessage.error('图片上传失败')
          }
        } catch {
          ElMessage.error('图片上传失败')
        }
      }
    }
  }
}

const handleCreated = (editor: IDomEditor) => {
  editorRef.value = editor
}

const handleEditorChange = () => {
  if (formRef.value) {
    formRef.value.validateField('content').catch(() => {})
  }
}

const skipRTFGroup = (rtf: string, start: number): number => {
  let i = start
  if (rtf[i] !== '{') return i
  let depth = 1
  i++
  while (i < rtf.length && depth > 0) {
    if (rtf[i] === '{') depth++
    else if (rtf[i] === '}') depth--
    i++
  }
  return i
}

const extractImagesFromRTF = (rtf: string): Blob[] => {
  const blobs: Blob[] = []
  let pos = 0
  while (true) {
    const pictIdx = rtf.indexOf('\\pict', pos)
    if (pictIdx === -1) break
    let i = pictIdx + 5 // 跳过 \pict

    // 跳过 \pict 后面所有参数控制字和嵌套组（如 {\*\picprop ...} \pngblip \picw123 等）
    while (i < rtf.length) {
      if (rtf[i] === '\\') {
        i++
        while (i < rtf.length && /[a-zA-Z*]/.test(rtf[i])) i++
        while (i < rtf.length && /[0-9-]/.test(rtf[i])) i++
        while (i < rtf.length && /\s/.test(rtf[i])) i++
      } else if (rtf[i] === '{') {
        i = skipRTFGroup(rtf, i)
        while (i < rtf.length && /\s/.test(rtf[i])) i++
      } else if (/\s/.test(rtf[i])) {
        i++
      } else if (rtf[i] === '}') {
        break
      } else {
        break
      }
    }

    // 读取十六进制数据
    let hexStr = ''
    while (i < rtf.length) {
      const ch = rtf[i]
      if (/[0-9a-fA-F]/.test(ch)) {
        hexStr += ch
      } else if (/\s/.test(ch)) {
        // skip whitespace
      } else if (ch === '}' || ch === '\\' || ch === '{') {
        break
      } else {
        if (hexStr.length > 0) break
      }
      i++
    }

    if (hexStr.length > 0 && hexStr.length % 2 === 0) {
      const bytes = new Uint8Array(hexStr.length / 2)
      for (let j = 0; j < hexStr.length; j += 2) {
        bytes[j / 2] = parseInt(hexStr.substring(j, j + 2), 16)
      }
      let type = 'image/png'
      if (bytes[0] === 0xFF && bytes[1] === 0xD8) type = 'image/jpeg'
      else if (bytes[0] === 0x47 && bytes[1] === 0x49) type = 'image/gif'
      else if (bytes[0] === 0x89 && bytes[1] === 0x50) type = 'image/png'
      else if (bytes[0] === 0x42 && bytes[1] === 0x4D) type = 'image/bmp'
      blobs.push(new Blob([bytes], { type }))
    }
    pos = pictIdx + 1
  }
  return blobs
}

const handleCustomPaste = (editor: IDomEditor, event: ClipboardEvent) => {
  const html = event.clipboardData?.getData('text/html')
  const rtf = event.clipboardData?.getData('text/rtf')
  const hasBase64 = html && html.includes('data:image')

  // 方案1：HTML 中包含 base64 图片，提取并上传
  if (hasBase64) {
    event.preventDefault()
    const div = document.createElement('div')
    div.innerHTML = html
    const imgs = div.querySelectorAll('img')
    Promise.all(
      Array.from(imgs).map(async (img) => {
        const src = img.getAttribute('src') || ''
        if (src.startsWith('data:image')) {
          try {
            const blob = base64ToBlob(src)
            const ext = blob.type.split('/')[1] || 'png'
            const file = new File([blob], `image.${ext}`, { type: blob.type })
            const formData = new FormData()
            formData.append('file', file)
            const res: any = await request.post('/upload', formData, {
              headers: { 'Content-Type': 'multipart/form-data' }
            })
            const url = res.data?.url || ''
            if (url) {
              img.setAttribute('src', url)
            }
          } catch {
            // 上传失败则保留原 base64
          }
        }
      })
    ).then(() => {
      editor.dangerouslyInsertHtml(div.innerHTML)
    })
    return false
  }

  // 方案1.5：HTML 有 img 但无 base64，且有 RTF，尝试从 RTF 提取图片
  if (html && rtf && html.includes('<img') && !hasBase64) {
    const pictIdx = rtf.indexOf('\\pict')
    console.log('[paste] rtf has \\pict:', pictIdx !== -1)
    if (pictIdx !== -1) {
      console.log('[paste] rtf around \\pict:', rtf.substring(Math.max(0, pictIdx - 100), pictIdx + 300))
    } else {
      console.log('[paste] rtf snippet:', rtf.substring(0, 800))
    }
    const blobs = extractImagesFromRTF(rtf)
    console.log('[paste] RTF images extracted:', blobs.length)
    if (blobs.length > 0) {
      event.preventDefault()
      Promise.all(
        blobs.map((blob) => {
          const ext = blob.type.split('/')[1] || 'png'
          const file = new File([blob], `image.${ext}`, { type: blob.type })
          const formData = new FormData()
          formData.append('file', file)
          return request.post('/upload', formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
          })
        })
      ).then((results: any[]) => {
        const urls = results.map((res) => res.data?.url || '').filter(Boolean)
        const div = document.createElement('div')
        div.innerHTML = html
        const imgs = div.querySelectorAll('img')
        imgs.forEach((img, index) => {
          if (urls[index]) {
            img.setAttribute('src', urls[index])
          }
        })
        // 移除仍然带有本地路径的图片，避免浏览器报 Not allowed to load local resource
        div.querySelectorAll('img').forEach((img) => {
          const src = img.getAttribute('src') || ''
          if (src.startsWith('file://')) {
            img.remove()
          }
        })
        editor.dangerouslyInsertHtml(div.innerHTML)
      })
      return false
    }
  }

  // 方案2：剪贴板中有独立的图片文件，直接上传
  const items = event.clipboardData?.items
  if (items) {
    const imageFiles: File[] = []
    for (let i = 0; i < items.length; i++) {
      if (items[i].type.startsWith('image/')) {
        const file = items[i].getAsFile()
        if (file) imageFiles.push(file)
      }
    }
    if (imageFiles.length > 0) {
      event.preventDefault()
      imageFiles.forEach((file) => {
        const formData = new FormData()
        formData.append('file', file)
        request.post('/upload', formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        }).then((res: any) => {
          const url = res.data?.url || ''
          if (url) {
            editor.insertNode({
              type: 'image',
              src: url,
              alt: '',
              href: '',
              children: [{ text: '' }]
            } as any)
          }
        })
      })
      return false
    }
  }

  return true
}

onBeforeUnmount(() => {
  const editor = editorRef.value
  if (editor == null) return
  editor.destroy()
})

const fetchData = async () => {
  loading.value = true
  try {
    const params: any = {
      page: queryForm.page,
      pageSize: queryForm.pageSize
    }
    if (queryForm.title) params.title = queryForm.title
    if (queryForm.categoryId !== undefined) params.categoryId = queryForm.categoryId
    if (queryForm.status !== undefined) params.status = queryForm.status
    if (queryForm.auditStatus !== undefined) params.auditStatus = queryForm.auditStatus
    const res: any = await getArticles(params)
    tableData.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (error) {
    // request interceptor 已处理错误提示
  } finally {
    loading.value = false
  }
}

const fetchCategories = async () => {
  try {
    const res: any = await getAllCategories()
    categoryList.value = res.data || []
  } catch (error) {
    // ignore
  }
}

const fetchTags = async () => {
  try {
    const res: any = await getAllTags()
    tagList.value = res.data || []
  } catch (error) {
    // ignore
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
  queryForm.auditStatus = undefined
  queryForm.page = 1
  fetchData()
}

const handleAdd = () => {
  dialogTitle.value = '新增文章'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = async (row: any) => {
  dialogTitle.value = '编辑文章'
  resetForm()
  Object.assign(form, {
    id: row.id,
    title: row.title,
    categoryId: row.categoryId,
    tagIds: row.tagIds || [],
    summary: row.summary || '',
    content: row.content || '',
    status: row.status,
    auditStatus: row.auditStatus ?? 0,
    isTop: row.isTop,
    isBold: row.isBold ?? 0,
    defaultColor: row.defaultColor || '',
    cover: row.cover || '',
    source: row.source || ''
  })
  dialogVisible.value = true
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除文章 "${row.title}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await deleteArticle(row.id)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleAudit = (row: any) => {
  ElMessageBox.confirm(`确定要审核通过文章 "${row.title}" 吗？`, '审核确认', {
    confirmButtonText: '通过',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await auditArticle(row.id, 2)
    ElMessage.success('审核成功')
    fetchData()
  })
}

const handleOffShelf = (row: any) => {
  ElMessageBox.confirm(`确定要下线文章 "${row.title}" 吗？`, '下线确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await updateArticleStatus(row.id, 2)
    ElMessage.success('下线成功')
    fetchData()
  })
}

const previewVisible = ref(false)
const previewData = reactive({
  title: '',
  author: '',
  authorCode: '',
  source: '',
  categoryName: '',
  createTime: '',
  summary: '',
  content: ''
})

const handlePreview = (row: any) => {
  Object.assign(previewData, {
    title: row.title,
    author: row.author,
    source: row.source || '',
    categoryName: row.categoryName,
    createTime: row.createTime,
    summary: row.summary || '',
    content: row.content || ''
  })
  previewVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    const data = {
      title: form.title,
      categoryId: form.categoryId as number,
      tagIds: form.tagIds,
      summary: form.summary,
      content: form.content,
      status: form.status,
      auditStatus: form.auditStatus,
      isTop: form.isTop,
      isBold: form.isBold,
      defaultColor: form.defaultColor,
      cover: form.cover,
      source: form.source
    }
    if (form.id) {
      await updateArticle(form.id, data)
      ElMessage.success('修改成功')
    } else {
      await createArticle(data)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } finally {
    submitLoading.value = false
  }
}

const resetForm = () => {
  form.id = undefined
  form.title = ''
  form.categoryId = undefined
  form.tagIds = []
  form.summary = ''
  form.content = ''
  form.status = 0
  form.auditStatus = 0
  form.isTop = 0
  form.isBold = 0
  form.defaultColor = ''
  form.cover = ''
  form.source = ''
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
  fetchCategories()
  fetchTags()
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

.article-form {
  height: 100%;
  display: flex;
  flex-direction: column;

  .editor-form-item {
    flex: 1;
    margin-bottom: 0;

    :deep(.el-form-item__content) {
      height: calc(100vh - 360px);
      display: block;
    }
  }
}

.editor-wrapper {
  border: 1px solid #ccc;
  border-radius: 4px;
  overflow: hidden;
  height: 100%;
  display: flex;
  flex-direction: column;

  :deep(.w-e-text-container) {
    flex: 1;
    overflow-y: auto;
  }
}

.preview-content {
  h2 {
    margin: 0 0 12px 0;
    color: #2c3e50;
  }

  .preview-meta {
    color: #999;
    font-size: 14px;
    margin-bottom: 16px;

    span {
      margin-right: 16px;
    }
  }

  .preview-summary {
    background: #f5f7fa;
    padding: 12px;
    border-radius: 8px;
    margin-bottom: 16px;
    color: #666;
  }

  .preview-body {
    line-height: 1.8;
    color: #333;
  }
}
</style>
