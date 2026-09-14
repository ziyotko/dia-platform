import request from '@/utils/request'

export interface WorkflowRoleQuery {
  page?: number
  pageSize?: number
  name?: string
  status?: number
  all?: boolean
}

export interface WorkflowRoleForm {
  id?: number
  name: string
  code: string
  description: string
  status: number
}

export function getWorkflowRoleList(params: WorkflowRoleQuery) {
  return request.get('/workflow-roles', { params })
}

/** 获取全部流程角色（供审批人下拉使用，不分页） */
export function getAllWorkflowRoles(status?: number) {
  return request.get('/workflow-roles', { params: { all: 1, status } })
}

/** 获取流程角色下拉选项（仅 id/名称，任意登录用户可用，用于展示「角色：xxx」） */
export function getWorkflowRoleOptions() {
  return request.get('/workflow-role-options')
}

export function createWorkflowRole(data: WorkflowRoleForm) {
  return request.post('/workflow-roles', data)
}

export function updateWorkflowRole(id: number, data: WorkflowRoleForm) {
  return request.put(`/workflow-roles/${id}`, data)
}

export function deleteWorkflowRole(id: number) {
  return request.delete(`/workflow-roles/${id}`)
}

export function getWorkflowRoleUsers(id: number) {
  return request.get(`/workflow-roles/${id}/users`)
}

export function updateWorkflowRoleUsers(id: number, userIds: number[]) {
  return request.put(`/workflow-roles/${id}/users`, { userIds })
}
