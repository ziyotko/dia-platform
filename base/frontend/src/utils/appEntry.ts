/**
 * 子应用入口 URL 组装（一次性票据版）。
 *
 * 与 base/DEPLOY.md「五、子应用接入」的约定保持一致：不再把 access token（`base_token`）放进 URL，
 * 而是现场签一张 60 秒、一次性消费的票据（`base_ticket`）：
 *   - 避免长期凭证进入浏览器历史 / Referer / 网关访问日志；
 *   - 子应用启动时 POST /business_base/api/auth/app-ticket/exchange 用票据换回 token 与用户信息。
 * 签票失败时返回带 `base_ticket_error` 的地址，子应用可据此提示「请重新从底座进入」，而不是静默无会话。
 */
import { createAppTicket } from '@/api/auth'

export async function buildAppEntryUrl(base: string): Promise<string> {
  if (!base) return ''
  let ticket = ''
  let failed = false
  try {
    const res: any = await createAppTicket()
    ticket = res?.data?.ticket || ''
  } catch (error) {
    console.error('[base] 签发子应用接入票据失败', error)
    failed = true
  }
  const params = new URLSearchParams()
  if (ticket) {
    params.set('base_ticket', ticket)
  } else if (failed) {
    params.set('base_ticket_error', '1')
  }
  const query = params.toString()
  return query ? base + (base.includes('?') ? '&' : '?') + query : base
}
