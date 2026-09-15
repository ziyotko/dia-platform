<template>
  <div class="workflow-task-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>我的待办</span>
          <el-button :icon="Refresh" @click="fetchData">刷新</el-button>
        </div>
      </template>

      <el-tabs v-model="box" @tab-change="handleTabChange">
        <el-tab-pane label="待我审批" name="todo" />
        <el-tab-pane label="我已处理" name="done" />
      </el-tabs>

      <el-table :data="tableData" v-loading="loading" border>
        <el-table-column label="标题" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.instance?.title || '-' }}</template>
        </el-table-column>
        <el-table-column label="流程" min-width="140">
          <template #default="{ row }">{{ row.instance?.workflowName || '-' }}</template>
        </el-table-column>
        <el-table-column prop="nodeName" label="审批节点" min-width="120" />
        <el-table-column label="发起人" width="110">
          <template #default="{ row }">{{ row.instance?.initiatorName || '-' }}</template>
        </el-table-column>
        <el-table-column label="提交时间" width="160">
          <template #default="{ row }">{{ formatTime(row.instance?.startAt) }}</template>
        </el-table-column>
        <el-table-column v-if="box === 'done'" label="处理结果" width="110">
          <template #default="{ row }">
            <el-tag :type="taskStatusOf(row.status).type">{{ taskStatusOf(row.status).label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="流程状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.instance" :type="instanceStatusOf(row.instance.status).type" effect="plain">
              {{ instanceStatusOf(row.instance.status).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">详情</el-button>
            <template v-if="box === 'todo'">
              <el-button link type="success" @click="openHandle(row, true)">通过</el-button>
              <el-button link type="danger" @click="openHandle(row, false)">驳回</el-button>
            </template>
            <span v-else class="comment-tip">{{ row.comment || '-' }}</span>
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

    <el-dialog v-model="handleVisible" :title="approveMode ? '审批通过' : '审批驳回'" width="520px">
      <el-alert
        v-if="currentTask.instance"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 12px"
        :title="`${currentTask.instance.workflowName} · ${currentTask.nodeName}`"
      />
      <el-form label-width="80px">
        <el-form-item label="审批意见">
          <el-input v-model="comment" type="textarea" :rows="3" :placeholder="approveMode ? '可选' : '建议填写驳回理由'" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleVisible = false">取消</el-button>
        <el-button :type="approveMode ? 'primary' : 'danger'" :loading="submitting" @click="handleSubmit">
          {{ approveMode ? '确认通过' : '确认驳回' }}
        </el-button>
      </template>
    </el-dialog>

    <workflow-detail v-model="detailVisible" :instance-id="detailId" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import {
  getMyWorkflowTasks,
  approveWorkflowTask,
  rejectWorkflowTask,
  instanceStatusMap,
  taskStatusMap
} from '@/api/workflow'
import type { WorkflowTask } from '@/api/workflow'
import WorkflowDetail from '@/components/WorkflowDetail.vue'

const loading = ref(false)
const submitting = ref(false)
const box = ref<'todo' | 'done'>('todo')
const tableData = ref<WorkflowTask[]>([])
const total = ref(0)
const query = reactive({ page: 1, size: 10 })
const handleVisible = ref(false)
const approveMode = ref(true)
const comment = ref('')
const currentTask = ref<WorkflowTask>({} as WorkflowTask)
const detailVisible = ref(false)
const detailId = ref<number>()

const statusOf = (map: typeof taskStatusMap, status: number) => map[status] || { label: '未知', type: 'info' as const }
const taskStatusOf = (status: number) => statusOf(taskStatusMap, status)
const instanceStatusOf = (status: number) => statusOf(instanceStatusMap, status)
const formatTime = (value?: string) => (value ? value.replace('T', ' ').slice(0, 19) : '-')

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getMyWorkflowTasks({ box: box.value, page: query.page, size: query.size })
    tableData.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

const handleTabChange = () => {
  query.page = 1
  fetchData()
}

const openHandle = (row: WorkflowTask, approve: boolean) => {
  currentTask.value = row
  approveMode.value = approve
  comment.value = ''
  handleVisible.value = true
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (approveMode.value) {
      await approveWorkflowTask(currentTask.value.id, comment.value)
      ElMessage.success('审批通过')
    } else {
      await rejectWorkflowTask(currentTask.value.id, comment.value)
      ElMessage.success('已驳回')
    }
    handleVisible.value = false
    fetchData()
  } finally {
    submitting.value = false
  }
}

const showDetail = (row: WorkflowTask) => {
  detailId.value = row.instanceId
  detailVisible.value = true
}

onMounted(fetchData)
</script>

<style scoped lang="scss">
.workflow-task-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
  .comment-tip {
    color: #94a3b8;
    font-size: 12px;
  }
}
</style>
