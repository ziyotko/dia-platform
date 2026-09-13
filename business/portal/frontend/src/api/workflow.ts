import request from '@/utils/request'

export interface WorkflowForm {
  id?: number
  name: string
  status: number
  description: string
}

export interface WorkflowNode {
  id?: number
  workflowId?: number
  name: string
  approverType?: 'user' | 'dept_head' | 'role'
  approverId?: number
  sortOrder?: number
}

export interface WorkflowQuery {
  name?: string
  page?: number
  pageSize?: number
  all?: boolean
}

export function getWorkflows(params: WorkflowQuery) {
  return request.get('/workflows', { params })
}

/** 获取全部流程（供栏目绑定流程等下拉使用，不分页） */
export function getAllWorkflows() {
  return request.get('/workflows', { params: { all: 1 } })
}

export function getWorkflowByID(id: number) {
  return request.get(`/workflows/${id}`)
}

export function createWorkflow(data: WorkflowForm) {
  return request.post('/workflows', data)
}

export function updateWorkflow(id: number, data: WorkflowForm) {
  return request.put(`/workflows/${id}`, data)
}

export function deleteWorkflow(id: number) {
  return request.delete(`/workflows/${id}`)
}

export function getWorkflowNodes(id: number) {
  return request.get(`/workflows/${id}/nodes`)
}

export function saveWorkflowNodes(id: number, nodes: WorkflowNode[]) {
  return request.put(`/workflows/${id}/nodes`, { nodes })
}
