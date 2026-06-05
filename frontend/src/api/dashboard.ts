import request from '@/utils/request'

export function getDashboardStats() {
  return request.get('/dashboard/stats')
}

export function getLoginLogs() {
  return request.get('/dashboard/login-logs')
}
