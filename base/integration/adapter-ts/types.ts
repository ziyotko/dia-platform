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

export interface BaseSession {
  token: string
  user: BaseUser
}

/**
 * 从 URL 解析底座平台传递的一次性票据（推荐流程）。
 * 底座只会在地址后追加 `base_ticket`（60 秒、一次性），不再追加长期 token。
 */
export function getBaseTicket(): string | null {
  const params = new URLSearchParams(window.location.search)
  return params.get('base_ticket')
}

/**
 * 用一次性票据换回底座会话（推荐入口）：
 *
 *   const session = await resolveBaseSession('http://127.0.0.1:8091/business_base/api')
 *   // session.token 可直接用于调底座接口（如 GET /auth/info）
 *
 * 换完会把 `base_ticket` 从地址栏抹掉（避免留在浏览器历史里）。
 * 若地址里没有票据，会回退兼容旧流程的 `base_token`（已废弃）。
 */
export async function resolveBaseSession(baseUrl: string): Promise<BaseSession | null> {
  const params = new URLSearchParams(window.location.search)
  const ticket = params.get('base_ticket')
  if (ticket) {
    try {
      const resp = await fetch(`${baseUrl}/auth/app-ticket/exchange`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ticket })
      })
      const body = await resp.json()
      if (body?.code !== 0 || !body?.data?.token) {
        return null
      }
      const session: BaseSession = {
        token: body.data.token,
        user: {
          userId: body.data.user?.id,
          username: body.data.user?.username,
          tenantId: body.data.user?.tenantId,
          realName: body.data.user?.realName
        }
      }
      setBaseToken(session.token)
      setBaseUser(session.user)
      params.delete('base_ticket')
      const query = params.toString()
      window.history.replaceState({}, '', window.location.pathname + (query ? `?${query}` : '') + window.location.hash)
      return session
    } catch (error) {
      console.error('[base-adapter] 票据兑换失败', error)
      return null
    }
  }

  // 兼容：旧流程直接把 token 放在 URL（base_token，已废弃）
  const legacy = getBaseToken()
  if (legacy) {
    setBaseToken(legacy)
    const user = getBaseUser()
    if (user) return { token: legacy, user }
  }
  return null
}

/**
 * 从 URL 解析底座平台传递的 token（旧流程，已废弃：底座不再生成 base_token，仅作兼容读取）
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
