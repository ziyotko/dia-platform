import request from '@/utils/request'

export interface UserRoleRef {
  id: number
  code?: string
  name: string
}

export interface User {
  id: number
  tenantId: number
  username: string
  realName?: string
  phone?: string
  email?: string
  avatar?: string
  status: number
  isAdmin: boolean
  /** 所属机构 ID（0 表示未分配） */
  organizationId?: number
  roles?: UserRoleRef[]
}

export interface UserQuery {
  page: number
  size: number
  keyword?: string
  /** 平台超管可按租户过滤，不传表示全部租户 */
  tenantId?: number
}

export interface UserForm {
  username: string
  password?: string
  realName?: string
  phone?: string
  email?: string
  isAdmin: boolean
  status: number
  /** 所属机构 ID（0 表示未分配） */
  organizationId?: number
  /** 平台超管可为指定租户创建用户；普通租户用户由后端强制为自身租户 */
  tenantId?: number
}

export function getUserList(params: UserQuery) {
  return request.get('/users', { params })
}

export function createUser(data: UserForm) {
  return request.post('/users', data)
}

export function updateUser(id: number, data: Partial<User>) {
  return request.put(`/users/${id}`, data)
}

export function deleteUser(id: number) {
  return request.delete(`/users/${id}`)
}

export function assignUserRoles(id: number, roleIds: number[]) {
  return request.post(`/users/${id}/roles`, { roleIds })
}

export function resetUserPassword(id: number, password: string) {
  return request.post(`/users/${id}/reset-password`, { password })
}
