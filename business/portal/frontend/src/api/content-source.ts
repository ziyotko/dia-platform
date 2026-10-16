import request from '@/utils/request'

export type ContentSourceProvider = 'wechat_official'
export type ContentSourceAccountType = 'service' | 'subscription'
export type ContentSourceSyncMode = 'published' | 'materials'
export type ContentSyncStatus = 'queued' | 'running' | 'success' | 'failed'

export interface ContentSource {
  id?: number
  name: string
  provider: ContentSourceProvider
  accountType: ContentSourceAccountType
  appId: string
  secretEnv: string
  secretConfigured?: boolean
  syncMode: ContentSourceSyncMode
  status: number
  defaultCategoryId?: number
  lastSyncAt?: string
  lastSyncStatus?: ContentSyncStatus | ''
  lastSyncMessage?: string
  createdAt?: string
}

export interface ContentSyncLog {
  id: number
  sourceId: number
  operator: string
  status: ContentSyncStatus
  startedAt?: string
  finishedAt?: string
  scannedCount: number
  createdCount: number
  skippedCount: number
  errorMessage?: string
  createdAt: string
}

export function getContentSources(params: {
  name?: string
  status?: number
  page?: number
  pageSize?: number
}) {
  return request.get('/content-sources', { params })
}

export function createContentSource(data: ContentSource) {
  return request.post('/content-sources', data)
}

export function updateContentSource(id: number, data: ContentSource) {
  return request.put(`/content-sources/${id}`, data)
}

export function deleteContentSource(id: number) {
  return request.delete(`/content-sources/${id}`)
}

export function testContentSource(id: number) {
  return request.post(`/content-sources/${id}/test`)
}

export function startContentSourceSync(id: number) {
  return request.post(`/content-sources/${id}/sync`)
}

export function getContentSourceLogs(id: number, params: { page?: number; pageSize?: number }) {
  return request.get(`/content-sources/${id}/logs`, { params })
}

export function getContentSyncLog(id: number) {
  return request.get(`/content-sync-logs/${id}`)
}
