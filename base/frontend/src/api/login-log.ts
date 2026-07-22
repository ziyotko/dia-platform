import request from '@/utils/request'

export interface LoginLog {
  id: number
  tenantId: number
  userId: number
  username: string
  ip: string
  agent: string
  status: number
  message: string
  createdAt: string
}

export interface LoginLogQuery {
  page: number
  size: number
  username?: string
  status?: number
  startAt?: string
  endAt?: string
}

export function getLoginLogList(params: LoginLogQuery) {
  return request.get('/login-logs', { params })
}

export function deleteLoginLogs(ids: number[]) {
  return request.post('/login-logs/delete', { ids })
}

export function clearLoginLogs(days: number) {
  return request.post('/login-logs/clear', null, { params: { days } })
}

export function exportLoginLogs(params: Omit<LoginLogQuery, 'page' | 'size'>) {
  return request.get('/login-logs/export', {
    params,
    responseType: 'blob'
  })
}
