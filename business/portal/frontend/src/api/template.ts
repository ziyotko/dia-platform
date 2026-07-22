import request from '@/utils/request'

export interface TemplateQuery {
  page?: number
  pageSize?: number
  name?: string
  type?: string
}

export interface TemplateForm {
  id?: number
  name: string
  code: string
  type: string
  description?: string
  status: number
}

export interface DesignForm {
  sourceCode: string
  layout: string
}

export function getTemplateList(params: TemplateQuery) {
  return request.get('/templates', { params })
}

export function createTemplate(data: TemplateForm) {
  return request.post('/templates', data)
}

export function updateTemplate(id: number, data: TemplateForm) {
  return request.put(`/templates/${id}`, data)
}

export function deleteTemplate(id: number) {
  return request.delete(`/templates/${id}`)
}

export function updateTemplateStatus(id: number, status: number) {
  return request.patch(`/templates/${id}/status`, { status })
}

export function saveTemplateDesign(id: number, data: DesignForm) {
  return request.put(`/templates/${id}/design`, data)
}
