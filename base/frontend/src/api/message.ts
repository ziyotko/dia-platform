import request from '@/utils/request'

export interface Message {
  id: number
  senderId: number
  senderName: string
  receiverId: number
  receiverType: string
  title: string
  content: string
  type: string
  priority: string
  /** 2草稿 3已发送 */
  status: number
  isRead: boolean
  readAt?: string
  sendAt?: string
}

export interface MessageQuery {
  page: number
  size: number
  /** inbox 收件箱 / sent 发件箱 */
  box?: 'inbox' | 'sent'
  /** -1 全部，0 未读，1 已读 */
  isRead?: number
  status?: number
}

export interface MessageTemplate {
  id: number
  code: string
  name: string
  channel: string
  subject?: string
  content: string
  variables?: string
  status: number
  description?: string
}

export function getMessageList(params: MessageQuery) {
  return request.get('/messages', { params })
}

/** 当前已接入（已配置）的站外发送渠道，如 ['email','wechat'] */
export function getMessageChannels() {
  return request.get<{ channels: string[] }>('/messages/channels')
}

/** 发送表单：可直接发送，也可按模板发送（未填标题/内容时后端用模板渲染） */
export interface SendMessageForm {
  receiverIds: number[]
  receiverType: string
  title: string
  content: string
  type?: string
  priority?: string
  channel?: string
  templateCode?: string
  vars?: Record<string, string>
}

export function sendMessage(data: SendMessageForm) {
  return request.post('/messages/send', data)
}

/** 草稿表单：receiverId = 0 表示广播草稿 */
export interface DraftForm {
  receiverId: number
  title: string
  content: string
  type?: string
  priority?: string
}

/** 新建草稿 */
export function createMessageDraft(data: DraftForm) {
  return request.post('/messages', data)
}

/** 编辑草稿（仅本人草稿） */
export function updateMessageDraft(id: number, data: DraftForm) {
  return request.put(`/messages/${id}`, data)
}

/** 发送已有草稿 */
export function sendMessageDraft(id: number) {
  return request.post(`/messages/${id}/send`)
}

/** 渲染模板中的 {{占位符}}（与后端 RenderMessageTemplate 同口径） */
export function renderMessageTemplate(text: string, vars: Record<string, string>) {
  return (text || '').replace(/\{\{\s*([\w.]+)\s*\}\}/g, (raw, key: string) =>
    vars[key] === undefined || vars[key] === '' ? raw : vars[key]
  )
}

export function getUnreadCount() {
  return request.get('/messages/unread-count')
}

export function markMessageRead(id: number) {
  return request.post(`/messages/${id}/read`)
}

/** 全部标为已读 */
export function markAllMessageRead() {
  return request.post('/messages/read-all', {})
}

export function deleteMessage(id: number) {
  return request.delete(`/messages/${id}`)
}

export function getMessageTemplateList(params: { page: number; size: number; keyword?: string }) {
  return request.get('/message-templates', { params })
}

export function createMessageTemplate(data: MessageTemplate) {
  return request.post('/message-templates', data)
}

export function updateMessageTemplate(id: number, data: MessageTemplate) {
  return request.put(`/message-templates/${id}`, data)
}

export function deleteMessageTemplate(id: number) {
  return request.delete(`/message-templates/${id}`)
}
