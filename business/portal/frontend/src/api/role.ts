import request from '@/utils/request'

export interface RoleQuery {
  page?: number
  pageSize?: number
  name?: string
}

export interface RoleForm {
  id?: number
  name: string
  code: string
  description: string
  status: number
  permissions?: string
}

export function getRoleList(params: RoleQuery) {
  return request.get('/roles', { params })
}

export function getAllRoles() {
  return request.get('/roles/all')
}

export function getRoleById(id: number) {
  return request.get(`/roles/${id}`)
}

export function createRole(data: RoleForm) {
  return request.post('/roles', data)
}

export function updateRole(id: number, data: RoleForm) {
  return request.put(`/roles/${id}`, data)
}

export function deleteRole(id: number) {
  return request.delete(`/roles/${id}`)
}

export function getRolePermissions(id: number) {
  return request.get(`/roles/${id}/permissions`)
}

export function updateRolePermissions(id: number, permissions: number[]) {
  return request.put(`/roles/${id}/permissions`, { permissions })
}

export function getMenuTree() {
  return request.get('/menus/tree')
}
