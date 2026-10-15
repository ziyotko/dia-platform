import request from '@/utils/request'

export interface LoginData {
  /** 用户名或邮箱：两者都传时后端优先按邮箱匹配，登录页按是否含 @ 只填其中一个 */
  account?: string
  email?: string
  password: string
  captcha_id: string
  captcha_code: string
}

export interface LoginResult {
  token: string
  user: {
    id: number
    username: string
    avatar: string
    roleIds: number[]
  }
}

export interface ProfileUser {
  id: number
  username: string
  email: string
  phone: string
  account: string
  roleName: string
  avatar: string
  createdAt: string
  onlineDays: number
  articleCount: number
  operationCount: number
  bio: string
}

export interface ProfileForm {
  email: string
  phone: string
  bio: string
  avatar: string
}

export interface ChangePasswordForm {
  oldPassword: string
  newPassword: string
}

export function getCaptcha() {
  return request.get<{ captcha_id: string; captcha_img: string }>('captcha')
}

export function login(data: LoginData) {
  return request.post<LoginResult>('login', data)
}

export function getUserInfo() {
  return request.get<{ user: ProfileUser }>('profile')
}

export function updateProfile(data: ProfileForm) {
  return request.put('profile', data)
}

export function changePassword(data: ChangePasswordForm) {
  return request.put('profile/password', data)
}

export function logout() {
  return request.post('logout')
}
