import request from '@/utils/request'

export const adminApi = {
  // Dashboard
  getDashboard: () => request.get('/admin/dashboard'),

  // Meetings
  getMeetings: (params?: any) => request.get('/admin/meetings', { params }),
  getMeeting: (id: number) => request.get(`/admin/meetings/${id}`),
  createMeeting: (data: any) => request.post('/admin/meetings', data),
  updateMeeting: (id: number, data: any) => request.put(`/admin/meetings/${id}`, data),
  deleteMeeting: (id: number) => request.delete(`/admin/meetings/${id}`),
  closeMeeting: (id: number) => request.post(`/admin/meetings/${id}/close`),
  saveAgendas: (id: number, data: any) => request.put(`/admin/meetings/${id}/agendas`, data),
  saveGuests: (id: number, data: any) => request.put(`/admin/meetings/${id}/guests`, data),

  // Registrations
  getRegistrations: (params?: any) => request.get('/admin/registrations', { params }),
  approveRegistration: (id: number, data?: any) => request.post(`/admin/registrations/${id}/approve`, data || {}),
  rejectRegistration: (id: number, data: any) => request.post(`/admin/registrations/${id}/reject`, data),
  promoteWaitlist: (id: number) => request.post(`/admin/registrations/${id}/promote`),
  getWaitlist: (params?: any) => request.get('/admin/registrations/waitlist', { params }),

  // Sign-in
  getSignIns: (params?: any) => request.get('/admin/sign-ins', { params }),
  signInByQR: (data: any) => request.post('/admin/sign-ins/qrcode', data),
  getSignInStats: (meetingId: number) => request.get(`/admin/meetings/${meetingId}/sign-in/stats`),

  // Votes
  getVotes: (params?: any) => request.get('/admin/votes', { params }),
  getVote: (id: number) => request.get(`/admin/votes/${id}`),
  createVote: (data: any) => request.post('/admin/votes', data),
  updateVote: (id: number, data: any) => request.put(`/admin/votes/${id}`, data),
  deleteVote: (id: number) => request.delete(`/admin/votes/${id}`),
  getVoteResults: (id: number) => request.get(`/admin/votes/${id}/results`),

  // Finance
  getOrders: (params?: any) => request.get('/admin/orders', { params }),
  getRefunds: (params?: any) => request.get('/admin/refunds', { params }),
  processRefund: (id: number, data: any) => request.post(`/admin/refunds/${id}/process`, data),
  getLedger: () => request.get('/admin/ledger'),

  // Live
  startLive: (meetingId: number) => request.post(`/admin/meetings/${meetingId}/live/start`),
  stopLive: (meetingId: number) => request.post(`/admin/meetings/${meetingId}/live/stop`),
  getLiveConfig: (meetingId: number) => request.get(`/admin/meetings/${meetingId}/live/config`),
  uploadVod: (meetingId: number, data: any) => request.post(`/admin/meetings/${meetingId}/vod/upload`, data),
  setReplayStatus: (meetingId: number, data: any) => request.put(`/admin/meetings/${meetingId}/vod/replay`, data),
  getViewingLogs: (params?: any) => request.get('/admin/viewing-logs', { params }),

  // Surveys
  getSurveys: (params?: any) => request.get('/admin/surveys', { params }),
  getSurvey: (id: number) => request.get(`/admin/surveys/${id}`),
  createSurvey: (data: any) => request.post('/admin/surveys', data),
  updateSurvey: (id: number, data: any) => request.put(`/admin/surveys/${id}`, data),
  deleteSurvey: (id: number) => request.delete(`/admin/surveys/${id}`),
  getSurveyResults: (id: number) => request.get(`/admin/surveys/${id}/results`),
  getSurveyAnswers: (id: number, params?: any) => request.get(`/admin/surveys/${id}/answers`, { params }),

  // Credits
  setCredits: (meetingId: number, data: any) => request.post(`/admin/meetings/${meetingId}/credits`, data),
  manualAdjust: (data: any) => request.post('/admin/credits/manual', data),
  getCreditRecords: (params?: any) => request.get('/admin/credits', { params }),
  getCreditStats: () => request.get('/admin/credits/stats'),

  // Archives
  archiveMeeting: (meetingId: number) => request.post(`/admin/meetings/${meetingId}/archive`),
  getArchives: (params?: any) => request.get('/admin/archives', { params }),
  getArchive: (id: number) => request.get(`/admin/archives/${id}`),
  addMaterial: (id: number, data: any) => request.post(`/admin/archives/${id}/materials`, data),

  // Notifications
  sendNotification: (data: any) => request.post('/admin/notifications', data),
  getSentNotifications: (params?: any) => request.get('/admin/notifications', { params }),
  getNotificationReceipt: (id: number) => request.get(`/admin/notifications/${id}/receipt`),

  // Audit
  getAuditLogs: (params?: any) => request.get('/admin/audit-logs', { params }),

  // Users
  getUsers: (params?: any) => request.get('/admin/users', { params }),
  getUser: (id: number) => request.get(`/admin/users/${id}`),
  createUser: (data: any) => request.post('/admin/users', data),
  updateUser: (id: number, data: any) => request.put(`/admin/users/${id}`, data),
  deleteUser: (id: number) => request.delete(`/admin/users/${id}`),
  setValidity: (id: number, data: any) => request.put(`/admin/users/${id}/validity`, data),

  // Admins
  getAdmins: (params?: any) => request.get('/admin/admins', { params }),
  createAdmin: (data: any) => request.post('/admin/admins', data),
  updateAdmin: (id: number, data: any) => request.put(`/admin/admins/${id}`, data),

  // Roles
  getRoles: () => request.get('/admin/roles'),
}
