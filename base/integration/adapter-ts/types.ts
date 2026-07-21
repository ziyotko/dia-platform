export interface BaseUser {
  userId: number
  username: string
  tenantId: number
  realName?: string
}

export interface BaseAdapterOptions {
  // 底座平台域名
  baseUrl: string
  // 当前子应用编码
  appCode: string
  // 登录页地址
  loginUrl: string
}

/**
 * 从 URL 解析底座平台传递的 token
 */
export function getBaseToken(): string | null {
  const params = new URLSearchParams(window.location.search)
  return params.get('base_token')
}

/**
 * 保存底座 token 到子应用本地
 */
export function setBaseToken(token: string): void {
  localStorage.setItem('base_token', token)
}

/**
 * 获取当前登录的底座用户信息
 */
export function getBaseUser(): BaseUser | null {
  const raw = localStorage.getItem('base_user')
  if (!raw) return null
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

export function setBaseUser(user: BaseUser): void {
  localStorage.setItem('base_user', JSON.stringify(user))
}
