<template>
  <div class="admin-accounts" v-loading="loading">
    <div class="page-header">
      <h3>管理权限</h3>
      <el-button type="primary" @click="openCreate">新增下属管理员</el-button>
    </div>

    <el-card>
      <el-table :data="accounts" stripe empty-text="暂无下属管理员">
        <el-table-column prop="username" label="用户名" min-width="160" />
        <el-table-column prop="mobile" label="手机号" min-width="150">
          <template #default="{ row }">{{ row.mobile || '-' }}</template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" min-width="200">
          <template #default="{ row }">{{ row.email || '-' }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="170" />
        <el-table-column label="操作" min-width="220" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" @click="openEdit(row)">修改</el-button>
            <el-button text type="warning" @click="resetPassword(row)">重置密码</el-button>
            <el-button text type="danger" @click="deleteAccount(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="editingId ? '修改下属管理员' : '新增下属管理员'"
      width="520px"
      :close-on-click-modal="false"
    >
      <el-form :model="form" label-width="90px">
        <el-form-item label="用户名" required>
          <el-input v-model="form.username" maxlength="20" placeholder="小写字母、数字、- 和 _" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="form.mobile" maxlength="11" placeholder="选填" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" maxlength="64" placeholder="选填" />
        </el-form-item>
        <p v-if="!editingId" class="form-hint">系统将生成随机初始密码，创建成功后仅展示一次。</p>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveAccount">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/admin'

interface AdminAccount {
  id: number
  username: string
  mobile: string
  email: string
  created_at: string
}

const accounts = ref<AdminAccount[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = reactive({ username: '', mobile: '', email: '' })

onMounted(fetchAccounts)

async function fetchAccounts() {
  loading.value = true
  try {
    const res = await adminApi.getAdminAccounts()
    accounts.value = res.data || []
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载下属管理员失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  Object.assign(form, { username: '', mobile: '', email: '' })
  dialogVisible.value = true
}

function openEdit(account: AdminAccount) {
  editingId.value = account.id
  Object.assign(form, { username: account.username, mobile: account.mobile, email: account.email })
  dialogVisible.value = true
}

async function saveAccount() {
  const username = form.username.trim()
  if (!username) {
    ElMessage.warning('请输入用户名')
    return
  }
  saving.value = true
  try {
    const data = { username, mobile: form.mobile.trim(), email: form.email.trim() }
    if (editingId.value) {
      await adminApi.updateAdminAccount(editingId.value, data)
      ElMessage.success('管理员信息已更新')
    } else {
      const res = await adminApi.createAdminAccount(data)
      await ElMessageBox.alert(`初始密码：${res.data.password}`, '下属管理员创建成功', {
        confirmButtonText: '我已保存',
      })
    }
    dialogVisible.value = false
    await fetchAccounts()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存管理员失败')
  } finally {
    saving.value = false
  }
}

async function resetPassword(account: AdminAccount) {
  try {
    await ElMessageBox.confirm(`确定重置管理员“${account.username}”的密码吗？`, '重置密码', {
      type: 'warning',
    })
  } catch {
    return
  }
  const res = await adminApi.resetAdminAccountPassword(account.id)
  await ElMessageBox.alert(`新密码：${res.data.password}`, '密码已重置', {
    confirmButtonText: '我已保存',
  })
}

async function deleteAccount(account: AdminAccount) {
  try {
    await ElMessageBox.confirm(
      `删除管理员“${account.username}”后将无法恢复，确定继续吗？`,
      '删除管理员',
      { type: 'warning' }
    )
  } catch {
    return
  }
  await adminApi.deleteAdminAccount(account.id)
  ElMessage.success('下属管理员已删除')
  await fetchAccounts()
}
</script>

<style scoped lang="scss">
.admin-accounts {
  width: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  h3 {
    font-size: 22px;
    font-weight: 600;
    color: var(--app-text-primary);
  }
}

.form-hint {
  margin: 0 0 0 90px;
  color: var(--app-text-secondary);
  font-size: 13px;
}
</style>
