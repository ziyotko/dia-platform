import request from '@/utils/request'

export interface TagForm {
  id?: number
  name: string
  color: string
  status: number
}

export function getTags(params: { name?: string; status?: number; page?: number; pageSize?: number }) {
  return request.get('/tags', { params })
}

export function getAllTags() {
  return request.get('/tags/all')
}

export function getTagArticleStats() {
  return request.get('/tags/stats')
}

export function createTag(data: TagForm) {
  return request.post('/tags', data)
}

export function updateTag(id: number, data: TagForm) {
  return request.put(`/tags/${id}`, data)
}

export function updateTagStatus(id: number, status: number) {
  return request.patch(`/tags/${id}/status`, { status })
}

export function deleteTag(id: number) {
  return request.delete(`/tags/${id}`)
}
