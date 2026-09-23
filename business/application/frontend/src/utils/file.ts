import { ElMessage } from 'element-plus'
import request from '@/utils/request'

// 上传文件不再对外公开（后端已移除静态托管）：预览/下载统一走带鉴权的接口
// （申报人 /member/files，管理端 /admin/files），前端拿 Blob 后自己造 objectURL。
// 同理，审计日志导出也是文件流（/admin/audit-logs/export）。
//
// 注意：这些请求都带 responseType: 'blob'，request.ts 里的统一 code 校验会被
// 跳过（响应体不是 JSON），错误处理在本文件内完成。

type FileSide = 'member' | 'admin'

async function openBlob(endpoint: string, params: Record<string, unknown>, fallbackName: string) {
  try {
    const res: any = await request.get(endpoint, { params, responseType: 'blob' })
    const blob = res.data as Blob
    // 后端出错时返回的仍是统一 JSON 结构（HTTP 200 + application/json）
    if (blob.type && blob.type.includes('json')) {
      const text = await blob.text()
      let msg = '文件不存在或无权访问'
      try {
        msg = JSON.parse(text)?.message || msg
      } catch {
        /* 保持默认提示 */
      }
      ElMessage.error(msg)
      return
    }
    const objectUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = objectUrl
    a.target = '_blank'
    if (blob.type && !blob.type.includes('pdf') && !blob.type.includes('image')) {
      a.download = fallbackName
    }
    a.click()
    // 稍后释放，避免刚打开就被回收
    setTimeout(() => URL.revokeObjectURL(objectUrl), 60_000)
  } catch (e: any) {
    ElMessage.error(e?.message || '文件下载失败')
  }
}

/** 打开/下载材料或证书附件（按当前身份选择鉴权接口） */
export function openFile(fileUrl?: string, side: FileSide = 'member', name = 'file') {
  if (!fileUrl) return
  if (/^https?:\/\//i.test(fileUrl)) {
    window.open(fileUrl, '_blank')
    return
  }
  openBlob(`/${side}/files`, { url: fileUrl }, name)
}

/** 导出审计日志 CSV（走后端鉴权接口，按当前筛选条件） */
export function exportAuditLogs(params: { module?: string; keyword?: string }) {
  openBlob('/admin/audit-logs/export', params, 'audit-logs.csv')
}
