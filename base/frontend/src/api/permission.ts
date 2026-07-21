import request from '@/utils/request'

export interface Permission {
  id: number
  appCode: string
  code: string
  name: string
  type: 'menu' | 'api' | 'button'
  parentId: number
  path?: string
  method?: string
  status: number
  children?: Permission[]
}

export function getPermissionTree(appCode?: string) {
  return request.get('/permissions/tree', { params: { appCode } })
}

export function createPermission(data: Permission) {
  return request.post('/permissions', data)
}

export function updatePermission(id: number, data: Permission) {
  return request.put(`/permissions/${id}`, data)
}

export function deletePermission(id: number) {
  return request.delete(`/permissions/${id}`)
}
