<template>
  <div class="page-container">
    <el-card shadow="hover" class="page-card">
      <template #header>
        <div class="card-header">
          <div class="header-title">
            <span class="title-badge">
              <el-icon :size="16"><Tickets /></el-icon>
            </span>
            <span class="title-text">模板列表</span>
            <el-tag v-if="templateTableData.length" size="small" type="info" effect="plain" round>
              共 {{ templateTableData.length }} 个
            </el-tag>
          </div>
          <el-radio-group
            v-model="activePageType"
            size="small"
            class="type-switch"
            @change="handleTypeChange"
          >
            <el-radio-button
              v-for="type in pageTypeList"
              :key="type.value"
              :value="type.value"
              :title="type.description"
            >
              {{ type.label }}
            </el-radio-button>
          </el-radio-group>
        </div>
      </template>

      <el-table
        :data="templateTableData"
        v-loading="templateLoading"
        highlight-current-row
        border
        stripe
        @current-change="handleTemplateSelect"
      >
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="name" label="模板名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="description" label="模板描述" min-width="180" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" />
      </el-table>

      <el-empty v-if="!templateTableData.length && !templateLoading" description="该页面类型下暂无启用的模板" />
    </el-card>

    <el-card v-if="selectedTemplate" shadow="hover" class="column-card">
      <template #header>
        <div class="card-header">
          <div class="breadcrumb-title">
            <span class="page-name">{{ selectedTemplate.name }}</span>
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
        <el-table-column prop="name" label="栏目名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="code" label="栏目编码" min-width="160" />
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
            <el-tag v-if="row.workflow?.name" size="small" type="warning">{{ row.workflow.name }}</el-tag>
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
        <el-table-column prop="createdAt" label="创建时间" width="170" />
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

      <el-empty v-if="!columnTableData.length && !columnLoading" description="该模板下暂无栏目，请添加" />
    </el-card>

    <el-empty v-else description="请先在上方模板列表中选择模板" class="select-tip" />

    <el-dialog
      v-model="columnDialogVisible"
      :title="columnDialogTitle"
      width="600px"
      destroy-on-close
    >
      <el-form ref="columnFormRef" :model="columnForm" :rules="columnFormRules" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="所属模板">
              <el-input :model-value="selectedTemplate?.name || ''" disabled placeholder="请先选择模板" />
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
import { Plus, Edit, Delete, CirclePlus, ArrowRight, Tickets } from '@element-plus/icons-vue'
import {
  getColumns,
  createColumn,
  updateColumn,
  deleteColumn
} from '@/api/column'
import { getAllTemplates } from '@/api/template'
import { getAllWorkflows } from '@/api/workflow'
import pinyin from 'js-pinyin'

interface ColumnItem {
  id: number
  name: string
  code: string
  templateId: number
  parentId?: number
  routePath?: string
  description?: string
  sort: number
  status: number
  displayType: number
  workflowId?: number
  workflow?: { id: number; name: string }
  createdAt: string
  children?: ColumnItem[]
}

const activePageType = ref<string>('home')
const selectedTemplate = ref<any | null>(null)

const toPinyinCode = (str: string): string => {
  if (!str) return ''
  const py = pinyin.getFullChars(str)
  return py.toLowerCase().replace(/[^a-z0-9]/g, '')
}

const allColumns = ref<ColumnItem[]>([])

const templateLoading = ref(false)

const columnLoading = ref(false)
const columnDialogVisible = ref(false)
const columnDialogTitle = ref('')
const columnSubmitLoading = ref(false)
const columnFormRef = ref()

const templateList = ref<any[]>([])

const displayTypeMap: Record<number, string> = {
  1: '轮播展示',
  2: '列表展示',
  3: '图片展示',
  4: '广告展示',
  5: '友链展示',
  6: '报刊展示',
  7: '数据展示',
  8: '视频展示',
  9: '其他展示'
}

const displayTypeOptions = [
  { label: '轮播展示', value: 1 },
  { label: '列表展示', value: 2 },
  { label: '图片展示', value: 3 },
  { label: '广告展示', value: 4 },
  { label: '友链展示', value: 5 },
  { label: '报刊展示', value: 6 },
    { label: '数据展示', value: 7 },
      { label: '视频展示', value: 8 },
  { label: '其他展示', value: 9 }
]

