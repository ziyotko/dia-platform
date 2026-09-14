import axios from 'axios'
import CryptoJS from 'crypto-js'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

// 扩展 axios 请求配置：raw 模式跳过统一响应校验，原样返回整个响应（含状态码）
declare module 'axios' {
  export interface AxiosRequestConfig<D = any> {
    raw?: boolean
  }
}

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

function createRequestNonce() {
  if (crypto.randomUUID) {
    return crypto.randomUUID()
  }

  const values = new Uint32Array(4)
  crypto.getRandomValues(values)
  return Array.from(values, (value) => value.toString(16).padStart(8, '0')).join('')
}

function sha256(message: string): string {
  return CryptoJS.SHA256(message).toString(CryptoJS.enc.Hex)
}

function hmacSha256(message: string, secret: string): string {
  return CryptoJS.HmacSHA256(message, secret).toString(CryptoJS.enc.Hex)
}

// 签名密钥由访问 Token 派生（与后端 utils/sign.go 的 DeriveSignKey 保持一致），
// 不再单独存储 signKey，降低密钥暴露面并随 Token 轮换。
function deriveSignKey(token: string): string {
  return hmacSha256(token, 'portal-replay-sign-v1')
}

// 计算最终请求的 target（path + query），必须与后端 c.Request.URL.RequestURI() 一致，
// 从而让 GET 查询参数也受到签名保护。
function getRequestTarget(config: any): string {
  try {
    // axios.getUri 会按 baseURL/url/params 序列化出最终会发送的完整 URL
    const fullUrl = axios.getUri(config)
    const parsed = new URL(fullUrl, 'http://local.invalid')
    return parsed.pathname + parsed.search
  } catch {
    const url = config.url || ''
    if (url.startsWith('http')) {
      const parsed = new URL(url)
      return parsed.pathname + parsed.search
    }
    const baseURL = (config.baseURL || '').replace(/\/+$/, '')
    const path = url.startsWith('/') ? url : '/' + url
    return (baseURL + path).split('?')[0]
  }
}

function getBodyString(body: unknown): string {
  if (body === undefined || body === null) return ''
  if (typeof body === 'string') return body
  if (body instanceof FormData) return ''
  if (Array.isArray(body) && body.length === 0) return ''
  if (typeof body === 'object' && Object.keys(body).length === 0) return ''
  return JSON.stringify(body)
}

// 将 ArrayBuffer 转为 crypto-js 的 WordArray，按字节正确计算哈希
function toWordArray(arrayBuffer: ArrayBuffer): any {
  const uint8 = new Uint8Array(arrayBuffer)
  const words: number[] = []
  for (let i = 0; i < uint8.length; i += 4) {
    words.push(
      (uint8[i] << 24) |
        ((uint8[i + 1] || 0) << 16) |
        ((uint8[i + 2] || 0) << 8) |
        (uint8[i + 3] || 0)
    )
  }
  return CryptoJS.lib.WordArray.create(words, uint8.length)
}

// 分块读取 Blob 并累加进哈希，避免大文件一次性进入内存阻塞主线程
async function hashBlob(blob: Blob, hasher: any): Promise<void> {
  const chunkSize = 1024 * 1024 // 1MB
  for (let offset = 0; offset < blob.size; offset += chunkSize) {
    const chunk = blob.slice(offset, offset + chunkSize)
    const buf = await chunk.arrayBuffer()
    hasher.update(toWordArray(buf))
  }
}

// 仅对小于阈值（默认 50MB）的上传计算文件内容哈希；超大文件回退为空串哈希，兼容旧逻辑
const MAX_MULTIPART_HASH_BYTES = 50 * 1024 * 1024

async function computeFormDataHash(formData: FormData): Promise<string> {
  const hasher = CryptoJS.algo.SHA256.create()
  let total = 0
  let hasFile = false
  for (const [, value] of formData.entries()) {
    if (value instanceof Blob) {
      hasFile = true
      total += value.size
      if (total > MAX_MULTIPART_HASH_BYTES) {
        return sha256('')
      }
      await hashBlob(value, hasher)
    }
  }
  if (!hasFile) return sha256('')
  return hasher.finalize().toString(CryptoJS.enc.Hex)
}

