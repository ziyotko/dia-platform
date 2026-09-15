export const applicationStatusMap: Record<string, string> = {
  draft: '草稿',
  submitted: '待初审',
  preliminary_rejected: '初审驳回',
  under_review: '待评审',
  reviewed: '评审完成',
  passed: '已通过',
  rejected: '未通过',
  published: '已公示',
  certified: '已发证',
}

export const applicationStatusType: Record<string, string> = {
  draft: 'info',
  submitted: 'warning',
  preliminary_rejected: 'danger',
  under_review: 'warning',
  reviewed: 'primary',
  passed: 'success',
  rejected: 'danger',
  published: 'success',
  certified: 'success',
}

export const batchStatusMap: Record<string, string> = {
  draft: '草稿',
  open: '申报中',
  reviewing: '评审中',
  closed: '已结束',
}

export const batchStatusType: Record<string, string> = {
  draft: 'info',
  open: 'success',
  reviewing: 'warning',
  closed: 'info',
}

export const reviewStatusMap: Record<string, string> = {
  pending: '待评审',
  scored: '已评分',
}

export const certStatusMap: Record<string, string> = {
  draft: '未颁发',
  issued: '已颁发',
  void: '已作废',
}

export const certStatusType: Record<string, string> = {
  draft: 'info',
  issued: 'success',
  void: 'danger',
}

// 列表里的「平均分」：avgScore 在后端默认是 0，无评分时必须显示为未评分而不是
// 0 分（历史问题：0 分被 `avgScore || '-'` 显示成 -）。scoredCount 由后端
// 列表接口统一填充。
export function avgScoreText(row: any): string {
  if (!row) return '-'
  const scored = Number(row.scoredCount ?? 0)
  if (!scored || row.avgScore === null || row.avgScore === undefined) return '-'
  return String(row.avgScore)
}

export function fileUrl(path?: string): string {
  if (!path) return ''
  if (path.startsWith('http://') || path.startsWith('https://')) return path
  return path.startsWith('/') ? path : `/${path}`
}

export function fmt(d?: string): string {
  if (!d) return '-'
  return d.replace('T', ' ').slice(0, 16)
}
