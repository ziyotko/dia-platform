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

// 各类型页面最后一次静态化成功时间（全站/首页/栏目页/专题页/详情页）
export function getStaticLatestTimes() {
  return request.get('/static-logs/latest-times')
}
