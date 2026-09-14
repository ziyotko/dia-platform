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

/** 日志保留策略：系统保留最近半年日志，「清空日志」仅删除该日期之前的记录 */
export const LOG_RETENTION_MONTHS = 6

/** 「清空日志」截止日期（YYYY-MM-DD）：该日期之前的日志可被清空（三个日志页面统一） */
export function getLogClearCutoff(): string {
  const now = new Date()
  const cutoff = new Date(now.getFullYear(), now.getMonth() - LOG_RETENTION_MONTHS, now.getDate())
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${cutoff.getFullYear()}-${pad(cutoff.getMonth() + 1)}-${pad(cutoff.getDate())}`
}

/** 「清空日志」确认文案（操作日志/登录日志/静态化日志共用，避免语义不一致） */
export function buildClearLogsConfirmText(): string {
  const cutoff = getLogClearCutoff()
  return `仅能清空 ${cutoff}（半年前）之前的日志，系统将保留 ${cutoff} 至今的最近半年日志。确定继续吗？此操作不可恢复！`
}

/** 「清空日志」成功提示（count 为实际删除条数） */
export function buildClearLogsSuccessText(count: number): string {
  const cutoff = getLogClearCutoff()
  return count > 0
    ? `已清空 ${count} 条 ${cutoff} 之前的日志，近半年日志已保留`
    : `暂无 ${cutoff} 之前的日志可清空，近半年日志已保留`
}
