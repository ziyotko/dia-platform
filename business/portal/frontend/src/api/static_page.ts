import request from '@/utils/request'

// 静态化页面列表：页面层已合并进模板，这里直接返回启用中的模板
// pageType 为空表示全部类型（home/column/detail/special）
export function getStaticPages(params?: { pageType?: string }) {
  return request.get('/static-pages', { params })
}
