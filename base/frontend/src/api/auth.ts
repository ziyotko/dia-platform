import request from '@/utils/request'

export interface LoginReq {
  username: string
  password: string
  tenantCode?: string
  /** 验证码字段名与 portal / member / application 保持一致 */
  captcha_id: string
  captcha_code: string
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
  return request.get<{ captcha_id: string; captcha_img: string }>('/auth/captcha')
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
