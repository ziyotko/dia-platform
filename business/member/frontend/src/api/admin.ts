import request from '@/utils/request'

export const adminApi = {
  // Member management
  getMembers: (params?: any) => request.get('/admin/members', { params }),
  createMember: (data: any) => request.post('/admin/members', data),
  // 查重：field 支持 username/mobile/email/contact_mobile/company_name/credit_code
  checkMemberExists: (field: string, value: string) => request.get('/admin/member-exists', { params: { field, value } }),
  getMember: (id: number) => request.get(`/admin/members/${id}`),
  updateMemberStatus: (id: number, status: string) => request.put(`/admin/members/${id}/status`, { status }),
  updateMemberLevel: (id: number, levelId: number, reason: string) => request.put(`/admin/members/${id}/level`, { level_id: levelId, reason }),
  resetMemberPassword: (id: number) => request.put(`/admin/members/${id}/reset-password`),
  getMemberLevelOptions: (id: number) => request.get(`/admin/members/${id}/level-options`),
  getMemberLevelChangesByMember: (id: number, params?: any) => request.get(`/admin/members/${id}/level-changes`, { params }),
  getMemberOrgs: (id: number) => request.get(`/admin/members/${id}/orgs`),
  deleteMember: (id: number) => request.delete(`/admin/members/${id}`),
  getMemberStats: () => request.get('/admin/member-stats'),
  getMemberLevelChanges: (params?: any) => request.get('/admin/member-level-changes', { params }),
  getLevelChangeYears: () => request.get('/admin/member-level-changes/years'),
  exportMemberLevelChanges: (params?: any) => request.get('/admin/member-level-changes/export', { params, responseType: 'blob' }),
  getProfileChanges: (params?: any) => request.get('/admin/member-profile-changes', { params }),

  // Applications
  getApplications: (params?: any) => request.get('/admin/applications', { params }),
  reviewApplication: (id: number, data: any) => request.put(`/admin/applications/${id}/review`, data),

  // Fees
  getFees: (params?: any) => request.get('/admin/fees', { params }),
  createFee: (data: any) => request.post('/admin/fees', data),
  updateFee: (id: number, data: any) => request.put(`/admin/fees/${id}`, data),
  confirmFee: (id: number, data?: any) => request.post(`/admin/fees/${id}/confirm`, data),
  deleteFee: (id: number) => request.delete(`/admin/fees/${id}`),
  getMemberFeeInfo: (id: number) => request.get(`/admin/members/${id}/fee-info`),
  issueInvoice: (id: number, data: FormData) => request.post(`/admin/fees/${id}/issue-invoice`, data, { headers: { 'Content-Type': 'multipart/form-data' } }),

  // Certificates
  createCertificate: (data: any) => request.post('/admin/certificates', data),
  getCertificates: (params?: any) => request.get('/admin/certificates', { params }),
  regenerateCertificate: (id: number) => request.post(`/admin/certificates/${id}/generate`),
  // 批量补生成 file_path 为空的存量证书 PDF
  regenerateMissingCertificates: (limit = 200) => request.post(`/admin/certificates-regenerate-missing`, null, { params: { limit } }),
  updateCertificate: (id: number, data: any) => request.put(`/admin/certificates/${id}`, data),
  // Certificate Templates
  getCertTemplates: () => request.get('/admin/certificate-templates'),
  getCertTemplate: (id: number) => request.get(`/admin/certificate-templates/${id}`),
  createCertTemplate: (data: any) => request.post('/admin/certificate-templates', data),
  updateCertTemplate: (id: number, data: any) => request.put(`/admin/certificate-templates/${id}`, data),
  deleteCertTemplate: (id: number) => request.delete(`/admin/certificate-templates/${id}`),

  // Organizations
  createOrg: (data: any) => request.post('/admin/organizations', data),
  updateOrg: (id: number, data: any) => request.put(`/admin/organizations/${id}`, data),
  deleteOrg: (id: number) => request.delete(`/admin/organizations/${id}`),
  getOrgLevels: (id: number) => request.get(`/admin/organizations/${id}/levels`),
  setOrgLevels: (id: number, levelIds: number[]) => request.put(`/admin/organizations/${id}/levels`, { level_ids: levelIds }),

  // Member level definitions
  getMemberLevels: (params?: any) => request.get('/admin/member-levels', { params }),
  createMemberLevel: (data: any) => request.post('/admin/member-levels', data),
  updateLevelDefinition: (id: number, data: any) => request.put(`/admin/member-levels/${id}`, data),
  deleteMemberLevel: (id: number) => request.delete(`/admin/member-levels/${id}`),
  moveLevelUp: (id: number) => request.put(`/admin/member-levels/${id}/move-up`),
  moveLevelDown: (id: number) => request.put(`/admin/member-levels/${id}/move-down`),

  // Messages
  getMessages: (params?: any) => request.get('/admin/messages', { params }),
  replyMessage: (id: number, reply: string) => request.put(`/admin/messages/${id}/reply`, { reply }),
	deleteMessage: (id: number) => request.delete(`/admin/messages/${id}`),
  updateCategory: (id: number, data: any) => request.put(`/admin/article-categories/${id}`, data),
  createCategory: (data: { name: string; sort: number }) => request.post('/admin/article-categories', data),
  deleteCategory: (id: number) => request.delete(`/admin/article-categories/${id}`),
  getArticles: (params?: any) => request.get('/admin/articles', { params }),
  reviewArticle: (id: number, data: any) => request.put(`/admin/articles/${id}/review`, data),
	deleteArticle: (id: number) => request.delete(`/admin/articles/${id}`),
  createAnnouncement: (data: any) => request.post('/admin/announcements', data),
  updateAnnouncement: (id: number, data: any) => request.put(`/admin/announcements/${id}`, data),
  deleteAnnouncement: (id: number) => request.delete(`/admin/announcements/${id}`),
  getAnnouncements: (params?: any) => request.get('/admin/announcements', { params }),

  // Fee Standards
  getFeeStandards: () => request.get('/admin/fee-standards'),
  getFeeStandardsByLevel: (levelId: number) => request.get(`/admin/fee-standards/levels/${levelId}`),
  upsertFeeStandard: (data: { level_id: number; year: number; amount: number }) => request.post('/admin/fee-standards', data),
  batchUpsertFeeStandard: (data: { level_id: number; items: { year: number; amount: number }[] }) => request.post('/admin/fee-standards/batch', data),
  deleteFeeStandard: (id: number) => request.delete(`/admin/fee-standards/${id}`),

  // System Configs
  getSystemConfigs: () => request.get('/admin/system-configs'),
  createSystemConfig: (data: { key: string; value: string; description?: string }) => request.post('/admin/system-configs', data),
  updateSystemConfig: (id: number, data: { key?: string; value?: string; description?: string }) => request.put(`/admin/system-configs/${id}`, data),
  deleteSystemConfig: (id: number) => request.delete(`/admin/system-configs/${id}`),

  // 协会章程（富文本正文，用富文本编辑器维护）
  getCharter: () => request.get('/admin/charter-content'),
  saveCharter: (content: string) => request.put('/admin/charter-content', { content }),
  // 协会章程 PDF 附件（供会员端「入会章程」下载）
  uploadCharterFile: (file: File) => {
    const fd = new FormData()
    fd.append('file', file)
    return request.post('/admin/charter-file', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
  },
  deleteCharterFile: () => request.delete('/admin/charter-file'),

  // Operation logs
  getOperationLogs: (params?: any) => request.get('/admin/operation-logs', { params })
}
