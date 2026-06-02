<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="栏目名称">
          <el-input v-model="queryForm.name" placeholder="请输入栏目名称" clearable />
        </el-form-item>
        <el-form-item label="页面类型">
          <el-select v-model="queryForm.pageType" placeholder="全部类型" clearable style="width: 140px">
            <el-option label="首页" value="home" />
            <el-option label="列表页" value="list" />
            <el-option label="详情页" value="detail" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryForm.status" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
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
          <span>栏目列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增栏目
          </el-button>
        </div>
      </template>

      <el-table
        :data="tableData"
        v-loading="loading"
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
        <el-table-column prop="pageType" label="页面类型" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="pageTypeTagType(row.pageType)" size="small">
              {{ pageTypeLabel(row.pageType) }}
            </el-tag>
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
              @change="(val: number) => handleStatusChange(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleAddChild(row)">
              <el-icon><CirclePlus /></el-icon>子栏目
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
    </el-card>

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
            <el-form-item label="上级栏目" prop="parentId">
              <el-tree-select
                v-model="form.parentId"
                :data="columnTreeOptions"
                :props="{ label: 'name', value: 'id', children: 'children' }"
                check-strictly
                clearable
                placeholder="请选择上级栏目"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="栏目名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入栏目名称" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="栏目编码" prop="code">
              <el-input v-model="form.code" placeholder="请输入栏目编码" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="排序" prop="sort">
              <el-input-number v-model="form.sort" :min="0" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="页面类型" prop="pageType">
              <el-select v-model="form.pageType" placeholder="请选择页面类型" style="width: 100%">
                <el-option label="首页" value="home">
                  <el-tag size="small" type="success">首页</el-tag>
                  <span style="margin-left: 8px; color: #606266">网站入口首页</span>
                </el-option>
                <el-option label="列表页" value="list">
                  <el-tag size="small" type="primary">列表页</el-tag>
                  <span style="margin-left: 8px; color: #606266">文章列表聚合</span>
                </el-option>
                <el-option label="详情页" value="detail">
                  <el-tag size="small" type="warning">详情页</el-tag>
                  <span style="margin-left: 8px; color: #606266">内容详情展示</span>
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="访问路径" prop="routePath">
              <el-input v-model="form.routePath" placeholder="如 /news" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="绑定模板" prop="template">
          <el-select v-model="form.template" placeholder="请选择页面模板" clearable style="width: 100%">
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
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入栏目描述" />
        </el-form-item>
        <el-form-item label="封面图" prop="cover">
          <el-input v-model="form.cover" placeholder="请输入封面图URL" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  RefreshRight,
  Plus,
  Edit,
  Delete,
  CirclePlus
} from '@element-plus/icons-vue'

interface ColumnItem {
  id: number
  name: string
  code: string
  parentId?: number
  routePath?: string
  pageType: 'home' | 'list' | 'detail'
  template?: string
  description?: string
  cover?: string
  sort: number
  status: number
  createTime: string
  children?: ColumnItem[]
  hasChildren?: boolean
}

const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const formRef = ref()

const queryForm = reactive({
  name: '',
  pageType: undefined as string | undefined,
  status: undefined as number | undefined
})

const form = reactive<Partial<ColumnItem>>({
  id: undefined,
  parentId: undefined,
  name: '',
  code: '',
  routePath: '',
  pageType: 'list',
  template: '',
  description: '',
  cover: '',
  sort: 0,
  status: 1
})

const formRules = {
  name: [{ required: true, message: '请输入栏目名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入栏目编码', trigger: 'blur' }],
  pageType: [{ required: true, message: '请选择页面类型', trigger: 'change' }],
  routePath: [{ required: true, message: '请输入访问路径', trigger: 'blur' }]
}

const tableData = ref<ColumnItem[]>([])

const columnTreeOptions = computed(() => {
  const toOptions = (items: ColumnItem[]): any[] => {
    return items.map(item => ({
      id: item.id,
      name: item.name,
      children: item.children ? toOptions(item.children) : undefined
    }))
  }
  return toOptions(tableData.value)
})

const pageTypeLabel = (type?: string) => {
  const map: Record<string, string> = {
    home: '首页',
    list: '列表页',
    detail: '详情页'
  }
  return map[type || ''] || type
}

const pageTypeTagType = (type?: string): any => {
  const map: Record<string, any> = {
    home: 'success',
    list: 'primary',
    detail: 'warning'
  }
  return map[type || ''] || 'info'
}

