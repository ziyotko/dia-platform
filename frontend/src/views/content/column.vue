<template>
  <div class="page-container">
    <div class="type-section">
      <div
        v-for="type in pageTypeList"
        :key="type.value"
        class="type-card"
        :class="{ active: activePageType === type.value }"
        @click="handleTypeChange(type.value)"
      >
        <div class="type-icon" :style="{ backgroundColor: type.bgColor, color: type.color }">
          <el-icon :size="28">
            <HomeFilled v-if="type.value === 'home'" />
            <List v-else-if="type.value === 'column'" />
            <Document v-else-if="type.value === 'detail'" />
            <Grid v-else />
          </el-icon>
        </div>
        <div class="type-info">
          <div class="type-name">{{ type.label }}</div>
          <div class="type-desc">{{ type.description }}</div>
        </div>
        <div class="type-arrow">
          <el-icon><ArrowRight /></el-icon>
        </div>
      </div>
    </div>

    <el-card shadow="hover" class="page-card">
      <template #header>
        <div class="card-header">
          <span>{{ currentTypeLabel }}页面列表</span>
          <el-button type="primary" @click="handleAddPage">
            <el-icon><Plus /></el-icon>新增页面
          </el-button>
        </div>
      </template>

      <el-table
        :data="pageTableData"
        v-loading="pageLoading"
        highlight-current-row
        border
        stripe
        @current-change="handlePageSelect"
      >
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="name" label="页面名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="code" label="页面编码" min-width="140" />
        <el-table-column prop="routePath" label="访问路径" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag v-if="row.routePath" size="small" type="info">{{ row.routePath }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="template" label="绑定模板" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.template">{{ row.template }}</span>
            <span v-else style="color: #c0c4cc">未绑定</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="(val: number) => handlePageStatusChange(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click.stop="handleEditPage(row)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button link type="danger" @click.stop="handleDeletePage(row)">
              <el-icon><Delete /></el-icon>删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card v-if="selectedPage" shadow="hover" class="column-card">
      <template #header>
        <div class="card-header">
          <div class="breadcrumb-title">
            <span class="page-name">{{ selectedPage.name }}</span>
            <el-icon class="breadcrumb-sep"><ArrowRight /></el-icon>
            <span class="list-name">栏目列表</span>
          </div>
          <el-button type="primary" @click="handleAddColumn">
            <el-icon><Plus /></el-icon>新增栏目
          </el-button>
        </div>
      </template>

      <el-table
        :data="columnTableData"
        v-loading="columnLoading"
        row-key="id"
        border
        stripe
        default-expand-all
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      >
        <el-table-column prop="name" label="栏目名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="code" label="栏目编码" min-width="140" />
        <el-table-column prop="routePath" label="访问路径" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag v-if="row.routePath" size="small" type="info">{{ row.routePath }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="displayType" label="展示方式" width="120" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="info">
              {{ displayTypeMap[row.displayType] || '其他展示' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="workflow" label="栏目审核" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag v-if="row.workflow" size="small" type="warning">{{ row.workflow.name }}</el-tag>
            <span v-else style="color: #c0c4cc">未绑定</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="(val: number) => handleColumnStatusChange(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleAddChildColumn(row)">
              <el-icon><CirclePlus /></el-icon>子栏目
            </el-button>
            <el-button link type="primary" @click="handleEditColumn(row)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button link type="danger" @click="handleDeleteColumn(row)">
              <el-icon><Delete /></el-icon>删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!columnTableData.length && !columnLoading" description="该页面下暂无栏目，请添加" />
    </el-card>

    <el-empty v-else description="请先选择左侧页面" class="select-tip" />

    <el-dialog
      v-model="pageDialogVisible"
      :title="pageDialogTitle"
      width="600px"
      destroy-on-close
    >
      <el-form ref="pageFormRef" :model="pageForm" :rules="pageFormRules" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="页面名称" prop="name">
              <el-input v-model="pageForm.name" placeholder="请输入页面名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="页面编码" prop="code">
              <el-input v-model="pageForm.code" placeholder="请输入页面编码" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="页面类型" prop="pageType">
              <el-select v-model="pageForm.pageType" placeholder="请选择页面类型" disabled style="width: 100%">
                <el-option label="首页" value="home" />
                <el-option label="栏目页" value="column" />
                <el-option label="详情页" value="detail" />
                <el-option label="专题页" value="special" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="访问路径" prop="routePath">
              <el-input v-model="pageForm.routePath" placeholder="如 /news" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="绑定模板" prop="templateId">
          <el-select v-model="pageForm.templateId" placeholder="请选择模板" clearable style="width: 100%">
            <el-option-group
              v-for="group in templateGroups"
              :key="group.label"
              :label="group.label"
            >
              <el-option
                v-for="item in group.options"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-option-group>
          </el-select>
        </el-form-item>
        <el-form-item label="页面描述" prop="description">
          <el-input v-model="pageForm.description" type="textarea" :rows="3" placeholder="请输入页面描述" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="pageForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pageDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="pageSubmitLoading" @click="handlePageSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="columnDialogVisible"
      :title="columnDialogTitle"
      width="600px"
      destroy-on-close
    >
      <el-form ref="columnFormRef" :model="columnForm" :rules="columnFormRules" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="所属页面" prop="pageId">
              <el-select v-model="columnForm.pageId" placeholder="请选择所属页面" disabled style="width: 100%">
                <el-option
                  v-for="page in allPages"
                  :key="page.id"
                  :label="page.name"
                  :value="page.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="栏目名称" prop="name">
              <el-input v-model="columnForm.name" placeholder="请输入栏目名称" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="栏目编码" prop="code">
              <el-input v-model="columnForm.code" placeholder="请输入栏目编码" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="排序" prop="sort">
              <el-input-number v-model="columnForm.sort" :min="0" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="上级栏目" prop="parentId">
          <el-tree-select
            v-model="columnForm.parentId"
            :data="columnTreeOptions"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            check-strictly
            clearable
            placeholder="请选择上级栏目（可选）"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="访问路径" prop="routePath">
          <el-input v-model="columnForm.routePath" placeholder="如 /news/company" />
        </el-form-item>
        <el-form-item label="栏目描述" prop="description">
          <el-input v-model="columnForm.description" type="textarea" :rows="3" placeholder="请输入栏目描述" />
        </el-form-item>
        <el-form-item label="展示方式" prop="displayType">
          <el-select v-model="columnForm.displayType" placeholder="请选择展示方式" style="width: 100%">
            <el-option
              v-for="opt in displayTypeOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="栏目审核" prop="workflowId">
          <el-select v-model="columnForm.workflowId" placeholder="请选择审核流程（可选）" clearable style="width: 100%">
            <el-option
              v-for="wf in workflowList"
              :key="wf.id"
              :label="wf.name"
              :value="wf.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="columnForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="columnDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="columnSubmitLoading" @click="handleColumnSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Edit,
  Delete,
  CirclePlus,
  ArrowRight,
  HomeFilled,
  List,
  Document,
  Grid
} from '@element-plus/icons-vue'
import {
  getPages,
  createPage,
  updatePage,
  deletePage
} from '@/api/page'
import {
  getColumns,
  createColumn,
  updateColumn,
  deleteColumn
} from '@/api/column'
import { getTemplateList } from '@/api/template'
import { getWorkflows } from '@/api/workflow'

interface PageItem {
  id: number
  name: string
  code: string
  pageType: 'home' | 'column' | 'detail' | 'special'
  routePath: string
  templateId?: number
  template?: string
  description?: string
  status: number
  createTime: string
}

interface ColumnItem {
  id: number
  name: string
  code: string
  pageId: number
  parentId?: number
  routePath?: string
  description?: string
  sort: number
  status: number
  displayType: number
  workflowId?: number
  workflow?: { id: number; name: string }
  createTime: string
  children?: ColumnItem[]
}

const activePageType = ref<string>('home')
const selectedPage = ref<PageItem | null>(null)

const allPages = ref<PageItem[]>([])
const allColumns = ref<ColumnItem[]>([])

const pageLoading = ref(false)
const pageDialogVisible = ref(false)
const pageDialogTitle = ref('')
const pageSubmitLoading = ref(false)
const pageFormRef = ref()

const columnLoading = ref(false)
const columnDialogVisible = ref(false)
const columnDialogTitle = ref('')
const columnSubmitLoading = ref(false)
const columnFormRef = ref()

const pageForm = reactive<Partial<PageItem>>({
  id: undefined,
  name: '',
  code: '',
  pageType: 'home',
  routePath: '',
  templateId: undefined,
  template: '',
  description: '',
  status: 1
})

const templateList = ref<any[]>([])

const pageFormRules = {
  name: [{ required: true, message: '请输入页面名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入页面编码', trigger: 'blur' }],
  routePath: [{ required: true, message: '请输入访问路径', trigger: 'blur' }]
}

const displayTypeMap: Record<number, string> = {
  1: '轮播展示',
  2: '新闻列表展示',
  3: '图片展示',
  4: '广告展示',
  5: '友链展示',
  6: 'tab页展示',
  9: '其他展示'
}

const displayTypeOptions = [
  { label: '轮播展示', value: 1 },
  { label: '新闻列表展示', value: 2 },
  { label: '图片展示', value: 3 },
  { label: '广告展示', value: 4 },
  { label: '友链展示', value: 5 },
  { label: 'tab页展示', value: 6 },
  { label: '其他展示', value: 9 }
]

const columnForm = reactive<Partial<ColumnItem>>({
  id: undefined,
  pageId: undefined,
  parentId: undefined,
  name: '',
  code: '',
  routePath: '',
  description: '',
  sort: 0,
  status: 1,
  displayType: 1,
  workflowId: undefined
})

const workflowList = ref<any[]>([])

const fetchWorkflows = async () => {
  try {
    const res: any = await getWorkflows({ pageSize: 1000 })
    workflowList.value = (res.data.list || []).filter((w: any) => w.status === 1)
  } catch (error) {
    console.error('获取流程列表失败', error)
  }
}

const columnFormRules = {
  name: [{ required: true, message: '请输入栏目名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入栏目编码', trigger: 'blur' }]
}

const pageTypeList = computed(() => [
  {
    value: 'home',
    label: '首页',
    description: '网站门户页',
    bgColor: '#e6f7ef',
    color: '#52c41a'
  },
  {
    value: 'column',
    label: '栏目页',
    description: '文章栏目聚合页',
    bgColor: '#e6f2ff',
    color: '#409eff'
  },
  {
    value: 'detail',
    label: '详情页',
    description: '内容详情展示页',
    bgColor: '#fff7e6',
    color: '#fa8c16'
  },
  {
    value: 'special',
    label: '专题页',
    description: '专题活动展示页',
    bgColor: '#f0e6ff',
    color: '#722ed1'
  }
])

const currentTypeLabel = computed(() => {
  const map: Record<string, string> = { home: '首页', column: '栏目页', detail: '详情页', special: '专题页' }
  return map[activePageType.value] || ''
})

const pageTableData = computed(() => {
  return allPages.value.filter(p => p.pageType === activePageType.value)
})

const templateGroups = computed(() => {
  const typeMap: Record<string, string> = { home: '首页模板', column: '栏目页模板', detail: '详情页模板', special: '专题页模板' }
  const type = pageForm.pageType || activePageType.value
  const label = typeMap[type] || '模板'
  const options = templateList.value.filter((t: any) => t.type === type)
  if (!options.length) return []
  return [{ label, options }]
})

const isRootColumn = (item: ColumnItem) => !item.parentId || item.parentId === 0

const columnTableData = computed(() => {
  if (!selectedPage.value) return []
  const pageId = selectedPage.value.id
  const buildTree = (items: ColumnItem[], parentId?: number): ColumnItem[] => {
    return items
      .filter(item => {
        if (item.pageId !== pageId) return false
        if (parentId === undefined) return isRootColumn(item)
        return item.parentId === parentId
      })
      .sort((a, b) => a.sort - b.sort)
      .map(item => ({
        ...item,
        children: buildTree(items, item.id)
      }))
  }
  return buildTree(allColumns.value, undefined)
})

const columnTreeOptions = computed(() => {
  if (!selectedPage.value) return []
  const pageId = selectedPage.value.id
  const buildOptions = (items: ColumnItem[], parentId?: number): any[] => {
    return items
      .filter(item => {
        if (item.pageId !== pageId) return false
        if (parentId === undefined) return isRootColumn(item)
        return item.parentId === parentId
      })
      .sort((a, b) => a.sort - b.sort)
      .map(item => ({
        id: item.id,
        name: item.name,
        children: buildOptions(items, item.id)
      }))
  }
  return buildOptions(allColumns.value, undefined)
})

const fetchData = async () => {
  pageLoading.value = true
  try {
    const res: any = await getPages({ pageType: activePageType.value })
    allPages.value = res.data || []
    selectedPage.value = null
  } catch (error) {
    console.error('获取页面列表失败', error)
  } finally {
    pageLoading.value = false
  }
}

const fetchTemplates = async () => {
  try {
    const res: any = await getTemplateList({ pageSize: 100 })
    templateList.value = res.data.list || []
  } catch (error) {
    console.error('获取模板列表失败', error)
  }
}

const fetchColumns = async () => {
  if (!selectedPage.value) return
  columnLoading.value = true
  try {
    const res: any = await getColumns({ pageId: selectedPage.value.id })
    allColumns.value = res.data || []
  } catch (error) {
    console.error('获取栏目列表失败', error)
  } finally {
    columnLoading.value = false
  }
}

watch(() => selectedPage.value, () => {
  fetchColumns()
})

const handleTypeChange = (type: string) => {
  activePageType.value = type
  selectedPage.value = null
  fetchData()
}

const handlePageSelect = (row: PageItem) => {
  selectedPage.value = row
}

const handleAddPage = () => {
  pageDialogTitle.value = '新增页面'
  resetPageForm()
  pageDialogVisible.value = true
}

const handleEditPage = (row: PageItem) => {
  pageDialogTitle.value = '编辑页面'
  Object.assign(pageForm, {
    id: row.id,
    name: row.name,
    code: row.code,
    pageType: row.pageType,
    routePath: row.routePath,
    templateId: row.templateId,
    template: row.template,
    description: row.description,
    status: row.status
  })
  pageDialogVisible.value = true
}

const handleDeletePage = (row: PageItem) => {
  const hasColumns = allColumns.value.some(c => c.pageId === row.id)
  const msg = hasColumns
    ? `页面 "${row.name}" 下存在栏目，删除页面将一并删除其下所有栏目，确定吗？`
    : `确定要删除页面 "${row.name}" 吗？`
  ElMessageBox.confirm(msg, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await deletePage(row.id)
    ElMessage.success('删除成功')
    fetchData()
    if (selectedPage.value?.id === row.id) {
      selectedPage.value = null
    }
  })
}

const handlePageStatusChange = async (_row: PageItem, val: number) => {
  ElMessage.success(`页面状态已${val === 1 ? '启用' : '禁用'}`)
}

const handlePageSubmit = async () => {
  const valid = await pageFormRef.value?.validate().catch(() => false)
  if (!valid) return
  pageSubmitLoading.value = true
  try {
    const selectedTemplate = templateList.value.find((t: any) => t.id === pageForm.templateId)
    const payload = {
      name: pageForm.name || '',
      code: pageForm.code || '',
      pageType: pageForm.pageType || activePageType.value,
      routePath: pageForm.routePath || '',
      templateId: pageForm.templateId,
      template: selectedTemplate?.code || selectedTemplate?.name || '',
      description: pageForm.description,
      status: pageForm.status ?? 1
    }
    if (pageForm.id) {
      await updatePage(pageForm.id, payload)
    } else {
      await createPage(payload)
    }
    ElMessage.success(pageForm.id ? '修改成功' : '新增成功')
    pageDialogVisible.value = false
    fetchData()
  } finally {
    pageSubmitLoading.value = false
  }
}

const resetPageForm = () => {
  pageForm.id = undefined
  pageForm.name = ''
  pageForm.code = ''
  pageForm.pageType = activePageType.value as any
  pageForm.routePath = ''
  pageForm.templateId = undefined
  pageForm.template = ''
  pageForm.description = ''
  pageForm.status = 1
}

const handleAddColumn = () => {
  if (!selectedPage.value) return
  columnDialogTitle.value = '新增栏目'
  resetColumnForm()
  columnForm.pageId = selectedPage.value.id
  columnDialogVisible.value = true
}

const handleAddChildColumn = (row: ColumnItem) => {
  columnDialogTitle.value = `新增子栏目 - ${row.name}`
  resetColumnForm()
  columnForm.pageId = row.pageId
  columnForm.parentId = row.id
  columnDialogVisible.value = true
}

const handleEditColumn = (row: ColumnItem) => {
  columnDialogTitle.value = '编辑栏目'
  Object.assign(columnForm, {
    id: row.id,
    pageId: row.pageId,
    parentId: row.parentId,
    name: row.name,
    code: row.code,
    routePath: row.routePath,
    description: row.description,
    sort: row.sort,
    status: row.status,
    displayType: row.displayType,
    workflowId: row.workflowId
  })
  columnDialogVisible.value = true
}

const handleDeleteColumn = async (row: ColumnItem) => {
  const hasChildren = allColumns.value.some(c => c.parentId === row.id)
  const msg = hasChildren
    ? `栏目 "${row.name}" 下存在子栏目，确定要一并删除吗？`
    : `确定要删除栏目 "${row.name}" 吗？`
  try {
    await ElMessageBox.confirm(msg, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteColumn(row.id)
    ElMessage.success('删除成功')
    fetchColumns()
  } catch (error) {
    // cancel
  }
}

const handleColumnStatusChange = async (row: ColumnItem, val: number) => {
  try {
    await updateColumn(row.id, {
      name: row.name,
      code: row.code,
      pageId: row.pageId,
      parentId: row.parentId,
      routePath: row.routePath,
      description: row.description,
      sort: row.sort,
      status: val,
      displayType: row.displayType,
      workflowId: row.workflowId
    })
    ElMessage.success(`栏目状态已${val === 1 ? '启用' : '禁用'}`)
  } catch (error) {
    row.status = val === 1 ? 0 : 1
  }
}

const handleColumnSubmit = async () => {
  const valid = await columnFormRef.value?.validate().catch(() => false)
  if (!valid) return
  columnSubmitLoading.value = true
  try {
    const payload = {
      name: columnForm.name || '',
      code: columnForm.code || '',
      pageId: columnForm.pageId ?? 0,
      parentId: columnForm.parentId,
      routePath: columnForm.routePath || '',
      description: columnForm.description,
      sort: columnForm.sort ?? 0,
      status: columnForm.status ?? 1,
      displayType: columnForm.displayType ?? 1,
      workflowId: columnForm.workflowId
    }
    if (columnForm.id) {
      await updateColumn(columnForm.id, payload)
    } else {
      await createColumn(payload)
    }
    ElMessage.success(columnForm.id ? '修改成功' : '新增成功')
    columnDialogVisible.value = false
    fetchColumns()
  } finally {
    columnSubmitLoading.value = false
  }
}

const resetColumnForm = () => {
  columnForm.id = undefined
  columnForm.pageId = selectedPage.value?.id
  columnForm.parentId = undefined
  columnForm.name = ''
  columnForm.code = ''
  columnForm.routePath = ''
  columnForm.description = ''
  columnForm.sort = 0
  columnForm.status = 1
  columnForm.displayType = 1
  columnForm.workflowId = undefined
}

onMounted(() => {
  fetchData()
  fetchTemplates()
  fetchWorkflows()
})
</script>

<style scoped lang="scss">
.page-container {
  .type-section {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 20px;
    margin-bottom: 20px;

    .type-card {
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 20px 24px;
      background: #fff;
      border-radius: 12px;
      border: 1px solid #e6f2ff;
      cursor: pointer;
      transition: all 0.3s ease;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 8px 24px rgba(64, 158, 255, 0.12);
        border-color: #409eff;
      }

      &.active {
        border-color: #409eff;
        background: linear-gradient(135deg, #f0f7ff 0%, #ffffff 100%);
        box-shadow: 0 4px 16px rgba(64, 158, 255, 0.15);

        .type-arrow {
          color: #409eff;
          transform: translateX(4px);
        }
      }

      .type-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 56px;
        height: 56px;
        border-radius: 12px;
        flex-shrink: 0;
      }

      .type-info {
        flex: 1;
        min-width: 0;

        .type-name {
          font-size: 16px;
          font-weight: 600;
          color: #2c3e50;
          margin-bottom: 6px;
        }

        .type-desc {
          font-size: 13px;
          color: #909399;
        }
      }

      .type-arrow {
        color: #c0c4cc;
        transition: all 0.3s ease;
        flex-shrink: 0;
      }
    }
  }

  .page-card {
    margin-bottom: 20px;
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

  .column-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-weight: 600;
      color: #2c3e50;

      .breadcrumb-title {
        display: flex;
        align-items: center;
        gap: 8px;

        .page-name {
          font-size: 16px;
          font-weight: 600;
          color: #409eff;
        }

        .breadcrumb-sep {
          color: #c0c4cc;
          font-size: 14px;
        }

        .list-name {
          font-size: 16px;
          color: #2c3e50;
        }
      }
    }
  }

  .select-tip {
    margin-top: 20px;
    border-radius: 12px;
    border: 1px dashed #dcdfe6;
    background: #fafafa;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
