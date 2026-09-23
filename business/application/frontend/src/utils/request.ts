import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
  // 兜底前缀与 config.yaml 的 api_prefix 保持一致（子路径部署时以 .env 的 VITE_API_BASE_URL 为准）
  baseURL: import.meta.env.VITE_API_BASE_URL || '/business_application/api',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' }
})

request.interceptors.request.use((config) => {
  const memberToken = localStorage.getItem('application-member-token')
  const adminToken = localStorage.getItem('application-admin-token')
  const url = config.url || ''

  if (url.includes('/admin/') && adminToken) {
    config.headers.Authorization = `Bearer ${adminToken}`
  } else if (memberToken) {
    config.headers.Authorization = `Bearer ${memberToken}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    // 文件流（responseType: 'blob'）没有统一响应结构：直接返回整个响应，
    // 由 utils/file.ts 判断是文件还是被包装成 JSON 的错误
    if (response.config?.responseType === 'blob') {
      return response as any
    }
    const res = response.data
    if (res.code !== 0 && res.code !== 200) {
      ElMessage.error(res.message || '请求失败')
      if (res.code === 401) {
        // 按发起请求的角色清理与跳转：管理端 token 过期必须回管理端登录页，
        // 且只能删除本应用的键——localStorage.clear() 会把同域部署的 portal/member/base 登录态一起清掉。
        const url = response.config?.url || ''
        const isAdmin = url.includes('/admin/')
        const keys = isAdmin
          ? ['application-admin-token', 'application-admin-role']
          : ['application-member-token']
        keys.forEach((key) => localStorage.removeItem(key))
        // 跳登录页必须带上部署子路径（BASE_URL = vite base），否则子路径部署下会跳到不存在的 /login
        window.location.href = `${import.meta.env.BASE_URL || '/'}${isAdmin ? 'admin/login' : 'login'}`
      }
      return Promise.reject(new Error(res.message))
    }
    return res
  },
  (error) => {
    // blob 请求失败时 axios 给出的是 Blob 错误体，提示统一放到 utils/file.ts
    if (error.config?.responseType === 'blob') {
      return Promise.reject(error)
    }
    ElMessage.error(error.message || '网络错误')
    return Promise.reject(error)
  }
)

export default request
