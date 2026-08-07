import request from '@/utils/request'

export function getDashboardStats() {
  return request.get('/dashboard/stats')
}

export function getLoginLogs() {
  return request.get('/dashboard/login-logs')
}

export function getVisitTrend(period: string, year?: number) {
  return request.get('/dashboard/visit-trend', { params: { period, year } })
}

export function getArticleTrend(period: string, year?: number) {
  return request.get('/dashboard/article-trend', { params: { period, year } })
}

export function getMyAuditArticles(params?: { page?: number; pageSize?: number }) {
  return request.get('/articles/my-audits', { params })
}
