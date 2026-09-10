import request from '@/utils/request'

export const adminApi = {
  // Dashboard
  getDashboard: () => request.get('/admin/dashboard'),

  // Upload
  uploadFile: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return request.post('/admin/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  // Categories (项目类别)
  getCategories: () => request.get('/admin/categories'),
  createCategory: (data: any) => request.post('/admin/categories', data),
  updateCategory: (id: number, data: any) => request.put(`/admin/categories/${id}`, data),
  deleteCategory: (id: number) => request.delete(`/admin/categories/${id}`),

  // Batches (申报批次)
  getBatches: (params?: any) => request.get('/admin/batches', { params }),
  getBatch: (id: number) => request.get(`/admin/batches/${id}`),
  createBatch: (data: any) => request.post('/admin/batches', data),
  updateBatch: (id: number, data: any) => request.put(`/admin/batches/${id}`, data),
  deleteBatch: (id: number) => request.delete(`/admin/batches/${id}`),
  publishBatch: (id: number) => request.post(`/admin/batches/${id}/publish`),
  closeBatch: (id: number) => request.post(`/admin/batches/${id}/close`),
  startReview: (id: number) => request.post(`/admin/batches/${id}/start-review`),

  // Applications (项目申报管理 / 初审)
  getApplications: (params?: any) => request.get('/admin/applications', { params }),
  getApplication: (id: number) => request.get(`/admin/applications/${id}`),
  preliminaryReview: (id: number, data: any) => request.post(`/admin/applications/${id}/preliminary`, data),
  assignReviewers: (id: number, data: any) => request.post(`/admin/applications/${id}/assign`, data),
  finalize: (id: number, data: any) => request.post(`/admin/applications/${id}/finalize`, data),
  publishResult: (id: number) => request.post(`/admin/applications/${id}/publish`),

  // Reviews (专家评审)
  getReviewers: () => request.get('/admin/reviewers'),
  getMyReviews: (params?: any) => request.get('/admin/reviews', { params }),
  getReview: (id: number) => request.get(`/admin/reviews/${id}`),
  submitReview: (id: number, data: any) => request.post(`/admin/reviews/${id}/submit`, data),
  getReviewAssignments: (applicationId: number) => request.get('/admin/review-assignments', { params: { applicationId } }),

  // Announcements (结果公示)
  getAnnouncements: (params?: any) => request.get('/admin/announcements', { params }),
  createAnnouncement: (data: any) => request.post('/admin/announcements', data),
  updateAnnouncement: (id: number, data: any) => request.put(`/admin/announcements/${id}`, data),
  deleteAnnouncement: (id: number) => request.delete(`/admin/announcements/${id}`),
  publishAnnouncement: (id: number) => request.post(`/admin/announcements/${id}/publish`),

  // Certificates (证书管理)
  getCertificates: (params?: any) => request.get('/admin/certificates', { params }),
  issueCertificate: (data: any) => request.post('/admin/certificates', data),
  updateCertificate: (id: number, data: any) => request.put(`/admin/certificates/${id}`, data),

  // Notifications (通知管理)
  sendNotification: (data: any) => request.post('/admin/notifications', data),

  // Users (申报人管理)
  getUsers: (params?: any) => request.get('/admin/users', { params }),
  createUser: (data: any) => request.post('/admin/users', data),
  updateUser: (id: number, data: any) => request.put(`/admin/users/${id}`, data),
  deleteUser: (id: number) => request.delete(`/admin/users/${id}`),
  setUserStatus: (id: number, data: any) => request.put(`/admin/users/${id}/status`, data),

  // Admins & Roles (账号/角色管理)
  getAdmins: (params?: any) => request.get('/admin/admins', { params }),
  createAdmin: (data: any) => request.post('/admin/admins', data),
  updateAdmin: (id: number, data: any) => request.put(`/admin/admins/${id}`, data),
  deleteAdmin: (id: number) => request.delete(`/admin/admins/${id}`),
  getRoles: () => request.get('/admin/roles'),

  // Audit logs
  getAuditLogs: (params?: any) => request.get('/admin/audit-logs', { params }),
}
