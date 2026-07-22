import axios from 'axios'
import CryptoJS from 'crypto-js'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

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

request.interceptors.response.use(
  (response) => {
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
    ElMessage.error(error.response?.data?.message || '网络错误')
    return Promise.reject(error)
  }
)

export default request
