import request from '@/utils/request'

export interface SettingItem {
  category: string
  key: string
  value: string
  type?: string
  remark?: string
}

export interface SettingMap {
  [category: string]: {
    [key: string]: string
  }
}

export function getSettings(category?: string) {
  return request.get('/settings', { params: { category } })
}

export function saveSettings(settings: SettingItem[]) {
  return request.put('/settings', { settings })
}

export interface EmailTestReq {
  host: string
  port: number
  username: string
  password: string
  from: string
  ssl: boolean
  to: string
}

export function testEmail(data: EmailTestReq) {
  return request.post('/settings/email/test', data)
}
