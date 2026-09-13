/**
 * 将后端返回的时间（ISO 字符串 / 时间戳 / Date）格式化为 `YYYY-MM-DD HH:mm:ss`。
 * 空值或非法时间返回空字符串。
 */
export function formatDateTime(value?: string | number | Date | null): string {
  if (!value) return ''
  const d = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(d.getTime())) return ''
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(
    d.getHours()
  )}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}
