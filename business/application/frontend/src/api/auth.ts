import request from '@/utils/request'

export const authApi = {
  getCaptcha: () => request.get('/captcha'),

  // Applicant (申报人)
  userLogin: (data: any) => request.post('/member/login', data),
  userRegister: (data: any) => request.post('/member/register', data),
  getUserProfile: () => request.get('/member/profile'),
  updateUserProfile: (data: any) => request.put('/member/profile', data),
  changeUserPassword: (data: any) => request.put('/member/change-password', data),
  // 退出登录：服务端把当前 Token 的 jti 写入黑名单（TTL = 剩余有效期），使其立即失效
  userLogout: () => request.post('/member/logout'),

  // Admin (管理人 / 评审人)
  adminLogin: (data: any) => request.post('/admin/login', data),
  getAdminProfile: () => request.get('/admin/profile'),
  updateAdminProfile: (data: any) => request.put('/admin/profile', data),
  changeAdminPassword: (data: any) => request.put('/admin/change-password', data),
  // 退出登录（管理端）：写黑名单，与管理端登录页对应
  adminLogout: () => request.post('/admin/logout'),
}
