import request from '@/utils/request'

export const memberApi = {
  // Browse
  getOpenBatches: (params?: any) => request.get('/member/batches', { params }),
  getCategories: () => request.get('/member/categories'),
  getAnnouncements: (params?: any) => request.get('/member/announcements', { params }),

  // Applications (项目申报)
  getMyApplications: (params?: any) => request.get('/member/applications', { params }),
  getApplication: (id: number) => request.get(`/member/applications/${id}`),
  createApplication: (data: any) => request.post('/member/applications', data),
  updateApplication: (id: number, data: any) => request.put(`/member/applications/${id}`, data),
  deleteApplication: (id: number) => request.delete(`/member/applications/${id}`),
  submitApplication: (id: number) => request.post(`/member/applications/${id}/submit`),
  withdrawApplication: (id: number) => request.post(`/member/applications/${id}/withdraw`),
  saveMaterials: (id: number, data: any) => request.put(`/member/applications/${id}/materials`, data),

  // Certificates (证书下载)
  getMyCertificates: () => request.get('/member/certificates'),

  // Notifications (进度通知)
  getNotifications: (params?: any) => request.get('/member/notifications', { params }),
  markRead: (id: number) => request.put(`/member/notifications/${id}/read`),
  getUnreadCount: () => request.get('/member/notifications/unread-count'),

  // Dashboard
  getDashboard: () => request.get('/member/dashboard'),

  // Upload
  uploadFile: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return request.post('/member/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
}
