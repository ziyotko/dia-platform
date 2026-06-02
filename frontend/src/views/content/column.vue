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
            <List v-else-if="type.value === 'list'" />
            <Document v-else />
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
        <el-table-column prop="template" label="绑定模板" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.template">{{ row.template }}</span>
            <span v-else style="color: #c0c4cc">未绑定</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" align="center" />
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
                <el-option label="列表页" value="list" />
                <el-option label="详情页" value="detail" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="访问路径" prop="routePath">
              <el-input v-model="pageForm.routePath" placeholder="如 /news" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="绑定模板" prop="template">
          <el-select v-model="pageForm.template" placeholder="请选择页面模板" clearable style="width: 100%">
            <el-option-group label="首页模板">
              <el-option label="default-home" value="default-home" />
              <el-option label="portal-home" value="portal-home" />
            </el-option-group>
            <el-option-group label="列表页模板">
              <el-option label="article-list" value="article-list" />
              <el-option label="news-list" value="news-list" />
              <el-option label="image-list" value="image-list" />
            </el-option-group>
            <el-option-group label="详情页模板">
              <el-option label="article-detail" value="article-detail" />
              <el-option label="page-detail" value="page-detail" />
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
        <el-form-item label="绑定模板" prop="template">
          <el-select v-model="columnForm.template" placeholder="请选择页面模板" clearable style="width: 100%">
            <el-option-group label="首页模板">
              <el-option label="default-home" value="default-home" />
              <el-option label="portal-home" value="portal-home" />
            </el-option-group>
            <el-option-group label="列表页模板">
              <el-option label="article-list" value="article-list" />
              <el-option label="news-list" value="news-list" />
              <el-option label="image-list" value="image-list" />
            </el-option-group>
            <el-option-group label="详情页模板">
              <el-option label="article-detail" value="article-detail" />
              <el-option label="page-detail" value="page-detail" />
            </el-option-group>
          </el-select>
        </el-form-item>
        <el-form-item label="栏目描述" prop="description">
          <el-input v-model="columnForm.description" type="textarea" :rows="3" placeholder="请输入栏目描述" />
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
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Edit,
  Delete,
  CirclePlus,
  ArrowRight,
  HomeFilled,
  List,
  Document
} from '@element-plus/icons-vue'

interface PageItem {
  id: number
  name: string
  code: string
  pageType: 'home' | 'list' | 'detail'
  routePath: string
  template: string
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
  template?: string
  description?: string
  sort: number
  status: number
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
  template: '',
  description: '',
  status: 1
})

const pageFormRules = {
  name: [{ required: true, message: '请输入页面名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入页面编码', trigger: 'blur' }],
  routePath: [{ required: true, message: '请输入访问路径', trigger: 'blur' }]
}

const columnForm = reactive<Partial<ColumnItem>>({
  id: undefined,
  pageId: undefined,
  parentId: undefined,
  name: '',
  code: '',
  routePath: '',
  template: '',
  description: '',
  sort: 0,
  status: 1
})

const columnFormRules = {
  name: [{ required: true, message: '请输入栏目名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入栏目编码', trigger: 'blur' }]
}

const pageTypeList = computed(() => [
  {
    value: 'home',
    label: '首页',
    description: '网站入口首页',
    bgColor: '#e6f7ef',
    color: '#52c41a'
  },
  {
    value: 'list',
    label: '列表页',
    description: '文章列表聚合页',
    bgColor: '#e6f2ff',
    color: '#409eff'
  },
  {
    value: 'detail',
    label: '详情页',
    description: '内容详情展示页',
    bgColor: '#fff7e6',
    color: '#fa8c16'
  }
])

const currentTypeLabel = computed(() => {
  const map: Record<string, string> = { home: '首页', list: '列表页', detail: '详情页' }
  return map[activePageType.value] || ''
})

const pageTableData = computed(() => {
  return allPages.value.filter(p => p.pageType === activePageType.value)
})

