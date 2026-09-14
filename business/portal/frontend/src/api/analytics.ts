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

export function getArticleAnalyticsTrend(period: string, year?: number) {
  return request.get('/analytics/article-trend', { params: { period, year } })
}
