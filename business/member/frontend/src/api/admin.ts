import request from '@/utils/request'

export const adminApi = {
  // Member management
  getMembers: (params?: any) => request.get('/admin/members', { params }),
  getMember: (id: number) => request.get(`/admin/members/${id}`),
  updateMemberStatus: (id: number, status: string) => request.put(`/admin/members/${id}/status`, { status }),
  updateMemberLevel: (id: number, level: string) => request.put(`/admin/members/${id}/level`, { level }),
  deleteMember: (id: number) => request.delete(`/admin/members/${id}`),
  getMemberStats: () => request.get('/admin/member-stats'),

  // Applications
  getApplications: (params?: any) => request.get('/admin/applications', { params }),
  reviewApplication: (id: number, data: any) => request.put(`/admin/applications/${id}/review`, data),

  // Fees
  getFees: (params?: any) => request.get('/admin/fees', { params }),
  createFee: (data: any) => request.post('/admin/fees', data),
  updateFee: (id: number, data: any) => request.put(`/admin/fees/${id}`, data),

  // Certificates
  createCertificate: (data: any) => request.post('/admin/certificates', data),
  updateCertificate: (id: number, data: any) => request.put(`/admin/certificates/${id}`, data),
  generateCertificate: (memberId: number) => request.post(`/admin/certificates/generate/${memberId}`),

  // Organizations
  createOrg: (data: any) => request.post('/admin/organizations', data),
  updateOrg: (id: number, data: any) => request.put(`/admin/organizations/${id}`, data),
  deleteOrg: (id: number) => request.delete(`/admin/organizations/${id}`),

  // Messages
  getMessages: (params?: any) => request.get('/admin/messages', { params }),
  replyMessage: (id: number, reply: string) => request.put(`/admin/messages/${id}/reply`, { reply }),

  // Articles
  createCategory: (data: any) => request.post('/admin/article-categories', data),
  updateCategory: (id: number, data: any) => request.put(`/admin/article-categories/${id}`, data),
  deleteCategory: (id: number) => request.delete(`/admin/article-categories/${id}`),
  getArticles: (params?: any) => request.get('/admin/articles', { params }),
  reviewArticle: (id: number, data: any) => request.put(`/admin/articles/${id}/review`, data),

  // Announcements
  createAnnouncement: (data: any) => request.post('/admin/announcements', data),
  updateAnnouncement: (id: number, data: any) => request.put(`/admin/announcements/${id}`, data),
  deleteAnnouncement: (id: number) => request.delete(`/admin/announcements/${id}`),
  getAnnouncements: (params?: any) => request.get('/admin/announcements', { params })
}
