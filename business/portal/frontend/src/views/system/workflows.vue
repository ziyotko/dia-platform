<template>
  <div class="page-container">
    <el-card shadow="hover" class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="流程名称">
          <el-input v-model="queryForm.name" placeholder="请输入流程名称" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>查询
          </el-button>
          <el-button @click="resetQuery">
            <el-icon><RefreshRight /></el-icon>重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="hover" class="table-card">
      <template #header>
        <div class="card-header">
          <span>流程列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增流程
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column type="index" width="60" align="center" />
        <el-table-column prop="name" label="流程名称" min-width="180" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="nodeCount" label="节点数" width="100" align="center" />
        <el-table-column prop="createdAt" label="创建时间" width="170" />
        <el-table-column label="操作" width="280" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDesign(row)">
              <el-icon><SetUp /></el-icon>设计流程
            </el-button>
            <el-button link type="primary" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button link type="danger" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="queryForm.page"
          v-model:page-size="queryForm.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 流程基本信息弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="520px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="90px"
      >
        <el-form-item label="流程名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入流程名称" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="流程说明" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入流程说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 流程设计弹窗 -->
    <el-dialog
      v-model="designVisible"
      title="设计流程"
      width="720px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <div class="design-header">
        <div class="design-title">{{ currentFlow.name }}</div>
        <el-button type="primary" size="small" @click="addNode">
          <el-icon><Plus /></el-icon>添加节点
        </el-button>
      </div>

      <div v-if="nodeList.length === 0" class="empty-nodes">
        <el-empty description="暂无节点，请点击上方按钮添加" />
      </div>

      <div v-else class="node-list">
        <div
          v-for="(node, index) in nodeList"
          :key="node.id"
          class="node-item"
        >
          <div class="node-index">{{ index + 1 }}</div>
          <div class="node-card">
            <div class="node-row">
              <el-input
                v-model="node.name"
                placeholder="节点名称"
                style="width: 160px"
                size="small"
              />
              <el-select
                v-model="node.approverType"
                placeholder="审批人类型"
                style="width: 130px"
                size="small"
                @change="node.approverId = undefined"
              >
                <el-option
                  v-for="type in getApproverTypes(index)"
                  :key="type.value"
                  :label="type.label"
                  :value="type.value"
                />
              </el-select>
              <el-select
                v-if="node.approverType === 'user'"
                v-model="node.approverId"
                placeholder="选择成员"
                clearable
                style="width: 160px"
                size="small"
                filterable
              >
                <el-option
                  v-for="user in userList"
                  :key="user.id"
                  :label="user.username"
                  :value="user.id"
                />
              </el-select>
              <el-select
                v-else-if="node.approverType === 'role'"
                v-model="node.approverId"
                placeholder="选择流程角色"
                clearable
                style="width: 180px"
                size="small"
                filterable
              >
                <el-option
                  v-for="role in workflowRoleList"
                  :key="role.id"
                  :label="role.name"
                  :value="role.id"
                />
              </el-select>
              <el-input
                v-else-if="node.approverType === 'dept_head'"
                value="部门负责人"
                disabled
                style="width: 160px"
                size="small"
              />
            </div>
            <div class="node-actions">
              <el-button
                link
                type="primary"
                size="small"
                :disabled="index === 0"
                @click="moveNode(index, -1)"
              >
                <el-icon><ArrowUp /></el-icon>上移
              </el-button>
              <el-button
                link
                type="primary"
                size="small"
                :disabled="index === nodeList.length - 1"
                @click="moveNode(index, 1)"
              >
                <el-icon><ArrowDown /></el-icon>下移
              </el-button>
              <el-button link type="danger" size="small" @click="removeNode(index)">
                <el-icon><Delete /></el-icon>删除
              </el-button>
            </div>
          </div>
          <div v-if="index < nodeList.length - 1" class="node-arrow">
            <el-icon size="18" color="#002fa7"><Bottom /></el-icon>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="designVisible = false">取消</el-button>
        <el-button type="primary" :loading="designLoading" @click="handleSaveDesign">
          保存设计
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  RefreshRight,
  Plus,
  Edit,
  Delete,
  SetUp,
  ArrowUp,
  ArrowDown,
  Bottom
} from '@element-plus/icons-vue'
import { getAllUsers } from '@/api/user'
import { getAllWorkflowRoles } from '@/api/workflow-role'
import {
  getWorkflows,
  createWorkflow,
  updateWorkflow,
  deleteWorkflow,
  getWorkflowNodes,
  saveWorkflowNodes
} from '@/api/workflow'

const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const submitLoading = ref(false)
const total = ref(0)
const formRef = ref()

const queryForm = reactive({
  page: 1,
  pageSize: 10,
  name: ''
})

const form = reactive({
  id: undefined as number | undefined,
  name: '',
  status: 1,
  description: ''
})

const formRules = {
  name: [{ required: true, message: '请输入流程名称', trigger: 'blur' }]
}

const tableData = ref<any[]>([])

const designVisible = ref(false)
const designLoading = ref(false)
const currentFlow = reactive({ id: 0, name: '' })
const nodeList = ref<any[]>([])
const userList = ref<any[]>([])
const workflowRoleList = ref<any[]>([])

const APPROVER_TYPES = [
  { value: 'user', label: '指定成员' },
  { value: 'dept_head', label: '部门负责人' },
  { value: 'role', label: '角色' }
]

const getApproverTypes = (index: number) => {
  if (index === 0) return APPROVER_TYPES
  return APPROVER_TYPES.filter(t => t.value !== 'dept_head')
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await getWorkflows({
      page: queryForm.page,
      pageSize: queryForm.pageSize,
      name: queryForm.name || undefined
    })
    tableData.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (error) {
    ElMessage.error('获取流程列表失败')
  } finally {
    loading.value = false
  }
}

