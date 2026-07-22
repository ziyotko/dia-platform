import request from '@/utils/request'

export interface DictItem {
  id?: number
  dictId?: number
  label: string
  value: string
  sort: number
  status: number
}

export interface Dict {
  id?: number
  code: string
  name: string
  description: string
  status: number
  items?: DictItem[]
}

export interface DictQuery {
  page: number
  size: number
  code?: string
  name?: string
  status?: number
}

export function getDictList(params: DictQuery) {
  return request.get('/dicts', { params })
}

export function getDictByCode(code: string) {
  return request.get(`/dicts/code/${code}`)
}

export function createDict(data: Dict) {
  return request.post('/dicts', data)
}

export function updateDict(id: number, data: Dict) {
  return request.put(`/dicts/${id}`, data)
}

export function deleteDict(id: number) {
  return request.delete(`/dicts/${id}`)
}

export function getDictDetail(id: number) {
  return request.get(`/dicts/${id}`)
}

export function saveDictItems(id: number, items: DictItem[]) {
  return request.post(`/dicts/${id}/items`, { items })
}
