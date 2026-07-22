import request from '@/utils/request'

export interface LoginReq {
  username: string
  password: string
  tenantCode?: string
  captchaId: string
  captchaCode: string
}

export interface UserInfo {
  id: number
  username: string
  realName: string
  tenantId: number
  isAdmin: boolean
  avatar?: string
}

export function login(data: LoginReq) {
  return request.post('/auth/login', data)
}

export function getCaptcha() {
  return request.get('/auth/captcha')
}

export function getUserInfo() {
  return request.get('/auth/info')
}

export function getUserMenus() {
  return request.get('/auth/menus')
}

export function getUserPermissions() {
  return request.get('/auth/permissions')
}

export function changePassword(data: { oldPwd: string; newPwd: string }) {
  return request.post('/auth/change-password', data)
}
