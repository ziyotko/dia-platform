import request from '@/utils/request'

export interface ArticleAnalyticsTrend {
  labels: string[]
  like: number[]
  share: number[]
  visit: number[]
  totalLike: number
  totalShare: number
  totalVisit: number
}

export function getArticleAnalyticsTrend(period: string) {
  return request.get('/analytics/article-trend', { params: { period } })
}
