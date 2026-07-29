import request from '@/utils/request'

export const dashboardApi = {
  getMemberDashboard: () => request.get('/member/dashboard')
}

export const applicationApi = {
  getMyApplications: () => request.get('/applications'),
  getApplication: (id: number) => request.get(`/applications/${id}`),
  createApplication: (data: any) => request.post('/applications', data),
  saveDraft: (data: any) => request.post('/applications/draft', data),
  withdraw: (id: number) => request.post(`/applications/${id}/withdraw`)
}

export const feeApi = {
  getMyFees: (params?: any) => request.get('/fees', { params }),
  payFee: (id: number) => request.post(`/fees/${id}/pay`)
}

export const certificateApi = {
  getMyCertificates: () => request.get('/certificates'),
  getCertificate: (id: number) => request.get(`/certificates/${id}`),
  renewCertificate: () => request.post('/certificates/renew')
}

export const orgApi = {
  getTree: () => request.get('/organizations/tree'),
  getMyOrgs: () => request.get('/member/orgs'),
  joinOrg: (orgId: number) => request.post('/member/orgs', { org_id: orgId }),
  leaveOrg: (id: number) => request.delete(`/member/orgs/${id}`)
}

export const messageApi = {
  getMyMessages: (params?: any) => request.get('/messages', { params }),
  getMessage: (id: number) => request.get(`/messages/${id}`),
  createMessage: (data: any) => request.post('/messages', data)
}

export const articleApi = {
  getCategories: () => request.get('/article-categories'),
  getMyArticles: (params?: any) => request.get('/articles', { params }),
  getArticle: (id: number) => request.get(`/articles/${id}`),
  createArticle: (data: any) => request.post('/articles', data),
  updateArticle: (id: number, data: any) => request.put(`/articles/${id}`, data),
  deleteArticle: (id: number) => request.delete(`/articles/${id}`),
  listPublished: (params?: any) => request.get('/published-articles', { params })
}

export const announcementApi = {
  getPublished: (params?: any) => request.get('/announcements', { params }),
  getDetail: (id: number) => request.get(`/announcements/${id}`)
}
