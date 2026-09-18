/**
 * 把数据库/接口返回的上传文件路径统一解析为「当前部署子路径」下的可访问 URL。
 *
 * 兼容历史上出现过的四种形态：
 *   - uploads/...                    相对路径（新接口返回的形式）
 *   - /uploads/...                   旧绝对路径
 *   - \\uploads\\... 或 /uploads\\...  旧数据里的反斜杠路径（Windows filepath 遗留）
 *   - /business_member/uploads/...   带部署前缀的绝对路径
 *
 * 说明：后端静态挂载为 `<server.upload_dir_prefix>/uploads`，
 * 因此旧的 `/uploads/...` 与反斜杠路径必须重新指到当前部署子路径下。
 */
export function fileUrl(path?: string | null): string {
  if (!path) return ''
  const raw = String(path).trim().replace(/\\/g, '/')
  if (/^https?:\/\//i.test(raw)) return raw
  const base = import.meta.env.BASE_URL || '/'
  const clean = raw.replace(/^\.?\//, '')
  return clean.startsWith('uploads/') ? `${base}${clean}` : `/${clean}`
}
