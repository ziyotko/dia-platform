import request from '@/utils/request'

export interface ColumnForm {
  id?: number
  name: string
  code: string
  pageId: number
  parentId?: number
  routePath?: string
  template?: string
  description?: string
  sort: number
  status: number
}

export function getColumns(params?: { pageId?: number; parentId?: number }) {
  return request.get('/columns', { params })
}

export function createColumn(data: ColumnForm) {
  return request.post('/columns', data)
}

export function updateColumn(id: number, data: ColumnForm) {
  return request.put(`/columns/${id}`, data)
}

export function deleteColumn(id: number) {
  return request.delete(`/columns/${id}`)
}
