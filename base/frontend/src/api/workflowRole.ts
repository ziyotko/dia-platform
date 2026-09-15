import request from '@/utils/request'

/** 流程角色：审批流程节点的候选人集合，与系统角色（Role）无关 */
export interface WorkflowRole {
  id: number
  tenantId?: number
  code: string
  name: string
  description?: string
  status: number
  userCount?: number
  users?: WorkflowRoleUser[]
}

export interface WorkflowRoleUser {
  id: number
  username: string
  realName?: string
}

export interface WorkflowRoleQuery {
  page: number
  size: number
  keyword?: string
  /** 状态筛选：不传为全部 */
  status?: number
  /** 平台超管可按租户过滤，不传表示全部租户 */
  tenantId?: number
}

/** 成员选择器用的用户选项 */
export interface UserOption {
  id: number
  tenantId: number
  username: string
  realName?: string
  status: number
}

export function getWorkflowRoleList(params: WorkflowRoleQuery) {
  return request.get('/workflow-roles', { params })
}

/** 流程角色详情（含成员） */
export function getWorkflowRole(id: number) {
  return request.get(`/workflow-roles/${id}`)
}

export function createWorkflowRole(data: Partial<WorkflowRole>) {
  return request.post('/workflow-roles', data)
}

export function updateWorkflowRole(id: number, data: Partial<WorkflowRole>) {
  return request.put(`/workflow-roles/${id}`, data)
}

export function deleteWorkflowRole(id: number) {
  return request.delete(`/workflow-roles/${id}`)
}

/** 覆盖式设置流程角色成员 */
export function assignWorkflowRoleUsers(id: number, userIds: number[]) {
  return request.post(`/workflow-roles/${id}/users`, { userIds })
}

/** 可选成员（仅 id/用户名/姓名，避免依赖用户管理权限） */
export function getWorkflowRoleUserOptions(params?: { keyword?: string; tenantId?: number; limit?: number }) {
  return request.get('/workflow-roles/user-options', { params })
}
