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
  permissions: number[]
}

export function getRoleList(params: RoleQuery) {
  return request.get('/roles', { params })
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

export function getAllPermissions() {
  return request.get('/permissions/all')
}