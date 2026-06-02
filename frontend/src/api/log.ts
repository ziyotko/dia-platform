import request from '@/utils/request'

export interface LogQuery {
  page?: number
  pageSize?: number
  username?: string
  type?: string
  startDate?: string
  endDate?: string
}

export function getLogList(params: LogQuery) {
  return request.get('/logs', { params })
}

export function clearLogs() {
  return request.delete('/logs')
}
