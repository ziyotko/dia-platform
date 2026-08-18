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
      width="96vw"
      top="2vh"
      class="design-dialog"
      destroy-on-close
    >
      <div class="source-editor">
        <el-input
          v-model="designForm.sourceCode"
          type="textarea"
          placeholder="请输入模板 HTML / Vue 源码..."
          class="code-textarea"
        />
      </div>
      <template #footer>
        <el-button @click="designDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="designSubmitLoading" @click="handleDesignSave">保存设计</el-button>
      </template>
    </el-dialog>

    <!-- 预览弹窗 -->
    <el-dialog
      v-model="previewDialogVisible"
      title="模板预览"
      width="96vw"
      top="3vh"
      class="preview-dialog"
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
            style="height: 100%"
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
  Monitor
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

const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const formRef = ref()

const designDialogVisible = ref(false)
const designSubmitLoading = ref(false)

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

const tableData = ref<TemplateItem[]>([])
const total = ref(0)

const previewHtml = computed(() => {
  if (!previewRow.value) return ''

  // 优先使用 sourceCode（提取 <template> 和 <style>）
  if (previewRow.value.sourceCode) {
    const code = previewRow.value.sourceCode
    const templateMatch = code.match(/<template>([\s\S]*?)<\/template>/i)
    const bodyHtml = templateMatch ? templateMatch[1].trim() : code
    const styleMatch = code.match(/<style[^>]*>([\s\S]*?)<\/style>/i)
    const styleCss = styleMatch ? styleMatch[1].trim() : ''

    return `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f8fafc; }
    .h5-container { position: relative; width: 100%; min-height: 100vh; overflow-x: hidden; }
    ${styleCss}
  </style>
</head>
<body>
  <div class="h5-container">
    ${bodyHtml}
  </div>
</body>
</html>
    `.trim()
  }

  // 其次使用 layout 生成预览
  if (previewRow.value.layout) {
    try {
      const items = JSON.parse(previewRow.value.layout) as Array<{
        type: string; label: string; x?: number; y?: number;
        bgColor?: string; fullWidth?: boolean;
      }>
      const componentsHtml = items.map((item, idx) => {
        const width = item.fullWidth ? 'calc(100% - 32px)' : '260px'
        const bg = item.bgColor || '#ffffff'
        const left = item.x ?? 20
        const top = item.y ?? (idx * 140 + 20)
        return `    <div class="comp-block" style="position:absolute;left:${left}px;top:${top}px;width:${width};background:${bg};border-radius:12px;padding:14px 16px;box-shadow:0 2px 8px rgba(0,0,0,0.06);border:1px solid #e2e8f0;">
      <div style="font-weight:700;font-size:13px;color:#1f2937;margin-bottom:4px;">${item.label}</div>
      <div style="font-size:12px;color:#94a3b8;">${item.type} 组件</div>
    </div>`
      }).join('\n')

      return `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f8fafc; }
    .pc-page { position: relative; width: 100%; min-height: 100vh; overflow-x: hidden; }
  </style>
</head>
<body>
  <div class="pc-page">
${componentsHtml}
  </div>
</body>
</html>
      `.trim()
    } catch {
      // 解析失败回退到空提示
    }
  }

  // 无内容提示
  return `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 0; display: flex; align-items: center; justify-content: center; height: 100vh; background: #f8fafc; color: #94a3b8; font-size: 14px; }
  </style>
</head>
<body>
  <div>暂无预览内容，请先进行模板设计</div>
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
    ElMessage.error('获取模板列表失败')
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
      ElMessage.error('删除模板失败')
    }
  })
}

const handleStatusChange = async (row: TemplateItem, val: number) => {
  try {
    await updateTemplateStatus(row.id, val)
    ElMessage.success(`模板状态已${val === 1 ? '启用' : '禁用'}`)
  } catch (error) {
    row.status = val === 1 ? 0 : 1
    ElMessage.error('更新状态失败')
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
    ElMessage.error('提交失败')
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
  designDialogVisible.value = true
}

const handleDesignSave = async () => {
  designSubmitLoading.value = true
  try {
    await saveTemplateDesign(designForm.id!, {
      sourceCode: designForm.sourceCode,
      layout: ''
    })
    ElMessage.success('模板设计保存成功')
    designDialogVisible.value = false
    fetchData()
  } catch (error) {
    ElMessage.error('保存设计失败')
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
    ElMessage.error('获取关联页面失败')
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.page-container {
  padding: 8px;

  .search-card {
    margin-bottom: 20px;
    border-radius: 20px;
    border: none;
    background: #ffffff;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.04);
    :deep(.el-card__body) {
      padding: 20px 24px;
    }
  }

  .table-card {
    border-radius: 20px;
    border: none;
    background: #ffffff;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.04);
    :deep(.el-card__header) {
      padding: 18px 24px;
      border-bottom: 1px solid #f0f2f5;
    }
    :deep(.el-card__body) {
      padding: 20px 24px 24px;
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-weight: 700;
      font-size: 16px;
      color: #1f2937;
      letter-spacing: 0.2px;
    }
  }

  .pagination {
    margin-top: 24px;
    display: flex;
    justify-content: flex-end;
  }

  .source-editor {
    height: calc(96vh - 260px);
    .code-textarea {
      height: 100%;
      :deep(.el-textarea__inner) {
        height: 100% !important;
        min-height: 100% !important;
        font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
        font-size: 13px;
        line-height: 1.7;
        border-radius: 12px;
        padding: 16px;
        background: #f8fafc;
        border: 1px solid #e2e8f0;
      }
    }
  }

  .preview-frame {
    border-radius: 20px;
    overflow: hidden;
    background: #ffffff;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.04);
    border: 1px solid #f1f5f9;

    .preview-header {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 14px 18px;
      background: #f8fafc;
      border-bottom: 1px solid #f1f5f9;
      font-weight: 700;
      font-size: 15px;
      color: #1f2937;
    }

    .preview-body {
      overflow: hidden;
      background: #fff;
      height: calc(94vh - 150px);
      iframe {
        display: block;
        width: 100%;
        height: 100%;
        border: none;
      }
    }
  }
}

/* Dialog global overrides scoped via :deep */
:deep(.el-dialog) {
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.1);
}
:deep(.design-dialog) {
  max-width: 96vw;
  margin-top: 0 !important;
}
:deep(.el-dialog__header) {
  padding: 18px 24px;
  margin-right: 0;
  border-bottom: 1px solid #f1f5f9;
}
:deep(.el-dialog__body) {
  padding: 20px 24px;
}
:deep(.el-dialog__footer) {
  padding: 14px 24px 18px;
  border-top: 1px solid #f1f5f9;
}
</style>
