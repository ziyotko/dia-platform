<template>
  <div class="workflow-instance-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>流程实例</span>
          <el-button type="primary" @click="handleStart">发起流程</el-button>
        </div>
      </template>

      <el-form :inline="true" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="标题 / 流程名称" clearable style="width: 200px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="流程">
          <el-select v-model="query.workflowId" placeholder="全部" clearable style="width: 200px">
            <el-option v-for="wf in workflowOptions" :key="wf.id" :label="wf.name" :value="wf.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" placeholder="全部" clearable style="width: 140px">
            <el-option v-for="(item, key) in instanceStatusMap" :key="key" :label="item.label" :value="Number(key)" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isAdmin">
          <el-checkbox v-model="onlyMine">只看我发起的</el-checkbox>
        </el-form-item>
        <el-form-item v-if="isSuperAdmin" label="所属租户">
          <tenant-select v-model="query.tenantId" placeholder="全部租户" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" border>
        <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
        <el-table-column prop="workflowName" label="流程" min-width="140" />
        <el-table-column v-if="isSuperAdmin" label="所属租户" min-width="130">
          <template #default="{ row }">{{ tenantName(row.tenantId) }}</template>
        </el-table-column>
        <el-table-column prop="initiatorName" label="发起人" width="110" />
        <el-table-column label="当前节点" min-width="120">
          <template #default="{ row }">{{ row.currentNode || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusOf(row.status).type">{{ statusOf(row.status).label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发起时间" width="160">
          <template #default="{ row }">{{ formatTime(row.startAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">详情</el-button>
            <el-button v-if="canCancel(row)" link type="danger" @click="handleCancel(row)">撤销</el-button>
            <el-button v-if="can('base:workflow-instance:delete')" link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="startDialogVisible" title="发起流程" width="600px">
      <el-form :model="startForm" :rules="startRules" ref="startFormRef" label-width="100px">
        <el-form-item label="流程" prop="workflowId">
          <el-select v-model="startForm.workflowId" filterable placeholder="请选择流程" style="width: 100%">
            <el-option v-for="wf in workflowOptions" :key="wf.id" :label="wf.name" :value="wf.id">
              <span>{{ wf.name }}</span>
              <span class="option-tip">{{ wf.nodes?.length || 0 }} 个节点</span>
            </el-option>
          </el-select>
          <div v-if="selectedWorkflow && !(selectedWorkflow.nodes || []).length" class="form-tip warn">
            该流程尚未配置审批节点，无法发起
          </div>
        </el-form-item>
        <el-form-item label="标题" prop="title">
          <el-input v-model="startForm.title" placeholder="如：2026 年度报告发布申请" />
        </el-form-item>
        <el-form-item label="业务类型">
          <el-input v-model="startForm.businessType" placeholder="可选，用于关联业务单据，如 article" />
        </el-form-item>
        <el-form-item label="业务ID">
          <el-input v-model="startForm.businessId" placeholder="可选" />
        </el-form-item>
        <el-form-item label="申请说明">
          <el-input v-model="startForm.content" type="textarea" :rows="3" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="startDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitStart">提交</el-button>
      </template>
    </el-dialog>

    <workflow-detail v-model="detailVisible" :instance-id="detailId" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getWorkflowInstances,
  getWorkflowOptions,
  startWorkflow,
  cancelWorkflow,
  deleteWorkflowInstance,
  instanceStatusMap
} from '@/api/workflow'
import type { StartWorkflowReq, Workflow, WorkflowInstance } from '@/api/workflow'
import { tenantName } from '@/utils/tenantOptions'
import { useUserStore } from '@/stores/user'
import TenantSelect from '@/components/TenantSelect.vue'
import WorkflowDetail from '@/components/WorkflowDetail.vue'

const userStore = useUserStore()
const isSuperAdmin = computed(() => userStore.userInfo?.tenantId === 0)
const isAdmin = computed(() => isSuperAdmin.value || !!userStore.userInfo?.isAdmin)
const can = (code: string) => userStore.can(code)

