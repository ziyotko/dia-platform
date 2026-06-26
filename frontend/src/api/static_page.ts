import request from '@/utils/request'

export function getStaticPages(params?: { pageType?: string; templateId?: number }) {
  return request.get('/static-pages', { params })
}
