<template>
  <div class="page-card">
    <div class="page-toolbar">
      <el-input v-model="keyword" placeholder="搜索用户名/姓名" clearable style="width:220px" @keyup.enter="fetch" @clear="fetch" />
      <el-button type="primary" @click="openDialog()">新增账号</el-button>
    </div>

    <el-table :data="list" v-loading="loading">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="username" label="用户名" width="130" />
      <el-table-column prop="realName" label="姓名" width="120" />
      <el-table-column label="角色" width="130">
        <template #default="{ row }">
          <el-tag :type="roleType(row.roleCode)">{{ roleName(row.roleCode) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="phone" label="电话" width="130" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="openDialog(row)">编辑</el-button>
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

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑账号' : '新增账号'" width="520px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="用户名"><el-input v-model="form.username" :disabled="!!form.id" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="form.password" type="password" :placeholder="form.id ? '留空不修改' : '请输入密码'" /></el-form-item>
        <el-form-item label="姓名"><el-input v-model="form.realName" /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.roleCode" style="width:100%">
            <el-option v-for="r in roleOptions" :key="r.code" :label="r.name" :value="r.code" />
          </el-select>
          <div class="form-hint">评审人账号请通过「专家库」新增，以保证账号与专家档案一一对应</div>
        </el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
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
const roles = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const keyword = ref('')
const loading = ref(false)
const dialogVisible = ref(false)
const form = reactive<any>({ id: 0, username: '', password: '', realName: '', roleCode: '', phone: '', email: '' })

function roleName(code: string) {
  const map: Record<string, string> = { super_admin: '超级管理员', manager: '管理人', reviewer: '评审人' }
  return map[code] || code
}
function roleType(code: string) {
  const map: Record<string, string> = { super_admin: 'danger', manager: 'primary', reviewer: 'warning' }
  return (map[code] || 'info') as any
}

// 评审人不在账号管理里创建（列表上仍显示既有评审人账号）；
// 编辑已有评审人时保留其角色选项，否则保存会把角色清空。
const roleOptions = computed(() => {
  const list = roles.value.filter((r: any) => r.code !== 'reviewer')
  if (form.id && form.roleCode === 'reviewer') {
    return [...list, { code: 'reviewer', name: '评审人' }]
  }
  return list
})

async function fetch() {
  loading.value = true
  try {
    const res = await adminApi.getAdmins({ page: page.value, pageSize, keyword: keyword.value })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function onPage(p: number) { page.value = p; fetch() }

function openDialog(row?: any) {
  Object.assign(form, row ? { ...row, password: '' } : { id: 0, username: '', password: '', realName: '', roleCode: 'manager', phone: '', email: '' })
  dialogVisible.value = true
}

async function save() {
  const payload = { password: form.password, realName: form.realName, roleCode: form.roleCode, phone: form.phone, email: form.email }
  if (form.id) await adminApi.updateAdmin(form.id, payload)
  else await adminApi.createAdmin({ ...payload, username: form.username })
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetch()
}

async function remove(row: any) {
  await ElMessageBox.confirm('确认删除该账号？', '提示', { type: 'warning' })
  await adminApi.deleteAdmin(row.id)
  ElMessage.success('删除成功')
  fetch()
}

onMounted(async () => {
  fetch()
  const res = await adminApi.getRoles()
  roles.value = res.data
})
</script>
