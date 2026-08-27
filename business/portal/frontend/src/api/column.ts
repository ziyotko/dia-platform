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
  displayType: number
  workflowId?: number
}

export function getColumns(params?: { pageId?: number; parentId?: number; displayType?: number }) {
  return request.get('/columns', { params })
}

export function getColumnPublishes() {
  return request.get('/columns/publishes')
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
