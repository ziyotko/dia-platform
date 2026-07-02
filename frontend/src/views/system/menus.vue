<template>
  <div class="page-container">
    <el-card shadow="hover" class="table-card">
      <template #header>
        <div class="card-header">
          <span>菜单管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增菜单
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
      >
        <el-table-column prop="name" label="菜单名称" min-width="160">
          <template #default="{ row }">
            <el-icon v-if="row.icon" style="margin-right: 6px;">
              <component :is="row.icon" />
            </el-icon>
            <span>{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路由路径" min-width="140" />
        <el-table-column prop="component" label="组件路径" min-width="180" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="type" label="类型" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.type === 'directory' ? 'primary' : 'success'" size="small">
              {{ row.type === 'directory' ? '目录' : '菜单' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '显示' : '隐藏' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleAddChild(row)">
              <el-icon><CirclePlus /></el-icon>子菜单
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
      width="500px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="90px"
      >
        <el-form-item label="上级菜单">
          <el-tree-select
            v-model="form.parentId"
            :data="menuTreeData"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="请选择上级菜单"
            clearable
            check-strictly
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="菜单类型" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio value="directory">目录</el-radio>
            <el-radio value="menu">菜单</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="菜单名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入菜单名称" />
        </el-form-item>
        <el-form-item label="路由路径" prop="path">
          <el-input v-model="form.path" placeholder="请输入路由路径" />
        </el-form-item>
        <el-form-item label="组件路径" v-if="form.type === 'menu'">
          <el-input v-model="form.component" placeholder="请输入组件路径" />
        </el-form-item>
        <el-form-item label="菜单图标">
          <div class="icon-select-row">
            <el-input v-model="form.icon" placeholder="请选择图标" readonly style="flex: 1">
              <template #prefix>
                <el-icon v-if="form.icon"><component :is="form.icon" /></el-icon>
              </template>
            </el-input>
            <el-button v-if="form.icon" @click="clearIcon">
              <el-icon><Close /></el-icon>
            </el-button>
            <el-button @click="iconPickerVisible = true">
              <el-icon><Search /></el-icon>
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="显示排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="显示状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">显示</el-radio>
            <el-radio :value="0">隐藏</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="iconPickerVisible"
      title="选择图标"
      width="640px"
      append-to-body
      destroy-on-close
    >
      <el-input
        v-model="iconSearch"
        placeholder="搜索图标名称"
        clearable
        style="margin-bottom: 16px"
      />
      <el-scrollbar max-height="400px">
        <div class="icon-grid">
          <div
            v-for="name in filteredIcons"
            :key="name"
            class="icon-item"
            :class="{ active: form.icon === name }"
            @click="selectIcon(name)"
          >
            <el-icon size="20"><component :is="name" /></el-icon>
            <span class="icon-name">{{ name }}</span>
          </div>
        </div>
      </el-scrollbar>
      <template #footer>
        <el-button @click="clearIcon">清空图标</el-button>
        <el-button @click="iconPickerVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete, CirclePlus, Search, Close } from '@element-plus/icons-vue'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { getMenuList, createMenu, updateMenu, deleteMenu, type MenuItem, type MenuForm } from '@/api/menus'

const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const formRef = ref()
const tableData = ref<MenuItem[]>([])
const iconPickerVisible = ref(false)
const iconSearch = ref('')

const iconNames = Object.keys(ElementPlusIconsVue)

const filteredIcons = computed(() => {
  if (!iconSearch.value) return iconNames
  const keyword = iconSearch.value.toLowerCase()
  return iconNames.filter((name) => name.toLowerCase().includes(keyword))
})

const selectIcon = (name: string) => {
  form.icon = name
  iconPickerVisible.value = false
}

const clearIcon = () => {
  form.icon = ''
  iconPickerVisible.value = false
}

const form = reactive<MenuForm>({
  id: undefined,
  parentId: undefined,
  name: '',
  path: '',
  component: '',
  icon: '',
  type: 'directory',
  sort: 0,
  status: 1
})

const formRules = {
  name: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
  path: [{ required: true, message: '请输入路由路径', trigger: 'blur' }],
  sort: [{ required: true, message: '请输入排序', trigger: 'blur' }]
}

const menuTreeData = ref<MenuItem[]>([])

const fetchMenus = async () => {
  loading.value = true
  try {
    const res: any = await getMenuList()
    tableData.value = res.data || []
    menuTreeData.value = [{ id: 0, parentId: 0, name: '顶级菜单', path: '', component: '', icon: '', type: 'directory', sort: 0, status: 1, children: [] }, ...tableData.value]
  } catch (error) {
    ElMessage.error('获取菜单列表失败')
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  dialogTitle.value = '新增菜单'
  resetForm()
  dialogVisible.value = true
}

const handleAddChild = (row: MenuItem) => {
  dialogTitle.value = '新增子菜单'
  resetForm()
  form.parentId = row.id
  dialogVisible.value = true
}

const handleEdit = (row: MenuItem) => {
  dialogTitle.value = '编辑菜单'
  Object.assign(form, {
    id: row.id,
    parentId: row.parentId === 0 ? undefined : row.parentId,
    name: row.name,
    path: row.path,
    component: row.component,
    icon: row.icon,
    type: row.type,
    sort: row.sort,
    status: row.status
  })
  dialogVisible.value = true
}

const handleDelete = (row: MenuItem) => {
  ElMessageBox.confirm(`确定要删除菜单 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deleteMenu(row.id)
      ElMessage.success('删除成功')
      fetchMenus()
    } catch (error) {
      ElMessage.error('删除菜单失败')
    }
  })
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    const data = { ...form }
    if (!data.parentId) data.parentId = 0
    if (data.id) {
      await updateMenu(data.id, data)
      ElMessage.success('修改成功')
    } else {
      await createMenu(data)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchMenus()
  } catch (error) {
    ElMessage.error('提交失败')
  } finally {
    submitLoading.value = false
  }
}

const resetForm = () => {
  form.id = undefined
  form.parentId = undefined
  form.name = ''
  form.path = ''
  form.component = ''
  form.icon = ''
  form.type = 'directory'
  form.sort = 0
  form.status = 1
}

onMounted(() => {
  fetchMenus()
})
</script>

<style scoped lang="scss">
.page-container {
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
}

.icon-select-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.icon-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 8px;

  .icon-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 10px 4px;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
    border: 1px solid transparent;

    &:hover {
      background: #f5f9ff;
      border-color: #d9ecff;
    }

    &.active {
      background: rgba(64, 158, 255, 0.12);
      border-color: #409eff;
    }

    .icon-name {
      margin-top: 4px;
      font-size: 11px;
      color: #606266;
      max-width: 100%;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
}
</style>