// 计算请求体哈希；multipart 时返回文件内容哈希，并在 fileHash 中回传用于上报 X-Body-Hash-Value
async function computeBodyHash(body: unknown): Promise<{ hash: string; fileHash?: string }> {
  if (body instanceof FormData) {
    const fileHash = await computeFormDataHash(body)
    return fileHash !== sha256('') ? { hash: fileHash, fileHash } : { hash: fileHash }
  }
  return { hash: sha256(getBodyString(body)) }
}

function createRequestSignature(
  signKey: string,
  method: string,
  target: string,
  timestamp: string,
  nonce: string,
  bodyHash: string
): string {
  const payload = `${method.toUpperCase()}|${target}|${timestamp}|${nonce}|${bodyHash}`
  return hmacSha256(payload, signKey)
}

request.interceptors.request.use(
  async (config) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }

    const timestamp = Date.now().toString()
    const nonce = createRequestNonce()
    config.headers['X-Request-Timestamp'] = timestamp
    config.headers['X-Request-Nonce'] = nonce

    const signKey = userStore.token ? deriveSignKey(userStore.token) : ''
    if (signKey && config.url) {
      const target = getRequestTarget(config)
      const { hash, fileHash } = await computeBodyHash(config.data)
      if (fileHash) {
        config.headers['X-Body-Hash-Value'] = fileHash
      }
      const signature = createRequestSignature(
        signKey,
        config.method || 'GET',
        target,
        timestamp,
        nonce,
        hash
      )
      config.headers['X-Request-Signature'] = signature
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 从响应体中提取业务错误信息（兼容对象 / JSON 字符串 / 纯文本）
function extractBizMessage(data: any): string {
  let parsed = data
  if (typeof data === 'string') {
    const trimmed = data.trim()
    if (!trimmed) return ''
    try {
      parsed = JSON.parse(trimmed)
    } catch {
      return trimmed
    }
  }
  if (parsed && typeof parsed === 'object') {
    const msg =
      parsed.message ||
      parsed.msg ||
      parsed.error ||
      parsed.detail ||
      parsed.errorMessage ||
      parsed.errmsg
    if (typeof msg === 'string' && msg.trim()) return msg.trim()
  }
  return ''
}

// 从 axios 错误中提取可读的具体错误信息，尽量展示后端返回的真实原因，而非笼统的“网络错误”
export function getErrorMessage(error: any): string {
  // 1. 优先取响应体中的业务错误字段（message/msg/error/纯文本等）
  const data = error?.response?.data
  if (data != null) {
    const msg = extractBizMessage(data)
    if (msg) return msg
  }

  // 2. HTTP 状态错误：补充状态文本，例如 401 Unauthorized
  const status = error?.response?.status
  if (status) {
    const statusText = error?.response?.statusText
    return statusText ? `${statusText}（${status}）` : `请求失败（HTTP ${status}）`
  }

  // 3. 请求超时 / 其他 axios 层错误
  if (error?.code === 'ECONNABORTED') return '请求超时，请稍后重试'
  if (error?.message && error.message !== 'Network Error') return error.message

  // 4. 兜底
  return '网络错误'
}

request.interceptors.response.use(
  (response) => {
    // raw 模式：跳过统一响应校验，原样返回整个 axios 响应（含 status），由调用方自行判断
    if ((response.config as any).raw) {
      return response
    }
    const res = response.data
    // 业务码只有三种：0=成功、1=失败、401=鉴权失效（后端不会把 HTTP 200 当业务成功码返回）
    if (res.code !== 0) {
      ElMessage.error(res.message || '请求失败')
      if (res.code === 401) {
        const userStore = useUserStore()
        userStore.logout()
        // 跳转到登录页：必须带上部署子路径（VITE_BASE_PATH，如 /caamm/），
        // 否则在子路径部署下会跳到不存在的一级路径 /login。
        const base = import.meta.env.BASE_URL || '/'
        window.location.href = `${base.endsWith('/') ? base : `${base}/`}login`
      }
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res
  },
  (error) => {
    // raw 模式：错误信息由调用方自行展示，避免重复弹窗；轮询等场景也不应刷屏
    const isRaw = !!(error?.config as any)?.raw
    if (!isRaw) {
      ElMessage.error(getErrorMessage(error))
    }
    return Promise.reject(error)
  }
)

export default request
