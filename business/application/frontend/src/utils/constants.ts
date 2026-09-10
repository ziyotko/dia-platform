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
  published: '已公示',
  closed: '已结束',
}

export const batchStatusType: Record<string, string> = {
  draft: 'info',
  open: 'success',
  reviewing: 'warning',
  published: 'primary',
  closed: 'info',
}

export const reviewStatusMap: Record<string, string> = {
  pending: '待评审',
  scored: '已评分',
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
