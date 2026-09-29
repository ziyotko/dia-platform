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

/** 会员专属内容（新闻/数据/视频） */
export interface MemberContentForm {
  id?: number
  memberColumnId: number | undefined
  title: string
  type: number
  summary?: string
  content?: string
  cover?: string
  videoUrl?: string
  attachmentName?: string
  attachmentUrl?: string
  source?: string
  publishTime?: string
  status: number
  isTop: number
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
