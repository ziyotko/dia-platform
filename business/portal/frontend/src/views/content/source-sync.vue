<template>
  <div class="page-container">
    <el-alert
      class="source-tip"
      type="info"
      show-icon
      :closable="false"
      title="公众号文章会以草稿形式导入图文管理；重复同步按微信文章 ID 自动跳过，不会重复建稿。"
    />

    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="账号名称">
          <el-input v-model="queryForm.name" placeholder="请输入公众号名称" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryForm.status" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch"><el-icon><Search /></el-icon>查询</el-button>
          <el-button @click="resetQuery"><el-icon><RefreshRight /></el-icon>重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="hover" class="table-card">
      <template #header>
        <div class="card-header">
          <div>
            <span>内容源账号</span>
            <span class="header-subtitle">当前接入微信公众号，后续可继续扩展其他来源</span>
          </div>
          <el-button type="primary" @click="handleAdd"><el-icon><Plus /></el-icon>新增公众号</el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column prop="name" label="账号名称" min-width="170" show-overflow-tooltip />
        <el-table-column label="账号类型" width="110" align="center">
          <template #default="{ row }">{{ accountTypeText(row.accountType) }}</template>
        </el-table-column>
        <el-table-column prop="appId" label="AppID" min-width="190" show-overflow-tooltip />
        <el-table-column prop="secretEnv" label="密钥环境变量" min-width="220" show-overflow-tooltip />
        <el-table-column label="凭据" width="105" align="center">
          <template #default="{ row }">
            <el-tag :type="row.secretConfigured ? 'success' : 'danger'" effect="light">
              {{ row.secretConfigured ? '已配置' : '未配置' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="采集范围" width="130" align="center">
          <template #default="{ row }">{{ syncModeText(row.syncMode) }}</template>
        </el-table-column>
        <el-table-column label="默认分类" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ categoryName(row.defaultCategoryId) }}</template>
        </el-table-column>
        <el-table-column label="最近同步" min-width="210">
          <template #default="{ row }">
            <div class="sync-status-cell">
              <div>
                <el-tag v-if="row.lastSyncStatus" :type="syncStatusType(row.lastSyncStatus)" size="small">
                  {{ syncStatusText(row.lastSyncStatus) }}
                </el-tag>
                <span v-else class="muted">尚未同步</span>
              </div>
              <span v-if="row.lastSyncAt" class="muted">{{ formatTime(row.lastSyncAt) }}</span>
              <span v-if="row.lastSyncMessage" class="last-message">{{ row.lastSyncMessage }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="85" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" :loading="testingIds.includes(row.id)" @click="handleTest(row)">测试</el-button>
            <el-button
              link
              type="success"
              :disabled="row.status !== 1 || !row.secretConfigured"
              :loading="syncingIds.includes(row.id)"
              @click="handleSync(row)"
            >同步</el-button>
            <el-button link type="primary" @click="openLogs(row)">记录</el-button>
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="queryForm.page"
          v-model:page-size="queryForm.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchData"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="650px" destroy-on-close :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="账号名称" prop="name">
          <el-input v-model="form.name" placeholder="例如：中国环保机械行业协会" />
        </el-form-item>
        <el-form-item label="来源类型">
          <el-select v-model="form.provider" disabled style="width: 100%">
            <el-option label="微信公众号" value="wechat_official" />
          </el-select>
        </el-form-item>
        <el-form-item label="公众号类型" prop="accountType">
          <el-radio-group v-model="form.accountType">
            <el-radio value="service">服务号</el-radio>
            <el-radio value="subscription">订阅号</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="AppID" prop="appId">
          <el-input v-model="form.appId" placeholder="请输入公众号 AppID" />
        </el-form-item>
        <el-form-item label="密钥环境变量" prop="secretEnv">
          <el-input v-model="form.secretEnv" placeholder="例如 PORTAL_WECHAT_APP_SECRET_CAMIE" />
          <div class="form-help">
            后台只保存变量名。请在 Portal 后端进程设置该环境变量为 AppSecret，然后重启后端；密钥不会写入数据库。
          </div>
        </el-form-item>
        <el-form-item label="采集范围" prop="syncMode">
          <el-select v-model="form.syncMode" style="width: 100%">
            <el-option label="已发表文章（推荐）" value="published" />
            <el-option label="永久图文素材" value="materials" />
          </el-select>
          <div class="form-help">
            “已发表文章”用于采集正式内容；账号接口权限不足时，可切换为“永久图文素材”。
          </div>
        </el-form-item>
        <el-form-item label="导入默认分类">
          <el-select v-model="form.defaultCategoryId" clearable placeholder="不指定，导入后人工分类" style="width: 100%">
            <el-option v-for="item in categories" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="logsVisible" :title="`${currentSourceName} · 同步记录`" width="900px">
      <el-table :data="logs" v-loading="logsLoading" border stripe>
        <el-table-column prop="createdAt" label="提交时间" width="180">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column prop="operator" label="操作人" width="120" show-overflow-tooltip />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="syncStatusType(row.status)">{{ syncStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="scannedCount" label="扫描" width="80" align="center" />
        <el-table-column prop="createdCount" label="新增" width="80" align="center" />
        <el-table-column prop="skippedCount" label="跳过" width="80" align="center" />
        <el-table-column label="结果" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">{{ row.errorMessage || syncResultText(row) }}</template>
        </el-table-column>
      </el-table>
      <div class="pagination">
        <el-pagination
          v-model:current-page="logsQuery.page"
          v-model:page-size="logsQuery.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="logsTotal"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchLogs"
          @current-change="fetchLogs"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, RefreshRight, Search } from '@element-plus/icons-vue'
import { getAllCategories } from '@/api/category'
import {
  createContentSource,
  deleteContentSource,
  getContentSourceLogs,
  getContentSources,
  getContentSyncLog,
  startContentSourceSync,
  testContentSource,
  updateContentSource,
  type ContentSource,
  type ContentSyncLog,
  type ContentSyncStatus
} from '@/api/content-source'

type CategoryOption = { id: number; name: string }

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('新增公众号')
const formRef = ref()
const tableData = ref<ContentSource[]>([])
const total = ref(0)
const categories = ref<CategoryOption[]>([])
const testingIds = ref<number[]>([])
const syncingIds = ref<number[]>([])

const logsVisible = ref(false)
const logsLoading = ref(false)
const logs = ref<ContentSyncLog[]>([])
const logsTotal = ref(0)
const currentSourceId = ref<number>()
const currentSourceName = ref('')
const logsQuery = reactive({ page: 1, pageSize: 10 })
const pollTimers = new Set<ReturnType<typeof setTimeout>>()

const queryForm = reactive({
  page: 1,
  pageSize: 10,
  name: '',
  status: undefined as number | undefined
})

const emptyForm = (): ContentSource => ({
  name: '',
  provider: 'wechat_official',
  accountType: 'service',
  appId: '',
  secretEnv: '',
  secretConfigured: false,
  syncMode: 'published',
  status: 1,
  defaultCategoryId: undefined
})
const form = reactive<ContentSource>(emptyForm())

const formRules = {
  name: [{ required: true, message: '请输入账号名称', trigger: 'blur' }],
  appId: [{ required: true, message: '请输入 AppID', trigger: 'blur' }],
  secretEnv: [{
    validator: (_rule: unknown, value: string, callback: (error?: Error) => void) => {
      if (!/^[A-Z][A-Z0-9_]*$/.test(value?.trim() || '')) callback(new Error('请输入大写字母、数字和下划线组成的环境变量名'))
      else callback()
    },
    trigger: 'blur'
  }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getContentSources({ ...queryForm })
    tableData.value = res.data?.list || []
    total.value = res.data?.total || 0
    syncingIds.value = tableData.value
      .filter((item) => item.lastSyncStatus === 'queued' || item.lastSyncStatus === 'running')
      .map((item) => item.id as number)
  } finally {
    loading.value = false
  }
}

const fetchCategories = async () => {
  const res: any = await getAllCategories()
  categories.value = res.data || []
}

const handleSearch = () => {
  queryForm.page = 1
  fetchData()
}

const resetQuery = () => {
  queryForm.name = ''
  queryForm.status = undefined
  queryForm.page = 1
  fetchData()
}

const handleAdd = () => {
  Object.assign(form, emptyForm())
  dialogTitle.value = '新增公众号'
  dialogVisible.value = true
}

const handleEdit = (row: ContentSource) => {
  Object.assign(form, {
    id: row.id,
    name: row.name,
    provider: row.provider,
    accountType: row.accountType,
    appId: row.appId,
    secretEnv: row.secretEnv || '',
    secretConfigured: row.secretConfigured,
    syncMode: row.syncMode,
    status: row.status,
    defaultCategoryId: row.defaultCategoryId
  })
  dialogTitle.value = '编辑公众号'
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    const payload = { ...form, secretEnv: form.secretEnv.trim() }
    if (payload.id) {
      await updateContentSource(payload.id, payload)
      ElMessage.success('公众号配置已更新')
    } else {
      await createContentSource(payload)
      ElMessage.success('公众号配置已创建')
    }
    dialogVisible.value = false
    await fetchData()
  } finally {
    submitLoading.value = false
  }
}

const handleDelete = async (row: ContentSource) => {
  await ElMessageBox.confirm(`确定删除公众号「${row.name}」吗？已有导入记录时系统会阻止删除。`, '删除确认', {
    type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消'
  })
  await deleteContentSource(row.id as number)
  ElMessage.success('删除成功')
  await fetchData()
}

const handleTest = async (row: ContentSource) => {
  const id = row.id as number
  testingIds.value = [...testingIds.value, id]
  try {
    await testContentSource(id)
    ElMessage.success('连接成功')
    await fetchData()
  } finally {
    testingIds.value = testingIds.value.filter((item) => item !== id)
  }
}

const handleSync = async (row: ContentSource) => {
  const id = row.id as number
  await ElMessageBox.confirm(`将从「${row.name}」采集文章并导入为草稿，是否继续？`, '开始同步', {
    type: 'info', confirmButtonText: '开始同步', cancelButtonText: '取消'
  })
  syncingIds.value = [...new Set([...syncingIds.value, id])]
  try {
    const res: any = await startContentSourceSync(id)
    const log = res.data as ContentSyncLog
    ElMessage.success('同步任务已提交')
    pollSync(log.id, id)
    await fetchData()
  } catch (error) {
    syncingIds.value = syncingIds.value.filter((item) => item !== id)
  }
}

const pollSync = (logId: number, sourceId: number) => {
  const timer = setTimeout(async () => {
    pollTimers.delete(timer)
    try {
      const res: any = await getContentSyncLog(logId)
      const log = res.data as ContentSyncLog
      if (log.status === 'success' || log.status === 'failed') {
        syncingIds.value = syncingIds.value.filter((item) => item !== sourceId)
        if (log.status === 'success') {
          ElMessage.success(`同步完成：新增 ${log.createdCount} 篇，跳过 ${log.skippedCount} 篇`)
        } else {
          ElMessage.error(log.errorMessage || '同步失败')
        }
        await fetchData()
        if (logsVisible.value && currentSourceId.value === sourceId) await fetchLogs()
        return
      }
      pollSync(logId, sourceId)
    } catch {
      syncingIds.value = syncingIds.value.filter((item) => item !== sourceId)
    }
  }, 2000)
  pollTimers.add(timer)
}

const openLogs = async (row: ContentSource) => {
  currentSourceId.value = row.id
  currentSourceName.value = row.name
  logsQuery.page = 1
  logsVisible.value = true
  await fetchLogs()
}

const fetchLogs = async () => {
  if (!currentSourceId.value) return
  logsLoading.value = true
  try {
    const res: any = await getContentSourceLogs(currentSourceId.value, { ...logsQuery })
    logs.value = res.data?.list || []
    logsTotal.value = res.data?.total || 0
  } finally {
    logsLoading.value = false
  }
}

const accountTypeText = (type: string) => type === 'subscription' ? '订阅号' : '服务号'
const syncModeText = (mode: string) => mode === 'materials' ? '永久图文素材' : '已发表文章'
const categoryName = (id?: number) => categories.value.find((item) => item.id === id)?.name || '未指定'

const syncStatusText = (status: ContentSyncStatus | string) => ({
  queued: '等待中', running: '同步中', success: '成功', failed: '失败'
}[status] || status)

const syncStatusType = (status: ContentSyncStatus | string): 'success' | 'danger' | 'warning' | 'info' => ({
  queued: 'info', running: 'warning', success: 'success', failed: 'danger'
}[status] as 'success' | 'danger' | 'warning' | 'info' || 'info')

const syncResultText = (log: ContentSyncLog) => {
  if (log.status === 'queued') return '等待执行'
  if (log.status === 'running') return '正在从公众号读取文章'
  return `扫描 ${log.scannedCount} 篇，新增 ${log.createdCount} 篇，跳过 ${log.skippedCount} 篇`
}

const formatTime = (value?: string) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'

onMounted(async () => {
  await Promise.all([fetchCategories(), fetchData()])
})

onBeforeUnmount(() => {
  for (const timer of pollTimers) clearTimeout(timer)
  pollTimers.clear()
})
</script>

<style scoped lang="scss">
.source-tip {
  margin-bottom: 16px;
}

.header-subtitle {
  margin-left: 12px;
  color: var(--app-text-secondary);
  font-size: 13px;
  font-weight: 400;
}

.sync-status-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.muted,
.last-message,
.form-help {
  color: var(--app-text-secondary);
  font-size: 12px;
}

.last-message {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.form-help {
  margin-top: 6px;
  line-height: 1.5;
}
</style>
