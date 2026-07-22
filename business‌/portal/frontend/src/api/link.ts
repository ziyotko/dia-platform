import request from '@/utils/request'

export interface LinkForm {
  id?: number
  name: string
  url: string
  logo?: string
  description?: string
  pageId: number
  columnId?: number
  sort: number
  status: number
}

export function getLinks(params: { name?: string; pageId?: number; columnId?: number; status?: number; page?: number; pageSize?: number }) {
  return request.get('/links', { params })
}

export function getLinkByID(id: number) {
  return request.get(`/links/${id}`)
}

export function createLink(data: LinkForm) {
  return request.post('/links', data)
}

export function updateLink(id: number, data: LinkForm) {
  return request.put(`/links/${id}`, data)
}

export function updateLinkStatus(id: number, status: number) {
  return request.patch(`/links/${id}/status`, { status })
}

export function deleteLink(id: number) {
  return request.delete(`/links/${id}`)
}