const columnForm = reactive<Partial<ColumnItem>>({
  id: undefined,
  templateId: undefined,
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
    const res: any = await getAllWorkflows()
    workflowList.value = (res.data.list || []).filter((w: any) => w.status === 1)
  } catch (error) {
    ElMessage.error('获取流程列表失败')
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
    description: '网站门户页'
  },
  {
    value: 'column',
    label: '栏目页',
    description: '文章栏目聚合页'
  },
  {
    value: 'detail',
    label: '详情页',
    description: '内容详情展示页'
  },
  {
    value: 'special',
    label: '专题页',
    description: '专题活动展示页'
  }
])

// 模板列表：仅展示当前页面类型、且状态为启用的模板（模板 type 与页面类型同值：home/column/detail/special）
const templateTableData = computed(() => {
  return templateList.value.filter((t: any) => t.type === activePageType.value && t.status === 1)
})

// 栏目直接挂在模板下（column.template_id，页面层已合并进模板）
const currentTemplateId = computed(() => Number(selectedTemplate.value?.id || 0))

const isRootColumn = (item: ColumnItem) => !item.parentId || item.parentId === 0

const columnTableData = computed(() => {
  const templateId = currentTemplateId.value
  if (!templateId) return []
  const buildTree = (items: ColumnItem[], parentId?: number): ColumnItem[] => {
    return items
      .filter(item => {
        if (item.templateId !== templateId) return false
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

// 编辑中的栏目自身 + 其全部后代 ID（作为上级候选时需排除）
const excludedParentIds = computed(() => {
  const blocked = new Set<number>()
  const rootId = Number(columnForm.id || 0)
  if (!rootId) return blocked
  blocked.add(rootId)
  const collect = (parentId: number) => {
    allColumns.value.forEach((item: ColumnItem) => {
      if (item.parentId === parentId && !blocked.has(item.id)) {
        blocked.add(item.id)
        collect(item.id)
      }
    })
  }
  collect(rootId)
  return blocked
})

const columnTreeOptions = computed(() => {
  const templateId = currentTemplateId.value
  if (!templateId) return []
  const blocked = excludedParentIds.value
  const buildOptions = (items: ColumnItem[], parentId?: number): any[] => {
    return items
      .filter(item => {
        if (item.templateId !== templateId) return false
        // 编辑时剔除自身及其全部下级：否则可以把栏目挂到自己的后代下形成环，
        // 该枝栏目会从栏目树/投放树中永久消失（后端已拦截，这里提前避免误选）
        if (blocked.has(item.id)) return false
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

// 读取当前页面类型的模板列表
const fetchTemplates = async () => {
  templateLoading.value = true
  try {
    const res: any = await getAllTemplates(activePageType.value)
    templateList.value = res.data.list || []
  } catch (error) {
    ElMessage.error('获取模板列表失败')
  } finally {
    templateLoading.value = false
  }
}

const fetchColumns = async () => {
  const templateId = currentTemplateId.value
  if (!templateId) {
    allColumns.value = []
    return
  }
  columnLoading.value = true
  try {
    const res: any = await getColumns({ templateId })
    allColumns.value = res.data || []
  } catch (error) {
    ElMessage.error('获取栏目列表失败')
  } finally {
    columnLoading.value = false
  }
}

watch(() => columnForm.name, (val) => {
  // 仅「新增」时按名称自动生成编码/访问路径：编辑时 handleEditColumn 会先写入原有 code/routePath，
  // 本回调随后触发并把它们覆盖为名称拼音，保存后原编码/访问路径丢失（静态化路径也随之变化）。
  if (columnForm.id) return
  if (val) {
    columnForm.code = toPinyinCode(val)
    columnForm.routePath = `/${toPinyinCode(val)}`
  }
})

const handleTypeChange = () => {
  // 注意：选项值已由 el-radio-group 的 v-model 写入 activePageType，这里只负责重载数据
  selectedTemplate.value = null
  allColumns.value = []
  fetchTemplates()
}

// 点击模板行：下方面目列表直接切换到该模板下的栏目
const handleTemplateSelect = (row: any) => {
  selectedTemplate.value = row || null
  allColumns.value = []
  fetchColumns()
}

const handleAddColumn = () => {
  if (!currentTemplateId.value) return
  columnDialogTitle.value = '新增栏目'
  resetColumnForm()
  columnDialogVisible.value = true
}

const handleAddChildColumn = (row: ColumnItem) => {
  columnDialogTitle.value = `新增子栏目 - ${row.name}`
  resetColumnForm()
  columnForm.templateId = row.templateId
  columnForm.parentId = row.id
  columnDialogVisible.value = true
}

const handleEditColumn = (row: ColumnItem) => {
  columnDialogTitle.value = '编辑栏目'
  Object.assign(columnForm, {
    id: row.id,
    templateId: row.templateId,
    parentId: row.parentId || undefined,
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
  // 后端在存在子栏目时拒绝删除（不做级联删除），这里提前拦下，避免“一并删除”的误导文案
  const hasChildren = allColumns.value.some(c => c.parentId === row.id)
  if (hasChildren) {
    ElMessage.warning(`栏目 "${row.name}" 下存在子栏目，请先删除或调整子栏目后再删除`)
    return
  }
  const msg = `确定要删除栏目 "${row.name}" 吗？`
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
      templateId: row.templateId,
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
      templateId: columnForm.templateId ?? 0,
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
  columnForm.templateId = currentTemplateId.value
  columnForm.parentId = undefined
  columnForm.name = ''
  columnForm.code = ''
  columnForm.routePath = ''
  columnForm.description = ''
  columnForm.sort = 0
  columnForm.status = 1
  columnForm.displayType = 2
  columnForm.workflowId = undefined
}

onMounted(() => {
  fetchTemplates()
  fetchWorkflows()
})
</script>

<style scoped lang="scss">
.page-container {
  /* 两张卡片的头部统一样式：紧凑内边距 + 极浅渐变，标题区更清爽 */
  .page-card,
  .column-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    :deep(.el-card__header) {
      padding: 12px 20px;
      border-bottom: 1px solid #eef4fd;
      background: linear-gradient(90deg, #f8fbff 0%, #ffffff 70%);
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 12px;
      font-weight: 600;
      color: #2c3e50;
    }
  }

  .page-card {
    margin-bottom: 20px;

    /* 图标圆底徽标 + 标题 + 启用模板数标签 */
    .header-title {
      display: flex;
      align-items: center;
      gap: 10px;
      min-width: 0;

      /* 图标圆底徽标 */
      .title-badge {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
        width: 28px;
        height: 28px;
        border-radius: 50%;
        color: #002fa7;
        background: linear-gradient(135deg, #e6f2ff 0%, #d3e4fb 100%);
        box-shadow: 0 2px 6px rgba(0, 47, 167, 0.18);
      }

      .title-text {
        font-size: 16px;
        letter-spacing: 0.5px;
      }
    }

    /* 胶囊 pill 分段控件：浅色底槽 + 选中白色胶囊 */
    .type-switch {
      flex-shrink: 0;
      gap: 2px;
      padding: 3px;
      border-radius: 999px;
      background: #eef3fb;

      :deep(.el-radio-button__inner) {
        padding: 5px 16px;
        border-radius: 999px;
        outline: none;
        background: transparent;
        color: #5a6b8c;
        font-weight: 500;
        box-shadow: none;
        transition: all 0.2s ease;
      }

      :deep(.el-radio-button__inner:hover) {
        color: #002fa7;
      }

      /* 覆盖 EP 默认的主色实心块，改为白色胶囊 + 品牌蓝文字 */
      :deep(.el-radio-button.is-active .el-radio-button__original-radio:not(:disabled) + .el-radio-button__inner) {
        color: #002fa7;
        background-color: #fff;
        box-shadow: 0 2px 8px rgba(0, 47, 167, 0.16);
      }
    }
  }

  .column-card {
    .card-header {
      .breadcrumb-title {
        display: flex;
        align-items: center;
        gap: 8px;

        .page-name {
          font-size: 16px;
          font-weight: 600;
          color: #002fa7;
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
