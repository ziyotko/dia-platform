import request from '@/utils/request'

export function getDashboardStats() {
  return request.get('/dashboard/stats')
}

export function getLoginLogs() {
  return request.get('/dashboard/login-logs')
}

export function getVisitTrend(period: string) {
  return request.get('/dashboard/visit-trend', { params: { period } })
}
