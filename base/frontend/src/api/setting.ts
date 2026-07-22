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
