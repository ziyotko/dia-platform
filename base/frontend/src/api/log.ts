import request from '@/utils/request'

export interface OperationLog {
  id: number
  tenantId: number
  userId: number
  username: string
  module: string
  action: string
  method: string
  path: string
  ip: string
  params: string
  result: string
  status: number
  duration: number
  operationAt: string
}

export interface LogQuery {
  page: number
  size: number
  username?: string
  module?: string
  action?: string
  method?: string
  path?: string
  status?: number
  startAt?: string
  endAt?: string
}

export function getLogList(params: LogQuery) {
  return request.get('/operation-logs', { params })
}

export function deleteLogs(ids: number[]) {
  return request.post('/operation-logs/delete', { ids })
}

export function clearLogs(days: number) {
  return request.post('/operation-logs/clear', null, { params: { days } })
}

export function exportLogs(params: Omit<LogQuery, 'page' | 'size'>) {
  return request.get('/operation-logs/export', {
    params,
    responseType: 'blob'
  })
}
