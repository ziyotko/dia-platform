import request from '@/utils/request'

export interface OrgQuery {
  name?: string
  orgType?: number
  status?: number
}

export interface OrgForm {
  id?: number
  parentId?: number
  name: string
  code: string
  orgType?: number
  orgLevel: number
  category: string
  region: string
  province: string
  city: string
  address: string
  manager: string
  managerCode: string
  sort: number
  status: number
  description: string
  userIds?: number[]
}

export interface DepartmentItem {
  id: number
  parentId: number
  orgId: number
  name: string
  code: string
  leader: string
  leaderCode: string
  sort: number
  status: number
  description: string
  userCount: number
  createdAt: string
  children?: DepartmentItem[]
  hasChildren?: boolean
}

export interface OrgItem {
  id: number
  parentId: number
  name: string
  code: string
  orgType: number
  orgTypeText: string
  orgLevel: number
  category: string
  region: string
  province: string
  city: string
  address: string
  manager: string
  managerCode: string
  sort: number
  status: number
  description: string
  userCount: number
  createdAt: string
  children?: OrgItem[]
  hasChildren?: boolean
  departments?: DepartmentItem[]
}

export function getOrgList(params?: OrgQuery) {
  return request.get('/organizations', { params })
}

export function getOrgTree() {
  return request.get('/organizations/tree')
}

export function getOrgUsers(id: number) {
  return request.get(`/organizations/${id}/users`)
}

export function createOrg(data: OrgForm) {
  return request.post('/organizations', data)
}

export function updateOrg(id: number, data: OrgForm) {
  return request.put(`/organizations/${id}`, data)
}

export function deleteOrg(id: number) {
  return request.delete(`/organizations/${id}`)
}

export function assignOrgUsers(id: number, userIds: number[]) {
  return request.put(`/organizations/${id}/users`, { userIds })
}
