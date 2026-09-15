<template>
  <el-drawer
    :model-value="modelValue"
    :title="detail ? `流程详情 - ${detail.instance.title}` : '流程详情'"
    size="640px"
    @update:model-value="(v: any) => emit('update:modelValue', v)"
  >
    <div v-loading="loading" class="workflow-detail">
      <template v-if="detail">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="流程">{{ detail.instance.workflowName }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusOf(detail.instance.status).type" size="small">
              {{ statusOf(detail.instance.status).label }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="发起人">{{ detail.instance.initiatorName }}</el-descriptions-item>
          <el-descriptions-item label="发起时间">{{ formatTime(detail.instance.startAt) }}</el-descriptions-item>
          <el-descriptions-item label="当前节点">{{ detail.instance.currentNode || '-' }}</el-descriptions-item>
          <el-descriptions-item label="结束时间">
            {{ detail.instance.endAt ? formatTime(detail.instance.endAt) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item v-if="detail.instance.businessType" label="业务类型">
            {{ detail.instance.businessType }}
          </el-descriptions-item>
          <el-descriptions-item v-if="detail.instance.businessId" label="业务ID">
            {{ detail.instance.businessId }}
          </el-descriptions-item>
          <el-descriptions-item label="申请说明" :span="2">
            {{ detail.instance.content || '-' }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="section-title">审批任务</div>
        <el-table :data="detail.tasks" border size="small">
          <el-table-column prop="nodeName" label="节点" min-width="130">
            <template #default="{ row }">
              {{ row.nodeName }}
              <el-tag v-if="row.approveMode === 'and'" size="small" type="warning" effect="plain">会签</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="approverName" label="审批人" min-width="100" />
          <el-table-column label="结果" width="100">
            <template #default="{ row }">
              <el-tag :type="taskStatusOf(row.status).type" size="small">{{ taskStatusOf(row.status).label }}</el-tag>
              <el-tooltip v-if="row.remindCount" :content="`已催办 ${row.remindCount} 次，最近 ${formatTime(row.remindedAt)}`">
                <el-tag size="small" type="danger" effect="plain" style="margin-left: 4px">催</el-tag>
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column prop="comment" label="意见" show-overflow-tooltip />
          <el-table-column label="处理时间" width="160">
            <template #default="{ row }">{{ row.handledAt ? formatTime(row.handledAt) : '-' }}</template>
          </el-table-column>
        </el-table>

        <div class="section-title">流转记录</div>
        <el-timeline>
          <el-timeline-item
            v-for="log in detail.logs"
            :key="log.id"
            :timestamp="formatTime(log.createdAt)"
            placement="top"
          >
            <div class="log-line">
              <el-tag size="small" effect="plain">{{ logActionMap[log.action] || log.action }}</el-tag>
              <span class="log-who">{{ log.operatorName }}</span>
              <span v-if="log.nodeName" class="log-node">（{{ log.nodeName }}）</span>
            </div>
            <div v-if="log.comment" class="log-comment">{{ log.comment }}</div>
          </el-timeline-item>
        </el-timeline>
      </template>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { getWorkflowInstanceDetail, instanceStatusMap, taskStatusMap, logActionMap } from '@/api/workflow'
import type { WorkflowInstanceDetail } from '@/api/workflow'

const props = defineProps<{ modelValue: boolean; instanceId?: number }>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const loading = ref(false)
const detail = ref<WorkflowInstanceDetail | null>(null)

const statusOf = (status: number) => instanceStatusMap[status] || { label: '未知', type: 'info' as const }
const taskStatusOf = (status: number) => taskStatusMap[status] || { label: '未知', type: 'info' as const }
const formatTime = (value?: string) => (value ? value.replace('T', ' ').slice(0, 19) : '-')

const load = async (id: number) => {
  loading.value = true
  try {
    const res: any = await getWorkflowInstanceDetail(id)
    detail.value = res.data
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.modelValue, props.instanceId] as const,
  ([visible, id]) => {
    if (visible && id) {
      detail.value = null
      load(id)
    }
  },
  { immediate: true }
)
</script>

<style scoped lang="scss">
.workflow-detail {
  .section-title {
    margin: 20px 0 12px;
    font-size: 14px;
    font-weight: 600;
    color: #1e293b;
  }
  .log-line {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .log-who {
    font-weight: 600;
    color: #334155;
  }
  .log-node {
    color: #94a3b8;
    font-size: 12px;
  }
  .log-comment {
    margin-top: 4px;
    font-size: 12px;
    color: #64748b;
    white-space: pre-wrap;
  }
}
</style>
