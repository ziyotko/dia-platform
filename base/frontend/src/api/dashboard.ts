import request from '@/utils/request'

export interface DailyPoint {
  /** YYYY-MM-DD */
  date: string
  count: number
}

export interface DashboardStats {
  // 资源概览（普通用户不返回，保持 0）
  tenantCount: number
  tenantEnabledCount: number
  appCount: number
  appIframeCount: number
  appProxyCount: number
  appInstanceCount: number
  userCount: number
  userEnabledCount: number
  userDisabledCount: number
  userAdminCount: number
  roleCount: number
  organizationCount: number
  messageCount: number

  // 我的待办（所有用户都有）
  myUnreadCount: number
  myTodoCount: number
  myRunningInstanceCount: number
  runningInstanceCount: number

  // 今日
  todayNewUserCount: number
  todayLoginCount: number
  todayLoginFailCount: number

  // 近 7 天趋势
  userTrend: DailyPoint[]
  loginTrend: DailyPoint[]
  messageTrend: DailyPoint[]
}

export function getDashboardStats() {
  return request.get<DashboardStats>('/dashboard/stats')
}