const columnTableData = computed(() => {
  if (!selectedPage.value) return []
  const pageId = selectedPage.value.id
  const buildTree = (items: ColumnItem[], parentId?: number): ColumnItem[] => {
    return items
      .filter(item => item.pageId === pageId && item.parentId === parentId)
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
      .filter(item => item.pageId === pageId && item.parentId === parentId)
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
    allPages.value = [
      { id: 1, name: '网站首页', code: 'home', pageType: 'home', routePath: '/', template: 'portal-home', status: 1, createTime: '2026-01-01 08:00:00' },
      { id: 2, name: '新闻列表页', code: 'news-list', pageType: 'list', routePath: '/news', template: 'article-list', status: 1, createTime: '2026-01-05 09:30:00' },
      { id: 3, name: '产品列表页', code: 'product-list', pageType: 'list', routePath: '/product', template: 'image-list', status: 1, createTime: '2026-01-10 14:00:00' },
      { id: 4, name: '招聘列表页', code: 'careers-list', pageType: 'list', routePath: '/careers', template: 'article-list', status: 0, createTime: '2026-02-01 10:00:00' },
      { id: 5, name: '文章详情页', code: 'article-detail', pageType: 'detail', routePath: '/article/:id', template: 'article-detail', status: 1, createTime: '2026-01-08 12:00:00' },
      { id: 6, name: '产品详情页', code: 'product-detail', pageType: 'detail', routePath: '/product/:id', template: 'page-detail', status: 1, createTime: '2026-01-11 15:00:00' },
      { id: 7, name: '关于我们页', code: 'about-page', pageType: 'detail', routePath: '/about', template: 'page-detail', status: 1, createTime: '2026-01-15 09:00:00' }
    ]

    allColumns.value = [
      { id: 101, name: '新闻中心', code: 'news-center', pageId: 2, routePath: '/news', sort: 1, status: 1, createTime: '2026-01-05 10:00:00' },
      { id: 102, name: '公司新闻', code: 'company-news', pageId: 2, parentId: 101, routePath: '/news/company', sort: 1, status: 1, createTime: '2026-01-06 10:00:00' },
      { id: 103, name: '行业动态', code: 'industry-news', pageId: 2, parentId: 101, routePath: '/news/industry', sort: 2, status: 1, createTime: '2026-01-07 11:00:00' },
      { id: 104, name: '国际新闻', code: 'global-news', pageId: 2, parentId: 101, routePath: '/news/global', sort: 3, status: 1, createTime: '2026-01-09 09:00:00' },
      { id: 201, name: '产品服务', code: 'product-service', pageId: 3, routePath: '/product', sort: 1, status: 1, createTime: '2026-01-10 14:30:00' },
      { id: 202, name: '智能硬件', code: 'smart-hardware', pageId: 3, parentId: 201, routePath: '/product/hardware', sort: 1, status: 1, createTime: '2026-01-12 10:00:00' },
      { id: 203, name: '软件服务', code: 'software-service', pageId: 3, parentId: 201, routePath: '/product/software', sort: 2, status: 1, createTime: '2026-01-13 11:00:00' },
      { id: 301, name: '招贤纳士', code: 'careers', pageId: 4, routePath: '/careers', sort: 1, status: 1, createTime: '2026-02-02 09:00:00' }
    ]

    selectedPage.value = null
  } finally {
    pageLoading.value = false
  }
}

const handleTypeChange = (type: string) => {
  activePageType.value = type
  selectedPage.value = null
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
  }).then(() => {
    ElMessage.success('删除成功')
    allPages.value = allPages.value.filter(p => p.id !== row.id)
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
  setTimeout(() => {
    ElMessage.success(pageForm.id ? '修改成功' : '新增成功')
    pageDialogVisible.value = false
    if (!pageForm.id) {
      allPages.value.push({
        id: Date.now(),
        name: pageForm.name || '',
        code: pageForm.code || '',
        pageType: pageForm.pageType as any,
        routePath: pageForm.routePath || '',
        template: pageForm.template || '',
        description: pageForm.description,
        status: pageForm.status ?? 1,
        createTime: new Date().toLocaleString()
      })
    }
    pageSubmitLoading.value = false
  }, 500)
}

const resetPageForm = () => {
  pageForm.id = undefined
  pageForm.name = ''
  pageForm.code = ''
  pageForm.pageType = activePageType.value as any
  pageForm.routePath = ''
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
    template: row.template,
    description: row.description,
    sort: row.sort,
    status: row.status
  })
  columnDialogVisible.value = true
}

const handleDeleteColumn = (row: ColumnItem) => {
  const hasChildren = allColumns.value.some(c => c.parentId === row.id)
  const msg = hasChildren
    ? `栏目 "${row.name}" 下存在子栏目，确定要一并删除吗？`
    : `确定要删除栏目 "${row.name}" 吗？`
  ElMessageBox.confirm(msg, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success('删除成功')
    allColumns.value = allColumns.value.filter(c => c.id !== row.id && c.parentId !== row.id)
  })
}

const handleColumnStatusChange = async (_row: ColumnItem, val: number) => {
  ElMessage.success(`栏目状态已${val === 1 ? '启用' : '禁用'}`)
}

const handleColumnSubmit = async () => {
  const valid = await columnFormRef.value?.validate().catch(() => false)
  if (!valid) return
  columnSubmitLoading.value = true
  setTimeout(() => {
    ElMessage.success(columnForm.id ? '修改成功' : '新增成功')
    columnDialogVisible.value = false
    if (!columnForm.id) {
      allColumns.value.push({
        id: Date.now(),
        pageId: columnForm.pageId ?? 0,
        parentId: columnForm.parentId,
        name: columnForm.name || '',
        code: columnForm.code || '',
        routePath: columnForm.routePath || '',
        template: columnForm.template || '',
        description: columnForm.description,
        sort: columnForm.sort ?? 0,
        status: columnForm.status ?? 1,
        createTime: new Date().toLocaleString()
      })
    }
    columnSubmitLoading.value = false
  }, 500)
}

const resetColumnForm = () => {
  columnForm.id = undefined
  columnForm.pageId = selectedPage.value?.id
  columnForm.parentId = undefined
  columnForm.name = ''
  columnForm.code = ''
  columnForm.routePath = ''
  columnForm.template = ''
  columnForm.description = ''
  columnForm.sort = 0
  columnForm.status = 1
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.page-container {
  .type-section {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
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
