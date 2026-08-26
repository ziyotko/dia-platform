import request from '@/utils/request'

export type StaticJobKind = 'site' | 'pages' | 'lists' | 'articles'

export interface StaticJobProgress {
  stage: string
  processed: number
  total: number
  generated_files: number
}

export type StaticJobStatus = 'queued' | 'running' | 'succeeded' | 'failed' | 'interrupted'

export interface StaticJob {
  id: string
  kind: StaticJobKind
  status: StaticJobStatus
  status_url: string
  cancel_url: string
  progress?: StaticJobProgress
  created_at?: string
  started_at?: string
  updated_at?: string
}

export interface StaticJobResponse {
  ok: boolean
  job?: StaticJob
}

// 发起静态化批量操作任务
// 后端代理转发至静态化程序，透传其 202 状态码与响应体（代表请求已发送，请等待处理结果）
// gray 仅对全站静态化 / 生成首页生效：1 开启首页整体变灰，2 关闭
export async function startStaticJob(
  kind: StaticJobKind,
  gray?: number
): Promise<{ status: number; data: StaticJobResponse }> {
  const params: Record<string, number> = {}
  if (kind === 'site' || kind === 'pages') {
    params.gray = gray ?? 2
  }
  // 注意：不传请求体（undefined），避免 axios 将 null 序列化为 "null" 导致与空请求体签名不一致
  const res: any = await request.post(`/static/${kind}`, undefined, { params, raw: true } as any)
  return { status: res.status, data: res.data }
}

// 查询任务状态：返回 HTTP 200 及任务结构（仅发起任务时返回 202）
// queued 等待执行 / running 正在执行 / succeeded 执行成功 / failed 执行失败 / interrupted 被取消或超时中断
export async function getStaticJob(id: string): Promise<{ status: number; data: StaticJobResponse }> {
  const res: any = await request.get(`/static/jobs/${id}`, { raw: true } as any)
  return { status: res.status, data: res.data }
}

// 首页重新生成的生成结果（静态化程序同步返回）
export interface StaticPageResult {
  generated_at: string
  duration_seconds: number
  page: string
  total_items: number
  generated_details: number
  generated_lists: number
  output: string
  gray: string
}

export interface StaticPageResponse {
  ok: boolean
  result?: StaticPageResult
  message?: string
  msg?: string
}

// 单页操作-首页重新生成：POST /static/page?name={页面名}
// 后端代理转发至静态化程序（自动附带 Authorization 验证令牌头），同步返回 HTTP 200 及生成结果；
// 输出目录与首页整体变灰由后端全局变量决定，前端仅需传页面名。
export async function startStaticPage(name: string): Promise<{ status: number; data: StaticPageResponse }> {
  const res: any = await request.post('/static/page', undefined, { params: { name }, raw: true } as any)
  return { status: res.status, data: res.data }
}

// 单页操作-栏目页重新生成：POST /static/list?column_name={栏目名称}
// 后端代理转发至静态化程序（自动附带 Authorization 验证令牌头），同步返回 HTTP 200 及生成结果；
// 输出目录由后端全局变量决定，前端仅需传栏目名称。
export async function startStaticList(columnName: string): Promise<{ status: number; data: StaticPageResponse }> {
  const res: any = await request.post('/static/list', undefined, { params: { column_name: columnName }, raw: true } as any)
  return { status: res.status, data: res.data }
}

// 单页操作-详情页重新生成：POST /static/article?id={文章ID}
// 后端代理转发至静态化程序（自动附带 Authorization 验证令牌头），同步返回 HTTP 200 及生成结果；
// 输出目录由后端全局变量决定，前端仅需传文章ID。
export async function startStaticArticle(id: number | string): Promise<{ status: number; data: StaticPageResponse }> {
  const res: any = await request.post('/static/article', undefined, { params: { id }, raw: true } as any)
  return { status: res.status, data: res.data }
}
