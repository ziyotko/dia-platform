import request from '@/utils/request'

export interface Role {
  id: number
  tenantId?: number
  code: string
  name: string
  status: number
  remark?: string
  menus?: { id: number }[]
  permissions?: { id: number }[]
}

export interface RoleQuery {
  page: number
  size: number
  keyword?: string
  /** 平台超管可按租户过滤，不传表示全部租户 */
  tenantId?: number
}

export function getRoleList(params: RoleQuery) {
  return request.get('/roles', { params })
}

/** 角色详情（含已分配菜单与权限） */
export function getRole(id: number) {
  return request.get(`/roles/${id}`)
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
