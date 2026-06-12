import request from '@/utils/request'

export interface StaticLogQuery {
  page?: number
  pageSize?: number
}

export function getStaticLogList(params: StaticLogQuery) {
  return request.get('/static-logs', { params })
}

export function clearStaticLogs() {
  return request.delete('/static-logs')
}
