import request from '@/utils/request'

export const authApi = {
  getCaptcha: () => request.get('/captcha'),
  register: (data: any) => request.post('/auth/register', data),
  checkExists: (data: any) => request.post('/auth/check-exists', data),
  login: (data: any) => request.post('/auth/login', data),
  // 退出登录：服务端把当前 Token 的 jti 写入黑名单（TTL = 剩余有效期），使其立即失效
  logout: () => request.post('/logout'),
  getProfile: () => request.get('/member/profile'),
  updateProfile: (data: any) => request.put('/member/profile', data),
  changePassword: (data: any) => request.put('/member/change-password', data),
  getSiteInfo: () => request.get('/site-info'),
  upload: (formData: FormData) => request.post('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  // 注册流程专用（匿名）：服务端固定 dir=certs、仅允许图片与 PDF
  uploadPublic: (formData: FormData) => request.post('/upload-public', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
