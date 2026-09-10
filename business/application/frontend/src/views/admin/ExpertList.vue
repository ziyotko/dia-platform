<template>
  <div class="page-card">
    <div class="page-toolbar">
      <div style="display:flex;gap:12px;flex-wrap:wrap">
        <el-input v-model="keyword" placeholder="搜索姓名/账号/领域/单位" clearable style="width:240px" @keyup.enter="fetch" @clear="fetch" />
        <el-select v-model="status" placeholder="全部状态" clearable style="width:140px" @change="fetch">
          <el-option label="启用" value="enabled" />
          <el-option label="停用" value="disabled" />
        </el-select>
        <el-button type="primary" @click="fetch">查询</el-button>
      </div>
      <el-button type="primary" @click="openDialog()">新增专家</el-button>
    </div>

    <el-table :data="list" v-loading="loading">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="姓名" width="100" />
      <el-table-column prop="username" label="登录账号" width="120" />
      <el-table-column prop="specialty" label="专业领域" min-width="160" />
      <el-table-column prop="title" label="职称" width="110" />
      <el-table-column prop="organization" label="单位" min-width="160" />
      <el-table-column prop="phone" label="电话" width="130" />
      <el-table-column label="评审任务" width="110">
        <template #default="{ row }">{{ row.scoredCount }}/{{ row.reviewCount }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag></template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" :type="row.status === 1 ? 'warning' : 'success'" @click="toggleStatus(row)">{{ row.status === 1 ? '停用' : '启用' }}</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      style="margin-top:16px;justify-content:flex-end"
      layout="total, prev, pager, next"
      :total="total" :page-size="pageSize" :current-page="page"
      @current-change="onPage"
    />

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑专家' : '新增专家'" width="600px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="姓名"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="登录账号"><el-input v-model="form.username" :disabled="!!form.id" /></el-form-item>
        <el-form-item label="登录密码"><el-input v-model="form.password" type="password" :placeholder="form.id ? '留空不修改' : '请输入初始密码'" /></el-form-item>
        <el-form-item label="专业领域"><el-input v-model="form.specialty" placeholder="如：人工智能、生物医药、新材料" /></el-form-item>
        <el-form-item label="职称"><el-input v-model="form.title" placeholder="如：教授、研究员、高级工程师" /></el-form-item>
        <el-form-item label="单位"><el-input v-model="form.organization" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
        <el-form-item label="专家简介"><el-input v-model="form.bio" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const keyword = ref('')
const status = ref('')
const loading = ref(false)
const dialogVisible = ref(false)
const form = reactive<any>({ id: 0, name: '', username: '', password: '', specialty: '', title: '', organization: '', phone: '', email: '', bio: '' })

async function fetch() {
  loading.value = true
  try {
    const res = await adminApi.getExperts({ page: page.value, pageSize, keyword: keyword.value, status: status.value })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }

function openDialog(row?: any) {
  Object.assign(form, row ? { ...row, password: '' } : { id: 0, name: '', username: '', password: '', specialty: '', title: '', organization: '', phone: '', email: '', bio: '' })
  dialogVisible.value = true
}

async function save() {
  const payload = {
    name: form.name, username: form.username, password: form.password,
    specialty: form.specialty, title: form.title, organization: form.organization,
    phone: form.phone, email: form.email, bio: form.bio,
  }
  if (form.id) await adminApi.updateExpert(form.id, payload)
  else await adminApi.createExpert(payload)
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetch()
}

async function toggleStatus(row: any) {
  await adminApi.setExpertStatus(row.id, { status: row.status === 1 ? 0 : 1 })
  ElMessage.success('状态已更新')
  fetch()
}

async function remove(row: any) {
  await ElMessageBox.confirm(`确认删除专家「${row.name}」？`, '提示', { type: 'warning' })
  await adminApi.deleteExpert(row.id)
  ElMessage.success('删除成功')
  fetch()
}

onMounted(fetch)
</script>
