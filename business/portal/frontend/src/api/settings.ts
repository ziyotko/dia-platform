import request from '@/utils/request'

export interface Settings {
  siteName: string
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
  logo: string
  icp: string
  copyright: string
}

export function getSettings() {
  return request.get<Settings>('/settings')
}

export function updateSettings(data: Partial<Settings>) {
  return request.put('/settings', data)
}

export function getPublicSiteInfo() {
  return request.get<SiteInfo>('/site-info')
}
