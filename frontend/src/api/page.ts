import request from '@/utils/request'

export interface PageForm {
  id?: number
  name: string
  code: string
  pageType: string
  routePath: string
  templateId?: number
  template?: string
  description?: string
  status: number
}

export function getPages(params?: { pageType?: string; templateId?: number }) {
  return request.get('/pages', { params })
}

export function createPage(data: PageForm) {
  return request.post('/pages', data)
}

export function updatePage(id: number, data: PageForm) {
  return request.put(`/pages/${id}`, data)
}

export function deletePage(id: number) {
  return request.delete(`/pages/${id}`)
}
