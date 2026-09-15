import request from '@/utils/request'

export interface Menu {
  id: number
  parentId: number
  appCode: string
  name: string
  icon?: string
  path: string
  component?: string
  type: 'directory' | 'menu' | 'button'
  permission?: string
  sort: number
  status: number
  hidden: boolean
  /** 预留：路由缓存暂未实现，后端字段保留 */
  keepAlive?: boolean
  target?: string
  children?: Menu[]
}

export function getMenuTree(appCode?: string) {
  return request.get('/menus/tree', { params: { appCode } })
}

export function createMenu(data: Menu) {
  return request.post('/menus', data)
}

export function updateMenu(id: number, data: Menu) {
  return request.put(`/menus/${id}`, data)
}

export function deleteMenu(id: number) {
  return request.delete(`/menus/${id}`)
}
