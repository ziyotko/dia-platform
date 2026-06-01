import request from '@/utils/request'

export interface LoginData {
  username: string
  password: string
  captchaId: string
  captchaCode: string
}

export interface LoginResult {
  token: string
  user: {
    id: number
    username: string
    nickname: string
    avatar: string
  }
}

export function getCaptcha() {
  return request.get<{ captchaId: string; imageBase64: string }>('/auth/captcha')
}

export function login(data: LoginData) {
  return request.post<LoginResult>('/auth/login', data)
}

export function getUserInfo() {
  return request.get('/auth/info')
}

export function logout() {
  return request.post('/auth/logout')
}