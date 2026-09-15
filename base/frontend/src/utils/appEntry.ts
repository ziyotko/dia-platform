/**
 * 子应用入口 URL 组装。
 *
 * 与 integration/README.md 的约定保持一致：无论从菜单（iframe 容器）还是「我的应用」卡片进入，
 * 都要带上底座会话参数，子应用才能用 base_token 调 /base/api/v1/auth/info 校验用户身份。
 */
export interface BaseSession {
  token?: string
  userId?: number | string
  username?: string
  tenantId?: number
}

/** 在入口地址后追加底座会话参数（已有 query 时用 & 追加） */
export function buildAppEntryUrl(base: string, session: BaseSession): string {
  if (!base) return ''
  const params = new URLSearchParams({
    base_token: session.token || '',
    user_id: String(session.userId ?? ''),
    username: session.username || '',
    tenant_id: String(session.tenantId ?? 0)
  })
  return base + (base.includes('?') ? '&' : '?') + params.toString()
}
