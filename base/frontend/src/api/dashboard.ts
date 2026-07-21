import request from '@/utils/request'

export interface DashboardStats {
  tenantCount: number
  appCount: number
  userCount: number
  roleCount: number
  organizationCount: number
  messageCount: number
}

export function getDashboardStats() {
  return request.get<DashboardStats>('/dashboard/stats')
}
