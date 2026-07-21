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
  status: number
  readAt?: string
  sendAt?: string
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

export function getMessageList(params: { page: number; size: number; status?: number }) {
  return request.get('/messages', { params })
}

export function sendMessage(data: {
  receiverIds: number[]
  receiverType: string
  title: string
  content: string
  type?: string
  priority?: string
}) {
  return request.post('/messages/send', data)
}

export function getUnreadCount() {
  return request.get('/messages/unread-count')
}

export function markMessageRead(id: number) {
  return request.post(`/messages/${id}/read`)
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
