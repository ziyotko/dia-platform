import request from '@/utils/request'

export interface ArticleForm {
  id?: number
  title: string
  categoryId: number
  tagIds: number[]
  summary: string
  content: string
  status: number
  isTop: number
  cover: string
  source: string
}

export function getArticles(params: { title?: string; categoryId?: number; status?: number; page?: number; pageSize?: number }) {
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

export function deleteArticle(id: number) {
  return request.delete(`/articles/${id}`)
}
