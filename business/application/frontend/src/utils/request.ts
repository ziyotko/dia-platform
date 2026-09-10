import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/application/api',
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
    const res = response.data
    if (res.code !== 0 && res.code !== 200) {
      ElMessage.error(res.message || '请求失败')
      if (res.code === 401) {
        localStorage.clear()
        window.location.href = '/application/login'
      }
      return Promise.reject(new Error(res.message))
    }
    return res
  },
  (error) => {
    ElMessage.error(error.message || '网络错误')
    return Promise.reject(error)
  }
)

export default request
