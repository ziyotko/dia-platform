import request from '@/utils/request'

export interface Settings {
  siteName: string
  logo: string
  icp: string
  copyright: string
  captchaEnabled: boolean
  lockEnabled: boolean
  maxFailCount: number
  lockDuration: number
  minPasswordLength: number
  tokenExpire: number
  smtpHost: string
  smtpPort: string
  fromEmail: string
  fromName: string
  emailPassword: string
  ssl: boolean
  themeColor: string
  sidebarStyle: string
  tagsView: boolean
  breadcrumb: boolean
}

export function getSettings() {
  return request.get<Settings>('/settings')
}

export function updateSettings(data: Partial<Settings>) {
  return request.put('/settings', data)
}
