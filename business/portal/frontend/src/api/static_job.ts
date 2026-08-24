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
