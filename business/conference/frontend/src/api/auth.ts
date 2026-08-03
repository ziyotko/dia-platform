import request from '@/utils/request'

export const authApi = {
  getCaptcha: () => request.get('/captcha'),

  memberLogin: (data: any) => request.post('/member/login', data),
  memberRegister: (data: any) => request.post('/member/register', data),
  getMemberProfile: () => request.get('/member/profile'),
  updateMemberProfile: (data: any) => request.put('/member/profile', data),
  changeMemberPassword: (data: any) => request.put('/member/change-password', data),

  adminLogin: (data: any) => request.post('/admin/login', data),
  getAdminProfile: () => request.get('/admin/profile'),
  updateAdminProfile: (data: any) => request.put('/admin/profile', data),
  changeAdminPassword: (data: any) => request.put('/admin/change-password', data),
}
