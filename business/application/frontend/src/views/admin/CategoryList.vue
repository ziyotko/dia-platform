<template>
  <div class="page-card">
    <div class="page-toolbar">
      <el-input v-model="keyword" placeholder="搜索名称" clearable style="width:220px" />
      <el-button type="primary" @click="openDialog()">新增类别</el-button>
    </div>

    <el-table :data="filtered" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="类别名称" min-width="160" />
      <el-table-column prop="description" label="说明" min-width="240" />
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑类别' : '新增类别'" width="480px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="form.description" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([])
const keyword = ref('')
const loading = ref(false)
const dialogVisible = ref(false)
const form = reactive<any>({ id: 0, name: '', description: '', sort: 0 })

const filtered = computed(() => list.value.filter((c) => !keyword.value || c.name.includes(keyword.value)))

async function fetch() {
  loading.value = true
  try {
    const res = await adminApi.getCategories()
    list.value = res.data
  } finally {
    loading.value = false
  }
}

function openDialog(row?: any) {
  Object.assign(form, row ? { ...row } : { id: 0, name: '', description: '', sort: 0 })
  dialogVisible.value = true
}

async function save() {
  if (form.id) await adminApi.updateCategory(form.id, { name: form.name, description: form.description, sort: form.sort })
  else await adminApi.createCategory({ name: form.name, description: form.description, sort: form.sort })
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetch()
}

async function remove(row: any) {
  await ElMessageBox.confirm('确认删除该类别？', '提示', { type: 'warning' })
  await adminApi.deleteCategory(row.id)
  ElMessage.success('删除成功')
  fetch()
}

onMounted(fetch)
</script>
