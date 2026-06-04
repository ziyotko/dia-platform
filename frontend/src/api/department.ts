import request from '@/utils/request'

export interface DepartmentQuery {
  page?: number
  pageSize?: number
  name?: string
  status?: number
}

export interface DepartmentForm {
  id?: number
  name: string
  code: string
  parentId: number
  leader: string
  leaderCode: string
  description: string
  status: number
  sort: number
  userIds?: number[]
}

export function getDepartmentList(params?: DepartmentQuery) {
  return request.get('/departments', { params })
}

export function getDepartmentTree() {
  return request.get('/departments/tree')
}

export function getDepartmentUsers(id: number) {
  return request.get(`/departments/${id}/users`)
}

export function createDepartment(data: DepartmentForm) {
  return request.post('/departments', data)
}

export function updateDepartment(id: number, data: DepartmentForm) {
  return request.put(`/departments/${id}`, data)
}

export function deleteDepartment(id: number) {
  return request.delete(`/departments/${id}`)
}

export function assignDepartmentUsers(id: number, userIds: number[]) {
  return request.put(`/departments/${id}/users`, { userIds })
}
