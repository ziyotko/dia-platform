import request from '@/utils/request'

export interface MenuItem {
  id: number
  parentId: number
  name: string
  path: string
  component: string
  icon: string
  type: string
  sort: number
  status: number
  children?: MenuItem[]
  createdAt?: string
  updatedAt?: string
}

export interface MenuForm {
  id?: number
  parentId?: number
  name: string
  path: string
  component: string
  icon: string
  type: string
  sort: number
  status: number
}

export function getMenuList() {
  return request.get('/menus')
}

export function getMenuTree() {
  return request.get('/menus/tree')
}

export function getUserMenus() {
  return request.get('/menus/user')
}

export function createMenu(data: MenuForm) {
  return request.post('/menus', data)
}

export function updateMenu(id: number, data: MenuForm) {
  return request.put(`/menus/${id}`, data)
}

export function deleteMenu(id: number) {
  return request.delete(`/menus/${id}`)
}
