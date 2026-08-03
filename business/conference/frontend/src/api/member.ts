import request from '@/utils/request'

export const memberApi = {
  // Meetings
  getMeetings: (params?: any) => request.get('/member/meetings', { params }),
  getMeetingDetail: (id: number) => request.get(`/member/meetings/${id}`),

  // Registrations
  register: (meetingId: number) => request.post(`/member/meetings/${meetingId}/register`),
  getRegistrationStatus: (meetingId: number) => request.get(`/member/meetings/${meetingId}/registration`),
  getMyRegistrations: (params?: any) => request.get('/member/registrations', { params }),
  cancelRegistration: (id: number) => request.delete(`/member/registrations/${id}`),

  // Sign-in
  getSignInStatus: (meetingId: number) => request.get(`/member/meetings/${meetingId}/sign-in/status`),
  generateQRCode: (meetingId: number, regId: number) => request.post(`/member/meetings/${meetingId}/sign-in/qrcode?registrationId=${regId}`),
  signInOnline: (meetingId: number) => request.post(`/member/meetings/${meetingId}/sign-in/online`),
  signOut: (meetingId: number) => request.post(`/member/meetings/${meetingId}/sign-out`),

  // Votes
  getAvailableVotes: () => request.get('/member/votes'),
  hasVoted: (id: number) => request.get(`/member/votes/${id}/has-voted`),
  castVote: (id: number, data: any) => request.post(`/member/votes/${id}/cast`, data),

  // Finance
  createOrder: (meetingId: number) => request.post(`/member/meetings/${meetingId}/orders`),
  payOrder: (id: number, data: any) => request.post(`/member/orders/${id}/pay`, data),
  applyRefund: (id: number, data: any) => request.post(`/member/orders/${id}/refund`, data),
  getMyOrders: (params?: any) => request.get('/member/orders', { params }),
  saveInvoice: (id: number, data: any) => request.put(`/member/orders/${id}/invoice`, data),
  getInvoice: (id: number) => request.get(`/member/orders/${id}/invoice`),

  // Live
  getLiveUrl: (meetingId: number) => request.get(`/member/meetings/${meetingId}/live/url`),
  getLiveMessages: (meetingId: number, params?: any) => request.get(`/member/meetings/${meetingId}/live/messages`, { params }),
  sendLiveMessage: (meetingId: number, data: any) => request.post(`/member/meetings/${meetingId}/live/messages`, data),
  startViewing: (meetingId: number, type: string) => request.post(`/member/meetings/${meetingId}/live/viewing?type=${type}`),

  // Surveys
  getAvailableSurveys: () => request.get('/member/surveys'),
  hasSubmitted: (id: number) => request.get(`/member/surveys/${id}/has-submitted`),
  submitSurvey: (id: number, data: any) => request.post(`/member/surveys/${id}/submit`, data),

  // Credits
  getMyCredits: (params?: any) => request.get('/member/credits', { params }),

  // Notifications
  getNotifications: (params?: any) => request.get('/member/notifications', { params }),
  markRead: (id: number) => request.put(`/member/notifications/${id}/read`),
  getUnreadCount: () => request.get('/member/notifications/unread-count'),

  // Dashboard
  getDashboard: () => request.get('/member/dashboard'),
}
