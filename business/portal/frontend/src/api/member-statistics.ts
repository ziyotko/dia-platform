import request from '@/utils/request'

export interface MemberStatisticsDataset {
  id: number
  memberColumnId: number
  statisticsYear: number
  comparisonYear: number
  title: string
  sourceFileName: string
  sourceSha256: string
  importedAt: string
  status: number
}

export interface MemberStatisticsSummary {
  recordCount: number
  enterpriseCount: number
  provinceCount: number
  revenueYearWan: string
  revenuePriorWan: string
  totalYoyPercent: string
  missingRdYear: number
  missingRdPrior: number
}

export interface MemberStatisticsDuplicate {
  enterpriseName: string
  sourceRows: number[]
}

export interface MemberStatisticsDetail {
  dataset: MemberStatisticsDataset
  columnName: string
  summary: MemberStatisticsSummary
  duplicateNames: MemberStatisticsDuplicate[]
  filters: { provinces: string[]; regions: string[]; subFields: string[] }
}

export interface MemberStatisticsRow {
  id: number
  sourceRow: number
  enterpriseName: string
  province: string
  region: string
  subField: string
  isBelt: number
  isAxis: number
  revenueYearWan: string
  revenuePriorWan: string
  yoySource: string | null
  rdYearWan: string | null
  rdPriorWan: string | null
  rdShareYear: string | null
  rdSharePrior: string | null
  source: string
}

export interface MemberStatisticsRowQuery {
  enterpriseName?: string
  province?: string
  region?: string
  subField?: string
  missingRd?: number
  page: number
  pageSize: number
}

export const getMemberStatisticsDatasets = () => request.get('/member-statistics/datasets')

export const getMemberStatisticsDataset = (id: number) =>
  request.get(`/member-statistics/datasets/${id}`)

export const getMemberStatisticsRows = (id: number, params: MemberStatisticsRowQuery) =>
  request.get(`/member-statistics/datasets/${id}/rows`, { params })

export const updateMemberStatisticsStatus = (id: number, status: 1 | 2) =>
  request.patch(`/member-statistics/datasets/${id}/status`, { status })
