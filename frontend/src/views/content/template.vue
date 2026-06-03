<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="模板名称">
          <el-input v-model="queryForm.name" placeholder="请输入模板名称" clearable />
        </el-form-item>
        <el-form-item label="模板类型">
          <el-select v-model="queryForm.type" placeholder="全部类型" clearable style="width: 140px">
            <el-option label="首页" value="home" />
            <el-option label="栏目页" value="column" />
            <el-option label="详情页" value="detail" />
            <el-option label="专题页" value="special" />
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
          <span>模板列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增模板
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="name" label="模板名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="code" label="模板编码" min-width="140" />
        <el-table-column prop="type" label="模板类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.type)" size="small">
              {{ typeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="应用页面数" width="110" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.pageCount > 0 ? 'primary' : 'info'"
              size="small"
              style="cursor: pointer"
              @click="handleViewPages(row)"
            >
              {{ row.pageCount }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
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
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column label="操作" width="260" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDesign(row)">
              <el-icon><Brush /></el-icon>设计
            </el-button>
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

    <!-- 新增/编辑弹窗 -->
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
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="模板名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入模板名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="模板编码" prop="code">
              <el-input v-model="form.code" placeholder="请输入模板编码" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="模板类型" prop="type">
              <el-select v-model="form.type" placeholder="请选择模板类型" style="width: 100%">
                <el-option label="首页" value="home" />
                <el-option label="栏目页" value="column" />
                <el-option label="详情页" value="detail" />
                <el-option label="专题页" value="special" />
              </el-select>
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
        <el-form-item label="模板描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入模板描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 模板设计弹窗 -->
    <el-dialog
      v-model="designDialogVisible"
      :title="`模板设计 - ${designForm.name}`"
      width="900px"
      top="5vh"
      destroy-on-close
    >
      <el-tabs v-model="activeDesignTab" type="border-card">
        <el-tab-pane label="可视化布局" name="visual">
          <div class="design-workspace">
            <div class="component-sidebar">
              <div class="sidebar-title">组件库</div>
              <div class="component-list">
                <div
                  v-for="comp in componentLibrary"
                  :key="comp.type"
                  class="component-item"
                  draggable="true"
                  @dragstart="handleDragStart(comp)"
                >
                  <el-icon :size="20"><component :is="comp.icon" /></el-icon>
                  <span class="component-name">{{ comp.label }}</span>
                </div>
              </div>
            </div>
            <div
              class="canvas-area"
              :class="{ dragging: isDragging }"
              @dragover.prevent
              @drop="handleDrop"
              @dragenter="isDragging = true"
              @dragleave="isDragging = false"
            >
              <template v-if="canvasItems.length">
                <div
                  v-for="(item, index) in canvasItems"
                  :key="item.id"
                  class="canvas-item"
                  :class="{ active: activeCanvasIndex === index }"
                  @click="activeCanvasIndex = index"
                >
                  <div class="item-header">
                    <span class="item-label">{{ item.label }}</span>
                    <div class="item-actions">
                      <el-icon class="action-icon" @click.stop="moveUp(index)" v-if="index > 0"><ArrowUp /></el-icon>
                      <el-icon class="action-icon" @click.stop="moveDown(index)" v-if="index < canvasItems.length - 1"><ArrowDown /></el-icon>
                      <el-icon class="action-icon delete" @click.stop="removeItem(index)"><Delete /></el-icon>
                    </div>
                  </div>
                  <div class="item-preview">
                    <component :is="item.icon" :size="32" />
                    <span>{{ item.label }} 组件</span>
                  </div>
                </div>
              </template>
              <div v-else class="canvas-placeholder">
                <el-icon :size="48"><DocumentAdd /></el-icon>
                <p>从左侧拖拽组件到此处进行模板设计</p>
              </div>
            </div>
            <div class="property-panel" v-if="activeCanvasItem">
              <div class="sidebar-title">属性设置</div>
              <el-form label-width="80px" size="small">
                <el-form-item label="组件名称">
                  <el-input v-model="activeCanvasItem.label" disabled />
                </el-form-item>
                <el-form-item label="背景颜色">
                  <el-color-picker v-model="activeCanvasItem.bgColor" show-alpha />
                </el-form-item>
                <el-form-item label="上边距">
                  <el-slider v-model="activeCanvasItem.marginTop" :max="64" show-input />
                </el-form-item>
                <el-form-item label="下边距">
                  <el-slider v-model="activeCanvasItem.marginBottom" :max="64" show-input />
                </el-form-item>
                <el-form-item label="是否全宽">
                  <el-switch v-model="activeCanvasItem.fullWidth" />
                </el-form-item>
              </el-form>
            </div>
            <div class="property-panel" v-else>
              <div class="sidebar-title">属性设置</div>
              <el-empty description="请选择一个组件" :image-size="80" />
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="源码编辑" name="source">
          <div class="source-editor">
            <el-input
              v-model="designForm.sourceCode"
              type="textarea"
              :rows="22"
              placeholder="请输入模板 HTML / Vue 源码..."
              class="code-textarea"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="designDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="designSubmitLoading" @click="handleDesignSave">保存设计</el-button>
      </template>
    </el-dialog>

    <!-- 预览弹窗 -->
    <el-dialog
      v-model="previewDialogVisible"
      title="模板预览"
      width="800px"
      destroy-on-close
    >
      <div class="preview-frame">
        <div class="preview-header">
          <el-icon><Monitor /></el-icon>
          <span>{{ previewRow?.name }}</span>
        </div>
        <div class="preview-body">
          <iframe
            v-if="previewRow"
            :srcdoc="previewHtml"
            frameborder="0"
            width="100%"
            height="400"
          />
        </div>
      </div>
    </el-dialog>

    <!-- 关联页面弹窗 -->
    <el-dialog
      v-model="linkDialogVisible"
      :title="`关联页面 - ${linkTemplateName}`"
      width="700px"
      destroy-on-close
    >
      <el-empty v-if="!linkPages.length" description="暂无关联页面" :image-size="80" />
      <el-table v-else :data="linkPages" border stripe max-height="400">
        <el-table-column prop="name" label="页面名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="code" label="页面编码" min-width="120" />
        <el-table-column prop="routePath" label="访问路径" min-width="140" show-overflow-tooltip />
        <el-table-column prop="pageType" label="页面类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.pageType)" size="small">{{ typeLabel(row.pageType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Edit,
  Delete,
  Search,
  RefreshRight,
  View,
  Brush,
  ArrowUp,
  ArrowDown,
  DocumentAdd,
  Monitor,
  HomeFilled,
  List,
  Document,
  Picture,
  Grid,
  VideoPlay,
  ChatDotSquare,
  Link,
  Calendar
} from '@element-plus/icons-vue'
import {
  getTemplateList,
  createTemplate,
  updateTemplate,
  deleteTemplate,
  updateTemplateStatus,
  saveTemplateDesign
} from '@/api/template'
import { getPages } from '@/api/page'

interface TemplateItem {
  id: number
  name: string
  code: string
  type: 'home' | 'column' | 'detail' | 'special'
  description?: string
  status: number
  pageCount: number
  createTime: string
  sourceCode?: string
  layout?: string
}

interface ComponentItem {
  type: string
  label: string
  icon: any
  id?: string
  bgColor?: string
  marginTop?: number
  marginBottom?: number
  fullWidth?: boolean
}

const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const formRef = ref()

const designDialogVisible = ref(false)
const designSubmitLoading = ref(false)
const activeDesignTab = ref('visual')
const isDragging = ref(false)
const activeCanvasIndex = ref<number | null>(null)
const canvasItems = ref<ComponentItem[]>([])
const draggedComp = ref<ComponentItem | null>(null)

const previewDialogVisible = ref(false)
const previewRow = ref<TemplateItem | null>(null)

const linkDialogVisible = ref(false)
const linkPages = ref<any[]>([])
const linkTemplateName = ref('')

const queryForm = reactive({
  name: '',
  type: '',
  page: 1,
  pageSize: 10
})

const form = reactive<Partial<TemplateItem>>({
  id: undefined,
  name: '',
  code: '',
  type: 'home',
  description: '',
  status: 1
})

const formRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入模板编码', trigger: 'blur' }],
  type: [{ required: true, message: '请选择模板类型', trigger: 'change' }]
}

const designForm = reactive({
  id: undefined as number | undefined,
  name: '',
  sourceCode: ''
})

const componentLibrary: ComponentItem[] = [
  { type: 'header', label: '顶部导航', icon: HomeFilled },
  { type: 'banner', label: '轮播图', icon: Picture },
  { type: 'grid', label: '宫格菜单', icon: Grid },
  { type: 'list', label: '文章列表', icon: List },
  { type: 'video', label: '视频区块', icon: VideoPlay },
  { type: 'notice', label: '公告栏', icon: ChatDotSquare },
  { type: 'link', label: '友情链接', icon: Link },
  { type: 'calendar', label: '日历活动', icon: Calendar },
  { type: 'footer', label: '页脚信息', icon: Document }
]

const tableData = ref<TemplateItem[]>([])
const total = ref(0)

const activeCanvasItem = computed<ComponentItem | null>(() => {
  if (activeCanvasIndex.value === null) return null
  return canvasItems.value[activeCanvasIndex.value] || null
})

const previewHtml = computed(() => {
  if (!previewRow.value) return ''
  return `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f7fa; }
    .preview-block { background: #fff; border-radius: 8px; padding: 24px; margin-bottom: 16px; box-shadow: 0 2px 8px rgba(0,0,0,0.04); }
    .preview-title { font-size: 18px; font-weight: 600; color: #2c3e50; margin-bottom: 12px; }
    .preview-text { color: #606266; line-height: 1.6; }
  </style>
</head>
<body>
  <div class="preview-block">
    <div class="preview-title">${previewRow.value.name}</div>
    <div class="preview-text">编码：${previewRow.value.code}</div>
    <div class="preview-text">类型：${typeLabel(previewRow.value.type)}</div>
    <div class="preview-text">${previewRow.value.description || '暂无描述'}</div>
  </div>
  <div class="preview-block">
    <div class="preview-title">模板预览区域</div>
    <div class="preview-text">此处展示模板实际渲染效果...</div>
  </div>
</body>
</html>
  `.trim()
})

const typeLabel = (type: string) => {
  const map: Record<string, string> = { home: '首页', column: '栏目页', detail: '详情页', special: '专题页' }
  return map[type] || type
}

const typeTagType = (type: string) => {
  const map: Record<string, any> = { home: 'success', column: 'primary', detail: 'warning', special: 'danger' }
  return map[type] || 'info'
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getTemplateList({
      page: queryForm.page,
      pageSize: queryForm.pageSize,
      name: queryForm.name || undefined,
      type: queryForm.type || undefined
    })
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    console.error('获取模板列表失败', error)
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
  queryForm.type = ''
  queryForm.page = 1
  fetchData()
}

const handleSizeChange = () => fetchData()
const handleCurrentChange = () => fetchData()

const handleAdd = () => {
  dialogTitle.value = '新增模板'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: TemplateItem) => {
  dialogTitle.value = '编辑模板'
  Object.assign(form, {
    id: row.id,
    name: row.name,
    code: row.code,
    type: row.type,
    description: row.description,
    status: row.status
  })
  dialogVisible.value = true
}

const handleDelete = (row: TemplateItem) => {
  const msg = row.pageCount > 0
    ? `模板 "${row.name}" 已被 ${row.pageCount} 个页面引用，删除后相关页面将失去模板绑定，确定吗？`
    : `确定要删除模板 "${row.name}" 吗？`
  ElMessageBox.confirm(msg, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deleteTemplate(row.id)
      ElMessage.success('删除成功')
      fetchData()
    } catch (error) {
      console.error('删除模板失败', error)
    }
  })
}

const handleStatusChange = async (row: TemplateItem, val: number) => {
  try {
    await updateTemplateStatus(row.id, val)
    ElMessage.success(`模板状态已${val === 1 ? '启用' : '禁用'}`)
  } catch (error) {
    row.status = val === 1 ? 0 : 1
    console.error('更新状态失败', error)
  }
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    if (form.id) {
      await updateTemplate(form.id, {
        name: form.name || '',
        code: form.code || '',
        type: form.type || 'home',
        description: form.description,
        status: form.status ?? 1
      })
      ElMessage.success('修改成功')
    } else {
      await createTemplate({
        name: form.name || '',
        code: form.code || '',
        type: form.type || 'home',
        description: form.description,
        status: form.status ?? 1
      })
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (error) {
    console.error('提交失败', error)
  } finally {
    submitLoading.value = false
  }
}

const resetForm = () => {
  form.id = undefined
  form.name = ''
  form.code = ''
  form.type = 'home'
  form.description = ''
  form.status = 1
}

const handleDesign = (row: TemplateItem) => {
  designForm.id = row.id
  designForm.name = row.name
  designForm.sourceCode = row.sourceCode || `<template>\n  <!-- ${row.name} -->\n  <div class="template-${row.code}">\n    \n  </div>\n</template>\n`
  canvasItems.value = []
  activeCanvasIndex.value = null
  activeDesignTab.value = 'visual'

  if (row.layout) {
    try {
      const parsed = JSON.parse(row.layout) as Omit<ComponentItem, 'icon'>[]
      canvasItems.value = parsed.map(item => {
        const lib = componentLibrary.find(c => c.type === item.type)
        return { ...item, icon: lib?.icon || Document } as ComponentItem
      })
    } catch {
      canvasItems.value = []
    }
  }

  designDialogVisible.value = true
}

const handleDragStart = (comp: ComponentItem) => {
  draggedComp.value = comp
}

const handleDrop = () => {
  isDragging.value = false
  if (!draggedComp.value) return
  canvasItems.value.push({
    ...draggedComp.value,
    id: Date.now().toString() + Math.random().toString(36).slice(2, 6),
    bgColor: '#ffffff',
    marginTop: 0,
    marginBottom: 0,
    fullWidth: false
  })
  draggedComp.value = null
}

const moveUp = (index: number) => {
  if (index <= 0) return
  const temp = canvasItems.value[index]
  canvasItems.value[index] = canvasItems.value[index - 1]
  canvasItems.value[index - 1] = temp
  if (activeCanvasIndex.value === index) activeCanvasIndex.value = index - 1
  else if (activeCanvasIndex.value === index - 1) activeCanvasIndex.value = index
}

const moveDown = (index: number) => {
  if (index >= canvasItems.value.length - 1) return
  const temp = canvasItems.value[index]
  canvasItems.value[index] = canvasItems.value[index + 1]
  canvasItems.value[index + 1] = temp
  if (activeCanvasIndex.value === index) activeCanvasIndex.value = index + 1
  else if (activeCanvasIndex.value === index + 1) activeCanvasIndex.value = index
}

const removeItem = (index: number) => {
  canvasItems.value.splice(index, 1)
  if (activeCanvasIndex.value === index) activeCanvasIndex.value = null
  else if (activeCanvasIndex.value !== null && activeCanvasIndex.value > index) {
    activeCanvasIndex.value--
  }
}

const handleDesignSave = async () => {
  designSubmitLoading.value = true
  try {
    const layout = JSON.stringify(canvasItems.value.map(item => ({
      type: item.type,
      label: item.label,
      id: item.id,
      bgColor: item.bgColor,
      marginTop: item.marginTop,
      marginBottom: item.marginBottom,
      fullWidth: item.fullWidth
    })))
    await saveTemplateDesign(designForm.id!, {
      sourceCode: designForm.sourceCode,
      layout
    })
    ElMessage.success('模板设计保存成功')
    designDialogVisible.value = false
    fetchData()
  } catch (error) {
    console.error('保存设计失败', error)
  } finally {
    designSubmitLoading.value = false
  }
}

const handlePreview = (row: TemplateItem) => {
  previewRow.value = row
  previewDialogVisible.value = true
}

const handleViewPages = async (row: TemplateItem) => {
  if (!row.pageCount) return
  linkTemplateName.value = row.name
  try {
    const res: any = await getPages({ templateId: row.id })
    linkPages.value = res.data || []
    linkDialogVisible.value = true
  } catch (error) {
    console.error('获取关联页面失败', error)
  }
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

  .design-workspace {
    display: flex;
    gap: 16px;
    height: 520px;

    .component-sidebar {
      width: 160px;
      flex-shrink: 0;
      background: #f5f7fa;
      border-radius: 8px;
      padding: 12px;
      overflow-y: auto;

      .sidebar-title {
        font-size: 14px;
        font-weight: 600;
        color: #2c3e50;
        margin-bottom: 12px;
        padding-bottom: 8px;
        border-bottom: 1px solid #e4e7ed;
      }

      .component-list {
        display: flex;
        flex-direction: column;
        gap: 8px;
      }

      .component-item {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 10px 12px;
        background: #fff;
        border-radius: 8px;
        border: 1px solid #e4e7ed;
        cursor: grab;
        transition: all 0.2s ease;

        &:hover {
          border-color: #409eff;
          color: #409eff;
          box-shadow: 0 2px 8px rgba(64, 158, 255, 0.12);
        }

        .component-name {
          font-size: 13px;
        }
      }
    }

    .canvas-area {
      flex: 1;
      background: #f5f7fa;
      border-radius: 8px;
      border: 2px dashed #dcdfe6;
      padding: 16px;
      overflow-y: auto;
      transition: all 0.3s ease;

      &.dragging {
        border-color: #409eff;
        background: #f0f7ff;
      }

      .canvas-placeholder {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 100%;
        color: #c0c4cc;
        gap: 12px;

        p {
          font-size: 14px;
        }
      }

      .canvas-item {
        background: #fff;
        border-radius: 8px;
        border: 1px solid #e4e7ed;
        margin-bottom: 12px;
        overflow: hidden;
        transition: all 0.3s ease;
        cursor: pointer;

        &:hover {
          border-color: #409eff;
          box-shadow: 0 4px 12px rgba(64, 158, 255, 0.1);
        }

        &.active {
          border-color: #409eff;
          box-shadow: 0 4px 16px rgba(64, 158, 255, 0.15);
        }

        .item-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 10px 14px;
          background: #fafafa;
          border-bottom: 1px solid #f0f0f0;

          .item-label {
            font-size: 13px;
            font-weight: 600;
            color: #2c3e50;
          }

          .item-actions {
            display: flex;
            gap: 8px;

            .action-icon {
              font-size: 14px;
              color: #909399;
              cursor: pointer;
              transition: color 0.2s;

              &:hover {
                color: #409eff;
              }

              &.delete:hover {
                color: #f56c6c;
              }
            }
          }
        }

        .item-preview {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 24px;
          gap: 8px;
          color: #909399;
          font-size: 13px;
        }
      }
    }

    .property-panel {
      width: 220px;
      flex-shrink: 0;
      background: #f5f7fa;
      border-radius: 8px;
      padding: 12px;
      overflow-y: auto;

      .sidebar-title {
        font-size: 14px;
        font-weight: 600;
        color: #2c3e50;
        margin-bottom: 12px;
        padding-bottom: 8px;
        border-bottom: 1px solid #e4e7ed;
      }
    }
  }

  .source-editor {
    .code-textarea {
      :deep(.el-textarea__inner) {
        font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
        font-size: 13px;
        line-height: 1.6;
      }
    }
  }

  .preview-frame {
    .preview-header {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 12px 16px;
      background: #f5f7fa;
      border-radius: 8px 8px 0 0;
      border: 1px solid #e4e7ed;
      border-bottom: none;
      font-weight: 600;
      color: #2c3e50;
    }

    .preview-body {
      border: 1px solid #e4e7ed;
      border-radius: 0 0 8px 8px;
      overflow: hidden;
      background: #fff;
    }
  }
}
</style>
