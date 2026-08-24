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

function getRequestPath(config: any): string {
  let url = config.url || ''
  if (url.startsWith('http')) {
    url = new URL(url).pathname
  } else {
    const baseURL = (config.baseURL || '').replace(/\/+$/, '')
    const path = url.startsWith('/') ? url : '/' + url
    url = baseURL + path
  }
  // 去掉 query string，只签 path
  return url.split('?')[0]
}

function getBodyString(body: unknown): string {
  if (body === undefined || body === null) return ''
  if (typeof body === 'string') return body
  if (body instanceof FormData) return ''
  if (Array.isArray(body) && body.length === 0) return ''
  if (typeof body === 'object' && Object.keys(body).length === 0) return ''
  return JSON.stringify(body)
}

function createRequestSignature(
  signKey: string,
  method: string,
  path: string,
  timestamp: string,
  nonce: string,
  body: unknown
): string {
  const bodyString = getBodyString(body)
  const bodyHash = sha256(bodyString)
  const payload = `${method.toUpperCase()}|${path}|${timestamp}|${nonce}|${bodyHash}`
  return hmacSha256(payload, signKey)
}

request.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }

    const timestamp = Date.now().toString()
    const nonce = createRequestNonce()
    config.headers['X-Request-Timestamp'] = timestamp
    config.headers['X-Request-Nonce'] = nonce

    if (userStore.signKey && config.url) {
      const path = getRequestPath(config)
      const signature = createRequestSignature(
        userStore.signKey,
        config.method || 'GET',
        path,
        timestamp,
        nonce,
        config.data
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
    if (res.code !== 0 && res.code !== 200) {
      ElMessage.error(res.message || '请求失败')
      if (res.code === 401) {
        const userStore = useUserStore()
        userStore.logout()
        window.location.href = '/login'
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
