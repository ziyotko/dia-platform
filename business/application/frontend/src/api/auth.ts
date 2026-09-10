import request from '@/utils/request'

export const authApi = {
  getCaptcha: () => request.get('/captcha'),

  // Applicant (申报人)
  userLogin: (data: any) => request.post('/member/login', data),
  userRegister: (data: any) => request.post('/member/register', data),
  getUserProfile: () => request.get('/member/profile'),
  updateUserProfile: (data: any) => request.put('/member/profile', data),
  changeUserPassword: (data: any) => request.put('/member/change-password', data),

  // Admin (管理人 / 评审人)
  adminLogin: (data: any) => request.post('/admin/login', data),
  getAdminProfile: () => request.get('/admin/profile'),
  updateAdminProfile: (data: any) => request.put('/admin/profile', data),
  changeAdminPassword: (data: any) => request.put('/admin/change-password', data),
}
