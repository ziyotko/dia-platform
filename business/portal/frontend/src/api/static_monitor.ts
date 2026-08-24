import request from '@/utils/request'

export interface StaticMonitor {
  online: boolean
  address: string
  httpStatus: number
  lastCheckTime: string
  message: string
}

export function getStaticMonitor() {
  return request.get<StaticMonitor>('/static-monitor')
}
