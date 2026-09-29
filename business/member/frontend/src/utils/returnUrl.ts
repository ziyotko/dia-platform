import type { Router } from 'vue-router'

/**
 * 登录回跳（returnUrl）工具。
 *
 * 跳登录页时用 `?returnUrl=<encodeURIComponent(目标地址)>` 带上原地址，
 * 登录（或注册）成功后再跳回去。三处入口都已接上：
 *  - 路由守卫拦截未登录访问 → 自动带原地址（`router/index.ts`）；
 *  - 登录态失效（401）踢回登录页 → 带上当前页面（`stores/user.ts` 的 `forceLogout`）；
 *  - 外部系统（如 portal 会员专区）可自行拼 `…/business_member/login?returnUrl=…`。
 *
 * 安全口径：只接受本站地址，其余一律忽略，避免登录页变成开放重定向跳板。
 *  - 允许：同源 http(s) 绝对地址（如 `https://host/business_portal/member-zone`）、
 *          以 `/` 开头的站内路径（带不带部署子路径都行：`/member/fees`、`/business_member/member/fees`）；
 *  - 拒绝：`//evil.com`、`javascript:` / `data:` 等伪协议、跨域地址、含控制字符或 `\` 的值
 *          （浏览器把 `\` 当 `/`，`/\evil.com` 等价于 `//evil.com`）、回到登录页自身（防死循环）。
 */

/** 回跳参数名 */
export const RETURN_URL_PARAM = 'returnUrl'

/** 兼容取值的其它参数名（外部系统常写 redirect） */
const COMPAT_RETURN_URL_PARAMS = ['redirect']

/** 部署子路径（vite base，如 `/business_member/`；根路径部署为 `/`） */
const BASE_PATH = (import.meta.env.BASE_URL || '/').replace(/\/+$/, '')

/** 跳转目标：站内路由 path（交给 vue-router，自动拼部署子路径）或同源完整地址 */
export type ReturnTarget = { kind: 'route'; path: string } | { kind: 'url'; url: string }

/** 从路由 query 里取回跳地址（`returnUrl` 优先，兼容 `redirect`） */
export function pickReturnUrl(query: unknown): string {
  if (!query || typeof query !== 'object') return ''
  const q = query as Record<string, unknown>
  for (const name of [RETURN_URL_PARAM, ...COMPAT_RETURN_URL_PARAMS]) {
    const value = q[name]
    const first = Array.isArray(value) ? value[0] : value
    if (typeof first === 'string' && first.trim()) return first
  }
  return ''
}

/** 去掉部署子路径，得到 vue-router 的 path */
function stripBasePath(path: string): string {
  if (!BASE_PATH) return path
  if (path === BASE_PATH) return '/'
  if (path.startsWith(`${BASE_PATH}/`)) return path.slice(BASE_PATH.length)
  return path
}

/** 是否是登录页（忽略 query/hash 与结尾斜杠） */
function isLoginPath(path: string): boolean {
  const pure = (path || '').replace(/[?#].*$/, '').replace(/\/+$/, '')
  return pure === '/login' || pure === `${BASE_PATH}/login`
}

/** 是否含危险字符：反斜杠（浏览器当 `/`，`/\evil.com` 等价 `//evil.com`）与控制字符 */
function hasUnsafeChars(value: string): boolean {
  for (const ch of value) {
    const code = ch.codePointAt(0) ?? 0
    if (ch === '\\' || code < 0x20 || code === 0x7f) return true
  }
  return false
}

/** 解析回跳地址；为空或非法返回 null */
export function resolveReturnUrl(raw?: string | null): ReturnTarget | null {
  const value = (raw || '').trim()
  if (!value) return null
  if (hasUnsafeChars(value)) return null

  // 绝对地址：只放行同源 http(s)，整页跳转（本站子路径也走整页，避免二次解析出错）
  if (/^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(value)) {
    let url: URL
    try {
      url = new URL(value)
    } catch {
      return null
    }
    if (url.protocol !== 'http:' && url.protocol !== 'https:') return null
    if (url.origin !== window.location.origin) return null
    return { kind: 'url', url: url.href }
  }

  // 站内路径：必须以单个 `/` 开头（`//host` 是协议相对地址）
  if (!value.startsWith('/') || value.startsWith('//')) return null
  const path = stripBasePath(value)
  if (!path.startsWith('/') || path.startsWith('//')) return null
  if (isLoginPath(path)) return null
  return { kind: 'route', path }
}

/**
 * 登录/注册成功后的统一跳转：有合法 returnUrl 就跳回去，否则去 fallback
 * （调用方用 `userStore.isAdmin ? '/admin/dashboard' : '/member/dashboard'`）。
 */
export function goAfterLogin(router: Router, raw: string | null | undefined, fallback: string): void {
  const target = resolveReturnUrl(raw)
  if (!target) {
    router.replace(fallback)
    return
  }
  if (target.kind === 'url') {
    window.location.replace(target.url)
    return
  }
  router.replace(target.path)
}

/** 路由守卫用：未登录时跳到登录页并带上原地址 */
export function loginRouteLocation(fullPath: string): { path: string; query?: Record<string, string> } {
  if (isLoginPath(fullPath)) return { path: '/login' }
  return { path: '/login', query: { [RETURN_URL_PARAM]: fullPath } }
}

/** 当前页面地址（站内 path + query，不含部署子路径）；已在登录页时返回空串 */
export function currentReturnUrl(): string {
  const path = stripBasePath(window.location.pathname)
  if (isLoginPath(path)) return ''
  return path + window.location.search
}

/** 整页跳登录页用的地址，带部署子路径，如 `/business_member/login?returnUrl=%2Fmember%2Ffees` */
export function loginPathWithReturnUrl(fullPath: string): string {
  const base = `${BASE_PATH}/`
  const query = fullPath && !isLoginPath(fullPath) ? `?${RETURN_URL_PARAM}=${encodeURIComponent(fullPath)}` : ''
  return `${base}login${query}`
}
