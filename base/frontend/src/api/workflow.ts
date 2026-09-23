import request from '@/utils/request'

/* ---------------- 流程定义 ---------------- */

/** 节点审批人类型 */
export type ApproverType = 'role' | 'user' | 'initiator'

/** 节点审批方式：or 或签（任一处理即完成）/ and 会签（全部通过才完成） */
export type ApproveMode = 'or' | 'and'

export interface WorkflowNode {
  id?: number
  workflowId?: number
  name: string
  sort?: number
  approverType: ApproverType
  approverId: number
  /** 审批方式，默认 or */
  approveMode?: ApproveMode
  /** 超时提醒（分钟），0 表示不提醒 */
  timeoutMinutes?: number
  description?: string
  /** 审批人展示名（后端回填） */
  approverName?: string
}

export interface Workflow {
  id: number
  tenantId?: number
  code: string
  name: string
  description?: string
  status: number
  nodes?: WorkflowNode[]
}

export interface WorkflowQuery {
  page: number
  size: number
  keyword?: string
  status?: number
  tenantId?: number
}

export interface ApproverOptions {
  roles: { id: number; name: string; tenantId: number; status: number }[]
  users: { id: number; username: string; realName?: string; status: number }[]
}

export function getWorkflowList(params: WorkflowQuery) {
  return request.get('/workflows', { params })
}

export function getWorkflow(id: number) {
  return request.get(`/workflows/${id}`)
}

export function createWorkflow(data: Partial<Workflow>) {
  return request.post('/workflows', data)
}

export function updateWorkflow(id: number, data: Partial<Workflow>) {
  return request.put(`/workflows/${id}`, data)
}

export function deleteWorkflow(id: number) {
  return request.delete(`/workflows/${id}`)
}

/** 覆盖式保存节点编排 */
export function saveWorkflowNodes(id: number, nodes: WorkflowNode[]) {
  return request.put(`/workflows/${id}/nodes`, { nodes })
}

/** 启用中的流程定义（发起流程时选择，任意登录用户可用） */
export function getWorkflowOptions() {
  return request.get('/workflows/options')
}

/** 节点审批人候选（流程角色 + 用户）。平台超管可传 tenantId 限定只取某租户的候选 */
export function getWorkflowApproverOptions(tenantId?: number) {
  return request.get('/workflows/approver-options', { params: tenantId ? { tenantId } : {} })
}

/* ---------------- 流程实例 ---------------- */

/** 1 审批中 2 已通过 3 已驳回 4 已撤销 */
export interface WorkflowInstance {
  id: number
  tenantId: number
  workflowId: number
  workflowName: string
  title: string
  businessType?: string
  businessId?: string
  content?: string
  initiatorId: number
  initiatorName: string
  currentNode?: string
  status: number
  startAt: string
  endAt?: string
}

export interface WorkflowTask {
  id: number
  instanceId: number
  nodeId: number
  nodeName: string
  nodeSort: number
  /** 审批方式快照：or 或签 / and 会签 */
  approveMode?: ApproveMode
  timeoutMinutes?: number
  remindedAt?: string
  remindCount?: number
  approverId: number
  approverName: string
  status: number
  comment?: string
  handledAt?: string
  createdAt: string
  instance?: WorkflowInstance
}

export interface WorkflowLog {
  id: number
  instanceId: number
  nodeName?: string
  operatorName: string
  action: string
  comment?: string
  createdAt: string
}

export interface WorkflowInstanceDetail {
  instance: WorkflowInstance
  tasks: WorkflowTask[]
  logs: WorkflowLog[]
}

export interface WorkflowInstanceQuery {
  page: number
  size: number
  keyword?: string
  workflowId?: number
  status?: number
  /** 只看我发起的（非管理员由后端强制为 true） */
  mine?: number
  tenantId?: number
}

export interface StartWorkflowReq {
  workflowId: number
  title: string
  businessType?: string
  businessId?: string
  content?: string
}

export function getWorkflowInstances(params: WorkflowInstanceQuery) {
  return request.get('/workflow-instances', { params })
}

export function getWorkflowInstanceDetail(id: number) {
  return request.get(`/workflow-instances/${id}`)
}

export function startWorkflow(data: StartWorkflowReq) {
  return request.post('/workflow-instances', data)
}

export function cancelWorkflow(id: number) {
  return request.post(`/workflow-instances/${id}/cancel`)
}

/** 删除流程实例（仅管理员，会一并删除任务与流转日志） */
export function deleteWorkflowInstance(id: number) {
  return request.delete(`/workflow-instances/${id}`)
}

/* ---------------- 我的待办 ---------------- */

export function getMyWorkflowTasks(params: { box: 'todo' | 'done'; page: number; size: number }) {
  return request.get('/workflow-tasks', { params })
}

/** 转办/加签可选的用户（本租户启用中的用户） */
export function getWorkflowTaskApproverOptions() {
  return request.get('/workflow-tasks/approver-options')
}

export function approveWorkflowTask(id: number, comment: string) {
  return request.post(`/workflow-tasks/${id}/approve`, { comment })
}

/** 转办：把待办交给同租户的另一个用户 */
export function transferWorkflowTask(id: number, userId: number, comment?: string) {
  return request.post(`/workflow-tasks/${id}/transfer`, { userId, comment })
}

/** 加签：在当前节点追加一个审批人 */
export function addApproverWorkflowTask(id: number, userId: number, comment?: string) {
  return request.post(`/workflow-tasks/${id}/add-approver`, { userId, comment })
}

export function rejectWorkflowTask(id: number, comment: string) {
  return request.post(`/workflow-tasks/${id}/reject`, { comment })
}

/** 流程实例状态文案与标签色 */
export const instanceStatusMap: Record<number, { label: string; type: 'primary' | 'success' | 'warning' | 'danger' | 'info' }> = {
  1: { label: '审批中', type: 'warning' },
  2: { label: '已通过', type: 'success' },
  3: { label: '已驳回', type: 'danger' },
  4: { label: '已撤销', type: 'info' }
}

/** 审批任务状态文案与标签色 */
export const taskStatusMap: Record<number, { label: string; type: 'primary' | 'success' | 'warning' | 'danger' | 'info' }> = {
  1: { label: '待处理', type: 'warning' },
  2: { label: '已通过', type: 'success' },
  3: { label: '已驳回', type: 'danger' },
  4: { label: '已失效', type: 'info' }
}

/** 流转日志动作文案 */
export const logActionMap: Record<string, string> = {
  start: '发起流程',
  approve: '审批通过',
  reject: '审批驳回',
  cancel: '撤销流程',
  'auto-pass': '自动通过',
  finish: '流程完成',
  transfer: '转办',
  'add-approver': '加签',
  remind: '超时提醒'
}

export const approveModeMap: Record<string, string> = {
  or: '或签',
  and: '会签'
}

export const approverTypeMap: Record<string, string> = {
  role: '流程角色',
  user: '指定用户',
  initiator: '发起人本人'
}
