<template>
  <div class="workflow-def-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>流程定义</span>
          <el-button v-if="can('base:workflow-def:create')" type="primary" @click="handleAdd">新增流程</el-button>
        </div>
      </template>

      <el-alert
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 16px"
        title="节点按顺序串行审批；同一节点内多个审批人为「或签」（任一人处理即通过该节点）；节点无有效审批人时自动通过"
      />

      <el-form :inline="true" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="流程名称 / 编码" clearable style="width: 200px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" placeholder="全部" clearable style="width: 140px">
            <el-option label="启用" :value="1" />
            <el-option label="停用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isSuperAdmin" label="所属租户">
          <tenant-select v-model="query.tenantId" placeholder="全部租户" clearable style="width: 220px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="tableData" v-loading="loading" border>
        <el-table-column prop="code" label="流程编码" min-width="140" />
        <el-table-column prop="name" label="流程名称" min-width="140" />
        <el-table-column v-if="isSuperAdmin" label="所属租户" min-width="140">
          <template #default="{ row }">{{ tenantName(row.tenantId) }}</template>
        </el-table-column>
        <el-table-column label="审批节点" min-width="220">
          <template #default="{ row }">
            <span v-if="!row.nodes || !row.nodes.length" class="empty-tip">未配置节点</span>
            <template v-else>
              <el-tag v-for="node in row.nodes" :key="node.id" size="small" effect="plain" style="margin-right: 6px">
                {{ node.sort }}. {{ node.name }}
              </el-tag>
            </template>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="流程说明" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button v-if="can('base:workflow-def:update')" link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button v-if="can('base:workflow-def:nodes')" link type="primary" @click="handleNodes(row)">节点编排</el-button>
            <el-button v-if="can('base:workflow-def:delete')" link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑流程' : '新增流程'" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item v-if="isSuperAdmin" label="所属租户" prop="tenantId">
          <tenant-select v-model="form.tenantId" :disabled="isEdit" :clearable="!isEdit" />
          <div class="form-tip">编辑时不允许变更所属租户</div>
        </el-form-item>
        <el-form-item label="流程编码" prop="code">
          <el-input v-model="form.code" placeholder="如：content-audit" />
          <div class="form-tip">租户内唯一，编辑时不可修改</div>
        </el-form-item>
        <el-form-item label="流程名称" prop="name">
          <el-input v-model="form.name" placeholder="如：内容审核流程" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
          <div class="form-tip">停用后不能发起新的流程实例，已发起的实例不受影响</div>
        </el-form-item>
        <el-form-item label="流程说明">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="nodeDialogVisible" :title="`节点编排 - ${currentWorkflow.name}`" width="900px">
      <el-alert
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 12px"
        title="按从上到下的顺序依次审批；审批人类型可选流程角色（角色成员或签）、指定用户、发起人本人"
      />
      <el-table :data="nodes" border size="small">
        <el-table-column label="顺序" width="90">
          <template #default="{ $index }">
            <el-button link :disabled="$index === 0" @click="moveNode($index, -1)">上移</el-button>
            <el-button link :disabled="$index === nodes.length - 1" @click="moveNode($index, 1)">下移</el-button>
          </template>
        </el-table-column>
        <el-table-column label="节点名称" min-width="150">
          <template #default="{ row }">
            <el-input v-model="row.name" size="small" placeholder="如：初审" />
          </template>
        </el-table-column>
        <el-table-column label="审批人类型" width="140">
          <template #default="{ row }">
            <el-select v-model="row.approverType" size="small" @change="row.approverId = 0">
              <el-option label="流程角色" value="role" />
              <el-option label="指定用户" value="user" />
              <el-option label="发起人本人" value="initiator" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="审批人" min-width="200">
          <template #default="{ row }">
            <span v-if="row.approverType === 'initiator'" class="empty-tip">发起人本人</span>
            <el-select v-else v-model="row.approverId" size="small" filterable placeholder="请选择">
              <el-option
                v-for="opt in approverCandidates(row.approverType)"
                :key="opt.id"
                :label="opt.label"
                :value="opt.id"
              />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="节点说明" min-width="160">
          <template #default="{ row }">
            <el-input v-model="row.description" size="small" placeholder="可选" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80">
          <template #default="{ $index }">
            <el-button link type="danger" @click="nodes.splice($index, 1)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-button style="margin-top: 12px" @click="addNode">添加节点</el-button>
      <div class="form-tip">提示：流程角色需先在「流程角色」中配置成员；角色无成员时该节点会自动通过</div>
      <template #footer>
        <el-button @click="nodeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveNodes">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getWorkflowList,
  getWorkflow,
  createWorkflow,
  updateWorkflow,
  deleteWorkflow,
  saveWorkflowNodes,
  getWorkflowApproverOptions
} from '@/api/workflow'
import type { ApproverOptions, ApproverType, Workflow, WorkflowNode } from '@/api/workflow'
import { tenantName } from '@/utils/tenantOptions'
import { useUserStore } from '@/stores/user'
import TenantSelect from '@/components/TenantSelect.vue'

