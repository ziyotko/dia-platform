import request from '@/utils/request'

export interface AdForm {
  id?: number
  name: string
  pageId: number
  columnId?: number
  image?: string
  link?: string
  sort: number
  status: number
  startTime?: string
  endTime?: string
}

export function getAds(params: { name?: string; pageId?: number; columnId?: number; status?: number; page?: number; pageSize?: number }) {
  return request.get('/ads', { params })
}

export function getAdByID(id: number) {
  return request.get(`/ads/${id}`)
}

export function createAd(data: AdForm) {
  return request.post('/ads', data)
}

export function updateAd(id: number, data: AdForm) {
  return request.put(`/ads/${id}`, data)
}

export function updateAdStatus(id: number, status: number) {
  return request.patch(`/ads/${id}/status`, { status })
}

export function deleteAd(id: number) {
  return request.delete(`/ads/${id}`)
}
