import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('application-member-token') || '')
  const userInfo = ref<any>(null)

  const isLoggedIn = computed(() => !!token.value)

  function setToken(val: string) {
    token.value = val
    localStorage.setItem('application-member-token', val)
  }

  async function login(username: string, password: string, captchaId: string, captchaCode: string) {
    const res = await authApi.userLogin({ username, password, captcha_id: captchaId, captcha_code: captchaCode })
    setToken(res.data.token)
    userInfo.value = res.data.user
    return res.data
  }

  async function fetchUserInfo() {
    const res = await authApi.getUserProfile()
    userInfo.value = res.data
  }

  function clearSession() {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('application-member-token')
  }

  // 本地强制登出：Token 已失效（401）时使用；不再请求服务端，避免二次 401 递归
  function forceLogout() {
    clearSession()
    window.location.href = `${import.meta.env.BASE_URL || '/'}login`
  }

  // 主动退出登录（用户点「退出登录」/改密后）：先让服务端把当前 Token 拉黑
  // （TTL = 剩余有效期），再清理本地会话。服务端失败不阻塞本地登出。
  async function logout() {
    if (token.value) {
      try {
        await authApi.userLogout()
      } catch {
        // 忽略：服务端登出失败也要完成本地登出
      }
    }
    forceLogout()
  }

  return { token, userInfo, isLoggedIn, setToken, login, fetchUserInfo, clearSession, forceLogout, logout }
})
