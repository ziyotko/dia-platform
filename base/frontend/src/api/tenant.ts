import request from '@/utils/request'

export interface Tenant {
  id: number
  code: string
  name: string
  status: number
  contactName?: string
  contactPhone?: string
  description?: string
}

export function getTenantList(params: { page: number; size: number; keyword?: string }) {
  return request.get('/tenants', { params })
}

export function createTenant(data: Tenant) {
  return request.post('/tenants', data)
}

export function updateTenant(id: number, data: Tenant) {
  return request.put(`/tenants/${id}`, data)
}

export function deleteTenant(id: number) {
  return request.delete(`/tenants/${id}`)
}
