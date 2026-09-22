/**
 * 后端下发的资源地址（上传文件、站点 Logo 等）→ 可直接用于 <img src> 的地址。
 *
 * - 空值 / undefined → ''
 * - 已是绝对地址（http/https/data/blob）→ 原样返回
 * - 其余（通常形如 `/business_portal/uploads/xxx.jpg`）→ 补当前站点 origin
 *
 * 2026-09-23 统一：此前 login / layout / settings 三处各写了一份实现（login 还是恒等返回），
 * 同一字段在不同页面会解析出不同地址（后端若改为下发相对路径，只有登录页会 404）。
 * 现收敛到本工具，新增页面一律 `import { resolveAssetUrl } from '@/utils/asset'`。
 */
export function resolveAssetUrl(url?: string | null): string {
  if (!url) return ''
  if (/^(https?:|data:|blob:)/i.test(url)) return url
  return `${window.location.origin}${url.startsWith('/') ? url : `/${url}`}`
}