const fetchUsers = async () => {
  try {
    const res: any = await getAllUsers(1)
    userList.value = res.data?.list || []
  } catch {
    userList.value = []
  }
}

const fetchWorkflowRoles = async () => {
  try {
    const res: any = await getAllWorkflowRoles(1)
    workflowRoleList.value = res.data?.list || []
  } catch {
    workflowRoleList.value = []
  }
}

const handleSearch = () => {
  queryForm.page = 1
  fetchData()
}

const resetQuery = () => {
  queryForm.name = ''
  queryForm.page = 1
  fetchData()
}

const handleAdd = () => {
  dialogTitle.value = '新增流程'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  dialogTitle.value = '编辑流程'
  Object.assign(form, {
    id: row.id,
    name: row.name,
    status: row.status,
    description: row.description || ''
  })
  dialogVisible.value = true
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除流程 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deleteWorkflow(row.id)
      ElMessage.success('删除成功')
      fetchData()
    } catch (error) {
      ElMessage.error('删除流程失败')
    }
  })
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    const data = {
      name: form.name,
      status: form.status,
      description: form.description
    }
    if (form.id) {
      await updateWorkflow(form.id, data)
      ElMessage.success('修改成功')
    } else {
      await createWorkflow(data)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (error) {
    ElMessage.error('提交流程失败')
  } finally {
    submitLoading.value = false
  }
}

const resetForm = () => {
  form.id = undefined
  form.name = ''
  form.status = 1
  form.description = ''
}

const handleSizeChange = (val: number) => {
  queryForm.pageSize = val
  fetchData()
}

const handleCurrentChange = (val: number) => {
  queryForm.page = val
  fetchData()
}

// 流程设计
const handleDesign = async (row: any) => {
  currentFlow.id = row.id
  currentFlow.name = row.name
  try {
    const res: any = await getWorkflowNodes(row.id)
    const nodes = res.data || []
    nodeList.value = nodes.map((n: any, idx: number) => ({
      id: n.id,
      name: n.name,
      approverType: (n.approverType === 'dept_head' && idx > 0 ? 'user' : n.approverType) || 'user',
      approverId: n.approverType === 'dept_head' && idx > 0 ? undefined : n.approverId
    }))
  } catch {
    nodeList.value = []
  }
  designVisible.value = true
}

const addNode = () => {
  nodeList.value.push({
    id: `n${Date.now()}`,
    name: '',
    approverType: 'user',
    approverId: undefined
  })
}

const removeNode = (index: number) => {
  nodeList.value.splice(index, 1)
}

const moveNode = (index: number, direction: number) => {
  const newIndex = index + direction
  if (newIndex < 0 || newIndex >= nodeList.value.length) return
  const temp = nodeList.value[index]
  nodeList.value[index] = nodeList.value[newIndex]
  nodeList.value[newIndex] = temp
}

const handleSaveDesign = async () => {
  // 校验节点名称
  for (let i = 0; i < nodeList.value.length; i++) {
    const node = nodeList.value[i]
    if (!node.name.trim()) {
      ElMessage.warning(`第 ${i + 1} 个节点名称不能为空`)
      return
    }
    if ((node.approverType === 'user' || node.approverType === 'role') && !node.approverId) {
      ElMessage.warning(`第 ${i + 1} 个节点请选择${node.approverType === 'user' ? '成员' : '流程角色'}`)
      return
    }
    if (node.approverType === 'dept_head' && i > 0) {
      ElMessage.warning('只有第一个节点可以设置部门负责人')
      return
    }
  }
  designLoading.value = true
  try {
    const nodes = nodeList.value.map((n, idx) => ({
      name: n.name,
      approverType: n.approverType,
      approverId: n.approverId || undefined,
      sortOrder: idx
    }))
    await saveWorkflowNodes(currentFlow.id, nodes)
    ElMessage.success('流程设计保存成功')
    designVisible.value = false
    fetchData()
  } catch (error) {
    ElMessage.error('保存流程设计失败')
  } finally {
    designLoading.value = false
  }
}

onMounted(() => {
  fetchData()
  fetchUsers()
  fetchWorkflowRoles()
})
</script>

<style scoped lang="scss">
.page-container {
  .search-card {
    margin-bottom: 20px;
    border-radius: 12px;
    border: 1px solid #e6f2ff;
  }

  .table-card {
    border-radius: 12px;
    border: 1px solid #e6f2ff;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-weight: 600;
      color: #2c3e50;
    }
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}

.design-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e6f2ff;

  .design-title {
    font-size: 16px;
    font-weight: 600;
    color: #2c3e50;
  }
}

.empty-nodes {
  padding: 40px 0;
}

.node-list {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0;
}

.node-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
}

.node-index {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: linear-gradient(135deg, #002fa7, #1746d3);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 8px;
  box-shadow: 0 2px 8px rgba(0, 47, 167, 0.25);
}

.node-card {
  width: 100%;
  max-width: 520px;
  background: #fff;
  border: 1px solid #e6f2ff;
  border-radius: 10px;
  padding: 14px 18px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
  transition: all 0.2s ease;

  &:hover {
    border-color: #002fa7;
    box-shadow: 0 4px 16px rgba(0, 47, 167, 0.12);
  }
}

.node-row {
  display: flex;
  gap: 12px;
  margin-bottom: 10px;
}

.node-actions {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
}

.node-arrow {
  padding: 8px 0;
}
</style>
