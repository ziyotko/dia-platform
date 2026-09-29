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

// 后端所有响应都是 HTTP 200（业务码在 body 里）：blob 请求失败时 body 其实是 {code,message} 的 JSON。
// 读出并返回该业务错误；不是业务错误（真的是文件流）时返回 null。
async function readBlobBizError(blob: Blob): Promise<{ code: number; message: string } | null> {
  try {
    const parsed = JSON.parse(await blob.text())
    if (parsed && typeof parsed.code === 'number' && parsed.code !== 0) {
      return { code: parsed.code, message: parsed.message || '请求失败' }
    }
  } catch {
    // 不是 JSON：视为正常文件流
  }
  return null
}

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
  async (response) => {
    // 文件下载/导出等非统一响应体：直接返回原始数据，跳过 {code} 校验。
    // 但后端出错时仍是 HTTP 200 + {code,message} 的 JSON，只是被 axios 当 Blob 返回：
    // 必须读出来判断，否则错误 JSON 会被当成 CSV/PDF 下载并提示「导出成功」。
    if (response.config.responseType === 'blob') {
      const blob = response.data as Blob
      if (blob instanceof Blob && blob.type.includes('json')) {
        const biz = await readBlobBizError(blob)
        if (biz) {
          ElMessage.error(biz.message)
          if (biz.code === 401) {
            useUserStore().forceLogout()
          }
          return Promise.reject(new Error(biz.message))
        }
      }
      return response.data
    }
    const res = response.data
    if (res.code !== 0 && res.code !== 200) {
      ElMessage.error(res.message || '请求失败')
      if (res.code === 401) {
        const userStore = useUserStore()
        userStore.forceLogout()
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
    if (status === 401) {
      msg = '登录状态已失效，请重新登录'
      useUserStore().forceLogout()
    } else if (status === 403) {
      // 403 只能来自反向代理/静态服务（本项目的无权限是 HTTP 200 + code=1），
      // 不能当成登录态失效把用户踢回登录页
      msg = '没有访问权限'
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
