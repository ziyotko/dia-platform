import request from '@/utils/request'

/** 会员栏目（会员栏目分类） */
export interface MemberColumnForm {
  id?: number
  name: string
  code: string
  description?: string
  sort: number
  status: number
}

/** 会员专属内容（按业务类型使用不同字段组合，未使用的类型字段留空） */
export interface MemberContentForm {
  id?: number
  memberColumnId: number | undefined
  title: string
  /** 1新闻 2数据 3视频 4报刊 */
  type: number
  source?: string
  publishTime?: string
  status: number
  isTop: number
  /** 新闻/视频/报刊：封面图 */
  cover?: string
  /** 新闻：文章内容、文章附件 */
  content?: string
  attachmentName?: string
  attachmentUrl?: string
  /** 数据 */
  dataYear?: string
  unitName?: string
  province?: string
  region?: string
  isBelt?: number
  isAxis?: number
  subField?: string
  mainBusinessIncome?: number
  /** 视频：完整视频、预览视频 */
  fullVideoUrl?: string
  previewVideoUrl?: string
  /** 报刊：期号、出版年月、摘要、报刊文件 */
  issueNo?: string
  publishYearMonth?: string
  summary?: string
  paperFileName?: string
  paperFileUrl?: string
}

export interface MemberColumnQuery {
  name?: string
  status?: number
  page?: number
  pageSize?: number
}

export interface MemberContentQuery {
  title?: string
  columnId?: number
  type?: number
  status?: number
  page?: number
  pageSize?: number
}

export function getMemberColumns(params: MemberColumnQuery) {
  return request.get('/member-columns', { params })
}

export function createMemberColumn(data: MemberColumnForm) {
  return request.post('/member-columns', data)
}

export function updateMemberColumn(id: number, data: MemberColumnForm) {
  return request.put(`/member-columns/${id}`, data)
}

export function updateMemberColumnStatus(id: number, status: number) {
  return request.patch(`/member-columns/${id}/status`, { status })
}

export function deleteMemberColumn(id: number) {
  return request.delete(`/member-columns/${id}`)
}

export function getMemberContents(params: MemberContentQuery) {
  return request.get('/member-contents', { params })
}

export function getMemberContent(id: number) {
  return request.get(`/member-contents/${id}`)
}

export function createMemberContent(data: MemberContentForm) {
  return request.post('/member-contents', data)
}

export function updateMemberContent(id: number, data: MemberContentForm) {
  return request.put(`/member-contents/${id}`, data)
}

export function updateMemberContentStatus(id: number, status: number) {
  return request.patch(`/member-contents/${id}/status`, { status })
}

export function deleteMemberContent(id: number) {
  return request.delete(`/member-contents/${id}`)
}
