import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/business_base/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

request.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    // 文件下载/导出等非统一响应体：直接返回原始数据，跳过 {code} 校验
    if (response.config.responseType === 'blob' || (response.config as any).raw) {
      return (response.config as any).raw ? response : response.data
    }
    const res = response.data
    if (res.code !== 0 && res.code !== 200) {
      ElMessage.error(res.message || '请求失败')
      if (res.code === 401) {
        // 只清本地会话（不要调 logout 接口：token 已失效会再次 401，造成递归请求）
        const userStore = useUserStore()
        userStore.clearSession()
        // 跳登录页必须带上部署子路径（BASE_URL = vite base），否则子路径部署下会跳到不存在的 /login
        window.location.href = `${import.meta.env.BASE_URL || '/'}login`
      }
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res
  },
  (error) => {
    // raw / 下载请求由调用方自行处理错误提示，避免重复弹窗
    const config: any = error.config || {}
    if (!config.raw && config.responseType !== 'blob') {
      ElMessage.error(error.response?.data?.message || '网络错误')
    }
    return Promise.reject(error)
  }
)

export default request