const fetchData = async () => {
  loading.value = true
  try {
    const mockData: ColumnItem[] = [
      {
        id: 1,
        name: '网站首页',
        code: 'home',
        routePath: '/',
        pageType: 'home',
        template: 'portal-home',
        sort: 0,
        status: 1,
        createTime: '2026-01-01 08:00:00'
      },
      {
        id: 2,
        name: '新闻中心',
        code: 'news',
        routePath: '/news',
        pageType: 'list',
        template: 'article-list',
        sort: 1,
        status: 1,
        createTime: '2026-01-05 09:30:00',
        children: [
          {
            id: 21,
            name: '公司新闻',
            code: 'company-news',
            parentId: 2,
            routePath: '/news/company',
            pageType: 'list',
            template: 'article-list',
            sort: 1,
            status: 1,
            createTime: '2026-01-06 10:00:00'
          },
          {
            id: 22,
            name: '行业动态',
            code: 'industry-news',
            parentId: 2,
            routePath: '/news/industry',
            pageType: 'list',
            template: 'news-list',
            sort: 2,
            status: 1,
            createTime: '2026-01-07 11:00:00'
          },
          {
            id: 23,
            name: '新闻详情',
            code: 'news-detail',
            parentId: 2,
            routePath: '/news/detail/:id',
            pageType: 'detail',
            template: 'article-detail',
            sort: 3,
            status: 1,
            createTime: '2026-01-08 12:00:00'
          }
        ]
      },
      {
        id: 3,
        name: '产品服务',
        code: 'product',
        routePath: '/product',
        pageType: 'list',
        template: 'image-list',
        sort: 2,
        status: 1,
        createTime: '2026-01-10 14:00:00',
        children: [
          {
            id: 31,
            name: '产品详情',
            code: 'product-detail',
            parentId: 3,
            routePath: '/product/detail/:id',
            pageType: 'detail',
            template: 'page-detail',
            sort: 1,
            status: 1,
            createTime: '2026-01-11 15:00:00'
          }
        ]
      },
      {
        id: 4,
        name: '关于我们',
        code: 'about',
        routePath: '/about',
        pageType: 'detail',
        template: 'page-detail',
        sort: 3,
        status: 1,
        createTime: '2026-01-15 09:00:00'
      },
      {
        id: 5,
        name: '招贤纳士',
        code: 'careers',
        routePath: '/careers',
        pageType: 'list',
        template: 'article-list',
        sort: 4,
        status: 0,
        createTime: '2026-02-01 10:00:00'
      }
    ]

    const filterData = (items: ColumnItem[]): ColumnItem[] => {
      return items
        .filter(item => {
          if (queryForm.name && !item.name.includes(queryForm.name)) return false
          if (queryForm.pageType && item.pageType !== queryForm.pageType) return false
          if (queryForm.status !== undefined && item.status !== queryForm.status) return false
          return true
        })
        .map(item => ({
          ...item,
          children: item.children ? filterData(item.children) : undefined
        }))
        .filter(item => {
          if (!queryForm.name) return true
          const hasVisibleChild = item.children && item.children.length > 0
          return item.name.includes(queryForm.name) || hasVisibleChild
        })
    }

    tableData.value = filterData(mockData)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  fetchData()
}

const resetQuery = () => {
  queryForm.name = ''
  queryForm.pageType = undefined
  queryForm.status = undefined
  fetchData()
}

const handleAdd = () => {
  dialogTitle.value = '新增栏目'
  resetForm()
  dialogVisible.value = true
}

const handleAddChild = (row: ColumnItem) => {
  dialogTitle.value = `新增子栏目 - ${row.name}`
  resetForm()
  form.parentId = row.id
  dialogVisible.value = true
}

const handleEdit = (row: ColumnItem) => {
  dialogTitle.value = '编辑栏目'
  Object.assign(form, {
    id: row.id,
    parentId: row.parentId,
    name: row.name,
    code: row.code,
    routePath: row.routePath,
    pageType: row.pageType,
    template: row.template,
    description: row.description,
    cover: row.cover,
    sort: row.sort,
    status: row.status
  })
  dialogVisible.value = true
}

const handleDelete = (row: ColumnItem) => {
  const hasChildren = row.children && row.children.length > 0
  const msg = hasChildren
    ? `栏目 "${row.name}" 下存在子栏目，确定要一并删除吗？`
    : `确定要删除栏目 "${row.name}" 吗？`
  ElMessageBox.confirm(msg, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleStatusChange = async (_row: ColumnItem, val: number) => {
  ElMessage.success(`栏目状态已${val === 1 ? '启用' : '禁用'}`)
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  setTimeout(() => {
    ElMessage.success(form.id ? '修改成功' : '新增成功')
    dialogVisible.value = false
    fetchData()
    submitLoading.value = false
  }, 500)
}

const resetForm = () => {
  form.id = undefined
  form.parentId = undefined
  form.name = ''
  form.code = ''
  form.routePath = ''
  form.pageType = 'list'
  form.template = ''
  form.description = ''
  form.cover = ''
  form.sort = 0
  form.status = 1
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
}
</style>