const loading = ref(false)
const tableData = ref<WorkflowInstance[]>([])
const total = ref(0)
const workflowOptions = ref<Workflow[]>([])
const onlyMine = ref(false)
const detailVisible = ref(false)
const detailId = ref<number>()
const startDialogVisible = ref(false)
const startFormRef = ref<any>(null)

const query = reactive<{ page: number; size: number; keyword: string; workflowId?: number; status?: number; tenantId?: number }>({
  page: 1,
  size: 10,
  keyword: '',
  workflowId: undefined,
  status: undefined,
  tenantId: undefined
})

const startForm = reactive<StartWorkflowReq>({
  workflowId: 0,
  title: '',
  businessType: '',
  businessId: '',
  content: ''
})

const startRules = {
  workflowId: [{ required: true, message: '请选择流程', trigger: 'change' }],
  title: [{ required: true, message: '请填写标题', trigger: 'blur' }]
}

const statusOf = (status: number) => instanceStatusMap[status] || { label: '未知', type: 'info' as const }
const formatTime = (value?: string) => (value ? value.replace('T', ' ').slice(0, 19) : '-')
const selectedWorkflow = computed(() => workflowOptions.value.find((wf) => wf.id === startForm.workflowId))

// 仅审批中的实例可撤销，且必须是发起人本人或管理员
const canCancel = (row: WorkflowInstance) =>
  row.status === 1 && (isAdmin.value || row.initiatorId === userStore.userInfo?.id)

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getWorkflowInstances({ ...query, mine: onlyMine.value ? 1 : undefined })
    tableData.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

const fetchOptions = async () => {
  const res: any = await getWorkflowOptions()
  workflowOptions.value = res.data || []
}

const handleSearch = () => {
  query.page = 1
  fetchData()
}

const handleReset = () => {
  query.keyword = ''
  query.workflowId = undefined
  query.status = undefined
  query.tenantId = undefined
  onlyMine.value = false
  query.page = 1
  fetchData()
}

const handleStart = async () => {
  startForm.workflowId = 0
  startForm.title = ''
  startForm.businessType = ''
  startForm.businessId = ''
  startForm.content = ''
  if (!workflowOptions.value.length) {
    await fetchOptions()
  }
  startDialogVisible.value = true
}

const handleSubmitStart = async () => {
  const valid = await startFormRef.value?.validate().catch(() => false)
  if (!valid) return
  if (selectedWorkflow.value && !(selectedWorkflow.value.nodes || []).length) {
    ElMessage.warning('该流程尚未配置审批节点，无法发起')
    return
  }
  await startWorkflow(startForm)
  ElMessage.success('已提交，等待审批')
  startDialogVisible.value = false
  fetchData()
}

const showDetail = (row: WorkflowInstance) => {
  detailId.value = row.id
  detailVisible.value = true
}

const handleCancel = async (row: WorkflowInstance) => {
  await ElMessageBox.confirm(`确认撤销「${row.title}」？撤销后流程立即终止。`, '提示', { type: 'warning' })
  await cancelWorkflow(row.id)
  ElMessage.success('已撤销')
  fetchData()
}

const handleDelete = async (row: WorkflowInstance) => {
  await ElMessageBox.confirm(`确认删除流程实例「${row.title}」？其审批任务与流转记录会一并删除。`, '提示', { type: 'warning' })
  await deleteWorkflowInstance(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(() => {
  fetchData()
  fetchOptions()
})
</script>

<style scoped lang="scss">
.workflow-instance-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .search-form {
    margin-bottom: 16px;
  }
  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
  .option-tip {
    float: right;
    color: #94a3b8;
    font-size: 12px;
  }
  .form-tip {
    width: 100%;
    font-size: 12px;
    color: #94a3b8;
    line-height: 1.6;
  }
  .form-tip.warn {
    color: #e6a23c;
  }
}
</style>