const userStore = useUserStore()
const isSuperAdmin = computed(() => userStore.userInfo?.tenantId === 0)
const can = (code: string) => userStore.can(code)

const loading = ref(false)
const tableData = ref<Workflow[]>([])
const total = ref(0)
const query = reactive<{ page: number; size: number; keyword: string; status?: number; tenantId?: number }>({
  page: 1,
  size: 10,
  keyword: '',
  status: undefined,
  tenantId: undefined
})

const dialogVisible = ref(false)
const nodeDialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<any>(null)
const currentWorkflow = ref<Workflow>({ id: 0, code: '', name: '', status: 1 })
const nodes = ref<WorkflowNode[]>([])
const approverOptions = ref<ApproverOptions>({ roles: [], users: [] })

const form = reactive<Workflow>({
  id: 0,
  tenantId: 0,
  code: '',
  name: '',
  description: '',
  status: 1
})

const rules = {
  code: [{ required: true, message: '请输入流程编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入流程名称', trigger: 'blur' }]
}

const approverCandidates = (type: ApproverType) => {
  if (type === 'role') {
    return approverOptions.value.roles.map((r) => ({
      id: r.id,
      label: r.status === 1 ? r.name : `${r.name}（已停用）`
    }))
  }
  return approverOptions.value.users.map((u) => ({
    id: u.id,
    label: `${u.realName ? u.realName + '（' + u.username + '）' : u.username}${u.status !== 1 ? ' - 已停用' : ''}`
  }))
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getWorkflowList(query)
    tableData.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  query.page = 1
  fetchData()
}

const handleReset = () => {
  query.keyword = ''
  query.status = undefined
  query.tenantId = undefined
  query.page = 1
  fetchData()
}

const resetForm = () => {
  form.id = 0
  form.tenantId = userStore.userInfo?.tenantId || 0
  form.code = ''
  form.name = ''
  form.description = ''
  form.status = 1
}

const handleAdd = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: Workflow) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    tenantId: row.tenantId,
    code: row.code,
    name: row.name,
    description: row.description || '',
    status: row.status
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (isEdit.value) {
    await updateWorkflow(form.id, form)
  } else {
    await createWorkflow({ ...form, tenantId: isSuperAdmin.value ? form.tenantId : undefined })
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetchData()
}

const handleDelete = async (row: Workflow) => {
  await ElMessageBox.confirm(`确认删除流程「${row.name}」？其节点配置也会被删除。`, '提示', { type: 'warning' })
  await deleteWorkflow(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

const handleNodes = async (row: Workflow) => {
  currentWorkflow.value = row
  const [detail, options]: any[] = await Promise.all([getWorkflow(row.id), getWorkflowApproverOptions()])
  approverOptions.value = options.data || { roles: [], users: [] }
  nodes.value = (detail.data?.nodes || []).map((n: WorkflowNode) => ({
    id: n.id,
    name: n.name,
    sort: n.sort,
    approverType: n.approverType,
    approverId: n.approverId,
    description: n.description || ''
  }))
  nodeDialogVisible.value = true
}

const addNode = () => {
  nodes.value.push({ name: '', approverType: 'role', approverId: 0, description: '' })
}

const moveNode = (index: number, delta: number) => {
  const target = index + delta
  const list = nodes.value
  ;[list[index], list[target]] = [list[target], list[index]]
}

const handleSaveNodes = async () => {
  for (const [index, node] of nodes.value.entries()) {
    if (!node.name) {
      ElMessage.warning(`第 ${index + 1} 个节点请填写名称`)
      return
    }
    if (node.approverType !== 'initiator' && !node.approverId) {
      ElMessage.warning(`第 ${index + 1} 个节点请选择${node.approverType === 'role' ? '流程角色' : '审批人'}`)
      return
    }
  }
  await saveWorkflowNodes(currentWorkflow.value.id, nodes.value)
  ElMessage.success('节点保存成功')
  nodeDialogVisible.value = false
  fetchData()
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.workflow-def-page {
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
  .empty-tip {
    color: #94a3b8;
    font-size: 12px;
  }
  .form-tip {
    width: 100%;
    font-size: 12px;
    color: #94a3b8;
    line-height: 1.6;
  }
}
</style>
