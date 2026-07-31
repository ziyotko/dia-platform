<template>
  <div class="admin-sysconfig" v-loading="loading">
    <div class="page-header">
      <h3>系统管理</h3>
      <el-button type="primary" @click="openCreate"><el-icon><Plus /></el-icon>&nbsp;新增配置</el-button>
    </div>
    <el-card>
      <el-table :data="list" stripe empty-text="暂无配置项">
        <el-table-column prop="key" label="配置项" width="220">
          <template #default="{ row }">
            <code class="key-text">{{ row.key }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="value" label="配置值" min-width="240">
          <template #default="{ row }">
            <span class="value-text">{{ row.value }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="说明" min-width="160" />
        <el-table-column prop="updated_at" label="更新时间" width="160">
          <template #default="{ row }">{{ row.updated_at?.replace('T', ' ').slice(0, 16) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button text size="small" type="primary" @click="editRow(row)">编辑</el-button>
            <el-button text size="small" type="danger" @click="delRow(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showDialog" :title="editingId ? '编辑配置' : '新增配置'" width="560px">
      <el-form :model="cfgForm" size="large" label-width="90px">
        <el-form-item label="配置项" required>
          <el-input v-model="cfgForm.key" :disabled="!!editingId" placeholder="例如：site_name" />
        </el-form-item>
        <el-form-item label="配置值">
          <el-input v-model="cfgForm.value" type="textarea" :rows="3" placeholder="请输入配置值" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="cfgForm.description" placeholder="请输入配置说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveRow">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

const list = ref<any[]>([])
const loading = ref(true)
const saving = ref(false)
const showDialog = ref(false)
const editingId = ref<number | null>(null)
const cfgForm = reactive({ key: '', value: '', description: '' })

onMounted(fetchData)

async function fetchData() {
  loading.value = true
  try {
    const r = await adminApi.getSystemConfigs()
    list.value = r.data || []
  } catch {} finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  Object.assign(cfgForm, { key: '', value: '', description: '' })
  showDialog.value = true
}

function editRow(row: any) {
  editingId.value = row.id
  Object.assign(cfgForm, { key: row.key, value: row.value, description: row.description })
  showDialog.value = true
}

async function saveRow() {
  if (!cfgForm.key.trim()) {
    ElMessage.warning('请输入配置项名称')
    return
  }
  saving.value = true
  try {
    const data = { key: cfgForm.key.trim(), value: cfgForm.value, description: cfgForm.description }
    if (editingId.value) {
      await adminApi.updateSystemConfig(editingId.value, data)
    } else {
      await adminApi.createSystemConfig(data)
    }
    ElMessage.success('保存成功')
    showDialog.value = false
    fetchData()
  } catch {} finally {
    saving.value = false
  }
}

async function delRow(row: any) {
  try {
    await ElMessageBox.confirm(`确认删除配置项「${row.key}」？`, '警告', { type: 'warning' })
    await adminApi.deleteSystemConfig(row.id)
    ElMessage.success('已删除')
    fetchData()
  } catch {}
}
</script>

<style scoped lang="scss">
.admin-sysconfig {
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
    color: #1a1a2e;
    margin: 0;
  }
}
.el-card {
  border-radius: 10px;
}
.key-text {
  background: #f0f7ff;
  color: #3b82f6;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 13px;
}
.value-text {
  white-space: pre-wrap;
  word-break: break-all;
  color: #374151;
}
</style>
