import request from '@/utils/request'

export interface CategoryForm {
  id?: number
  name: string
  code: string
  description?: string
  sort: number
  status: number
}

export function getCategories(params: { name?: string; status?: number; page?: number; pageSize?: number }) {
  return request.get('/categories', { params })
}

export function getAllCategories() {
  return request.get('/categories/all')
}

export function createCategory(data: CategoryForm) {
  return request.post('/categories', data)
}

export function updateCategory(id: number, data: CategoryForm) {
  return request.put(`/categories/${id}`, data)
}

export function updateCategoryStatus(id: number, status: number) {
  return request.patch(`/categories/${id}/status`, { status })
}

export function deleteCategory(id: number) {
  return request.delete(`/categories/${id}`)
}
