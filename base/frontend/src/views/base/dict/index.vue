<template>
  <div class="dict-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>数据字典</span>
          <el-button v-if="can('base:dict:create')" type="primary" @click="handleAdd">新增字典</el-button>
        </div>
      </template>

      <el-form :inline="true" :model="query" class="search-form">
        <el-form-item label="字典编码">
          <el-input v-model="query.code" placeholder="请输入" clearable />
        </el-form-item>
        <el-form-item label="字典名称">
          <el-input v-model="query.name" placeholder="请输入" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchData">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="code" label="字典编码" min-width="150" />
        <el-table-column prop="name" label="字典名称" min-width="150" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button v-if="can('base:dict:update')" link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button v-if="can('base:dict:save-items')" link type="primary" @click="handleEditItems(row)">字典项</el-button>
            <el-button v-if="can('base:dict:delete')" link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.size"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑字典' : '新增字典'" width="600px">
      <el-form :model="form" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="字典编码" prop="code">
          <el-input v-model="form.code" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="字典名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="itemDialogVisible" title="字典项管理" width="700px">
      <el-button type="primary" @click="handleAddItem" style="margin-bottom: 16px">新增项</el-button>
      <el-table :data="currentItems" border>
        <el-table-column prop="label" label="显示名">
          <template #default="{ row }">
            <el-input v-model="row.label" />
          </template>
        </el-table-column>
        <el-table-column prop="value" label="字典值">
          <template #default="{ row }">
            <el-input v-model="row.value" />
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="100">
          <template #default="{ row }">
            <el-input-number v-model="row.sort" :min="0" controls-position="right" style="width: 80px" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80">
          <template #default="{ $index }">
            <el-button link type="danger" @click="handleRemoveItem($index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="itemDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveItems" v-if="can('base:dict:save-items')">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDictList, createDict, updateDict, deleteDict, saveDictItems, getDictDetail } from '@/api/dict'
import type { Dict, DictItem, DictQuery } from '@/api/dict'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
// 按钮级权限：与后端 base:dict:* 权限点对齐
const can = (code: string) => userStore.can(code)

const loading = ref(false)
const tableData = ref<Dict[]>([])
const total = ref(0)
const dialogVisible = ref(false)
const itemDialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref()
const currentDictId = ref<number>(0)
const currentItems = ref<DictItem[]>([])

const defaultQuery: DictQuery = { page: 1, size: 10, code: '', name: '', status: undefined }
const query = reactive<DictQuery>({ ...defaultQuery })

const form = reactive<Dict>({ code: '', name: '', description: '', status: 1 })
const formRules = {
  code: [{ required: true, message: '请输入字典编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入字典名称', trigger: 'blur' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getDictList(query)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  Object.assign(query, defaultQuery)
  fetchData()
}

const handleAdd = () => {
  isEdit.value = false
  Object.assign(form, { code: '', name: '', description: '', status: 1 })
  dialogVisible.value = true
}

const handleEdit = (row: Dict) => {
  isEdit.value = true
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  try {
    if (isEdit.value && form.id) {
      await updateDict(form.id, form)
    } else {
      await createDict(form)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    fetchData()
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const handleDelete = (row: Dict) => {
  ElMessageBox.confirm('确认删除该字典？', '提示', { type: 'warning' }).then(async () => {
    await deleteDict(row.id as number)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleEditItems = async (row: Dict) => {
  currentDictId.value = row.id as number
  try {
    const res: any = await getDictDetail(row.id as number)
    currentItems.value = res.data.items || []
    itemDialogVisible.value = true
  } catch (error) {
    ElMessage.error('加载字典项失败')
  }
}

const handleAddItem = () => {
  currentItems.value.push({ label: '', value: '', sort: currentItems.value.length, status: 1 })
}

const handleRemoveItem = (index: number) => {
  currentItems.value.splice(index, 1)
}

const handleSaveItems = async () => {
  try {
    await saveDictItems(currentDictId.value, currentItems.value)
    ElMessage.success('保存成功')
    itemDialogVisible.value = false
    fetchData()
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

fetchData()
</script>

<style scoped lang="scss">
.dict-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .search-form {
    margin-bottom: 20px;
  }
  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
