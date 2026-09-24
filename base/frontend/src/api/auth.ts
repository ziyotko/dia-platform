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

export interface LoginResp {
  token: string
  refresh_token: string
  expires_in: number
}

export function login(data: LoginReq) {
  return request.post<LoginResp>('/auth/login', data)
}

/** 用 refresh token 换发新的 access + refresh（轮换：旧 refresh 立即失效） */
export function refreshSession(refreshToken: string) {
  return request.post<LoginResp>('/auth/refresh', { refresh_token: refreshToken })
}

/** 为自己签发子应用一次性接入票据（60 秒有效、只用一次） */
export function createAppTicket() {
  return request.post<{ ticket: string; expires_in: number }>('/auth/app-ticket')
}

/** 子应用侧：用一次性票据换回底座会话（子应用调用） */
export function exchangeAppTicket(ticket: string) {
  return request.post('/auth/app-ticket/exchange', { ticket })
}

export function getCaptcha() {
  return request.get<{ captcha_id: string; captcha_img: string }>('/auth/captcha')
}

/** 公开站点信息（无需登录）：登录页据此决定是否展示验证码，并展示平台名称/Logo/版权 */
export interface SiteInfo {
  captchaEnabled: boolean
  platformName?: string
  logo?: string
  copyright?: string
}

export function getSiteInfo() {
  return request.get<SiteInfo>('/site-info')
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

/** 登出：服务端把当前 access token 加入黑名单并作废 refresh token（失败也不影响前端清理本地会话）。
 * token / refresh_token 显式传入：调用方会紧接着清空 store，不能依赖请求拦截器再去读取。 */
export function logout(token?: string, refreshToken?: string) {
  const config = token ? { headers: { Authorization: `Bearer ${token}` } } : undefined
  return request.post('/auth/logout', { refresh_token: refreshToken || '' }, config)
}
