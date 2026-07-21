import request from '@/utils/request'

export interface Organization {
  id: number
  parentId: number
  code: string
  name: string
  leader?: string
  phone?: string
  email?: string
  sort: number
  status: number
  description?: string
  children?: Organization[]
}

export function getOrganizationTree() {
  return request.get('/organizations/tree')
}

export function createOrganization(data: Organization) {
  return request.post('/organizations', data)
}

export function updateOrganization(id: number, data: Organization) {
  return request.put(`/organizations/${id}`, data)
}

export function deleteOrganization(id: number) {
  return request.delete(`/organizations/${id}`)
}
