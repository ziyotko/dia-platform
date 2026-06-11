import request from '@/utils/request'

export interface ArticleForm {
  id?: number
  title: string
  categoryId: number
  tagIds: number[]
  summary: string
  content: string
  status: number
  auditStatus?: number
  isTop: number
  isBold: number
  defaultColor: string
  cover: string
  source: string
}

export function getArticles(params: { title?: string; categoryId?: number; status?: number; auditStatus?: number; page?: number; pageSize?: number }) {
  return request.get('/articles', { params })
}

export function getArticleByID(id: number) {
  return request.get(`/articles/${id}`)
}

export function createArticle(data: ArticleForm) {
  return request.post('/articles', data)
}

export function updateArticle(id: number, data: ArticleForm) {
  return request.put(`/articles/${id}`, data)
}

export function updateArticleStatus(id: number, status: number) {
  return request.patch(`/articles/${id}/status`, { status })
}

export function auditArticle(id: number, auditStatus: number) {
  return request.patch(`/articles/${id}/audit`, { auditStatus })
}

export function restartArticleAudit(id: number) {
  return request.post(`/articles/${id}/audit-restart`)
}

export function withdrawArticleAudit(id: number) {
  return request.post(`/articles/${id}/audit-withdraw`)
}

export function getArticleAuditProgress(id: number) {
  return request.get(`/articles/${id}/audit-progress`)
}

export function advanceArticleAudit(id: number, columnId: number, remark: string) {
  return request.post(`/articles/${id}/audit-advance`, { columnId, remark })
}

export function rejectArticleAudit(id: number, columnId: number, remark: string) {
  return request.post(`/articles/${id}/audit-reject`, { columnId, remark })
}

export function getArticleAuditHistory(id: number, columnId: number) {
  return request.get(`/articles/${id}/audit-history`, { params: { columnId } })
}

export function setArticleColumns(id: number, columnIds: number[]) {
  return request.put(`/articles/${id}/columns`, { columnIds })
}

export function deleteArticle(id: number) {
  return request.delete(`/articles/${id}`)
}

export function getArticleColumnPublishes(params: { articleTitle?: string; columnId?: number; page?: number; pageSize?: number }) {
  return request.get('/articles/column-publishes', { params })
}
