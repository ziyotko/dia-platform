import request from '@/utils/request'

export interface App {
  id: number
  code: string
  name: string
  icon?: string
  type: 'iframe' | 'proxy' | 'micro'
  frontendUrl?: string
  backendUrl?: string
  apiPrefix?: string
  status: number
  sort: number
  description?: string
}

export function getAppList(params: { page: number; size: number; keyword?: string }) {
  return request.get('/apps', { params })
}

export function createApp(data: App) {
  return request.post('/apps', data)
}

export function updateApp(id: number, data: App) {
  return request.put(`/apps/${id}`, data)
}

export function deleteApp(id: number) {
  return request.delete(`/apps/${id}`)
}

export function getMyApps() {
  return request.get('/app-instances/my')
}

export function getAppInstances(params: { page: number; size: number; tenantId?: number }) {
  return request.get('/app-instances', { params })
}

export function createAppInstance(data: { tenantId: number; appId: number; status: number; config?: string }) {
  return request.post('/app-instances', data)
}

export function updateAppInstance(id: number, data: { status: number; config?: string }) {
  return request.put(`/app-instances/${id}`, data)
}

export function deleteAppInstance(id: number) {
  return request.delete(`/app-instances/${id}`)
}
