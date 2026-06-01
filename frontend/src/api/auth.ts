import request from '@/utils/request'

export interface LoginData {
  account: string
  password: string
  captcha_id: string
  captcha_code: string
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
  return request.get<{ captcha_id: string; captcha_img: string }>('/captcha')
}

export function login(data: LoginData) {
  return request.post<LoginResult>('/login', data)
}

export function getUserInfo() {
  return request.get('/profile')
}

export function logout() {
  return request.post('/logout')
}