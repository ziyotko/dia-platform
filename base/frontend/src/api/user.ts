import request from '@/utils/request'

export interface User {
  id: number
  username: string
  realName?: string
  phone?: string
  email?: string
  avatar?: string
  status: number
  isAdmin: boolean
}

export function getUserList(params: { page: number; size: number; keyword?: string }) {
  return request.get('/users', { params })
}

export function createUser(data: User) {
  return request.post('/users', data)
}

export function updateUser(id: number, data: User) {
  return request.put(`/users/${id}`, data)
}

export function deleteUser(id: number) {
  return request.delete(`/users/${id}`)
}

export function assignUserRoles(id: number, roleIds: number[]) {
  return request.post(`/users/${id}/roles`, { roleIds })
}

export function resetUserPassword(id: number) {
  return request.post(`/users/${id}/reset-password`)
}
