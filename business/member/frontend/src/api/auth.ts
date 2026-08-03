import request from '@/utils/request'

export const authApi = {
  getCaptcha: () => request.get('/captcha'),
  register: (data: any) => request.post('/auth/register', data),
  checkExists: (data: any) => request.post('/auth/check-exists', data),
  login: (data: any) => request.post('/auth/login', data),
  getProfile: () => request.get('/member/profile'),
  updateProfile: (data: any) => request.put('/member/profile', data),
  changePassword: (data: any) => request.put('/member/change-password', data),
  sendResetEmail: (email: string) => request.post('/auth/send-reset-email', { email }),
  resetPassword: (data: any) => request.post('/auth/reset-password', data),
  getSiteInfo: () => request.get('/site-info'),
  upload: (formData: FormData) => request.post('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
