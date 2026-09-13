import request from '@/utils/request'

export interface UserQuery {
  page?: number
  pageSize?: number
  username?: string
  account?: string
  status?: number
}

export interface UserForm {
  id?: number
  username: string
  account: string
  email: string
  phone: string
  status: number
  roleIds: number[]
  orgIds?: number[]
  sex?: number
}

export function getUserList(params: UserQuery) {
  return request.get('/users', { params })
}

/**
 * 获取全部用户（供下拉选项使用）。
 * 走后端 all=1 不分页，避免用户数超过 pageSize 上限时被静默截断。
 */
export function getAllUsers(status?: number) {
  return request.get('/users', { params: { all: 1, status } })
}

export function createUser(data: UserForm) {
  return request.post('/users', data)
}

export function updateUser(id: number, data: UserForm) {
  return request.put(`/users/${id}`, data)
}

export function deleteUser(id: number) {
  return request.delete(`/users/${id}`)
}

export function updateUserStatus(id: number, status: number) {
  return request.patch(`/users/${id}/status`, { status })
}

export function checkUserUnique(field: string, value: string, excludeId?: number) {
  return request.get('/users/check-unique', {
    params: { field, value, excludeId }
  })
}

export function importUsers(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/users/import', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}