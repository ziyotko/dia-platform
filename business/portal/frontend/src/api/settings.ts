import request from '@/utils/request'

export interface Settings {
  siteName: string
  siteUrl: string
  logo: string
  icp: string
  copyright: string
  orgName: string
  orgCode: string
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
  staticPath: string
  homeGray: boolean
  homeStaticTimeEnabled: boolean
  homeStaticTime: string
  columnStaticTimeEnabled: boolean
  columnStaticTime: string
  specialStaticTimeEnabled: boolean
  specialStaticTime: string
  detailStaticTimeEnabled: boolean
  detailStaticTime: string
  staticProgramAddr: string
  staticProgramTokenName: string
}

export interface SiteInfo {
  siteName: string
  siteUrl: string
  logo: string
  icp: string
  copyright: string
}

export function getMinPasswordLengthSettings() {
  return request.get<Settings>('/minPasswordLengthSettings')
}

export function getSettings() {
  return request.get<Settings>('/settings')
}

export function updateSettings(data: Partial<Settings>) {
  return request.put('/settings', data)
}

/** 邮件（SMTP）连接测试：后端按当前设置建立连接与认证，不发送邮件 */
export function testEmailConnection() {
  return request.post('/settings/test-email')
}

export function getPublicSiteInfo() {
  return request.get<SiteInfo>('/site-info')
}
