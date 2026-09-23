import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

const request = axios.create({
  // fallback 必须与 .env 的 VITE_API_BASE_URL、vite.config.ts 的默认值、后端 api_prefix 一致，
  // 否则 .env 未注入时（如直接在 shell 里 build）全部请求 404
  baseURL: import.meta.env.VITE_API_BASE_URL || '/business_member/api',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' }
})

request.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

request.interceptors.response.use(
  (response) => {
    // 文件下载/导出等非统一响应体：直接返回原始数据，跳过 {code} 校验
    if (response.config.responseType === 'blob') {
      return response.data
    }
    const res = response.data
    if (res.code !== 0 && res.code !== 200) {
      ElMessage.error(res.message || '请求失败')
      if (res.code === 401) {
        const userStore = useUserStore()
        userStore.logout()
      }
      return Promise.reject(new Error(res.message))
    }
    return res
  },
  (error) => {
    // 业务错误统一走 HTTP 200 + code（上面的成功分支处理），
    // 这里只处理网关/网络层异常，避免直接抛出英文 “Request failed with status code xxx”
    const status = error?.response?.status
    let msg = error?.message || '网络错误'
    if (status === 401 || status === 403) {
      msg = '登录状态已失效，请重新登录'
      useUserStore().logout()
    } else if (status === 413) {
      msg = '上传内容过大，请压缩后重试'
    } else if (status === 429) {
      msg = '请求过于频繁，请稍后再试'
    } else if (typeof status === 'number' && status >= 500) {
      msg = '服务暂时不可用，请稍后重试'
    } else if (!error?.response) {
      msg = '网络异常，请检查网络连接后重试'
    }
    ElMessage.error(msg)
    return Promise.reject(error)
  }
)

export default request
