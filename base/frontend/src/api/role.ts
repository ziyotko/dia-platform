import request from '@/utils/request'

export interface Role {
  id: number
  code: string
  name: string
  status: number
  remark?: string
}

export function getRoleList(params: { page: number; size: number; keyword?: string }) {
  return request.get('/roles', { params })
}

export function createRole(data: Role) {
  return request.post('/roles', data)
}

export function updateRole(id: number, data: Role) {
  return request.put(`/roles/${id}`, data)
}

export function deleteRole(id: number) {
  return request.delete(`/roles/${id}`)
}

export function assignRoleMenus(id: number, menuIds: number[]) {
  return request.post(`/roles/${id}/menus`, { menuIds })
}

export function assignRolePermissions(id: number, permissionIds: number[]) {
  return request.post(`/roles/${id}/permissions`, { permissionIds })
}
