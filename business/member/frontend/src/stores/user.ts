import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import { currentReturnUrl, loginPathWithReturnUrl } from '@/utils/returnUrl'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('member-token') || '')
  const userInfo = ref<any>(null)
  const menus = ref<any[]>([])

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => userInfo.value?.is_admin || false)

  function setToken(val: string) {
    token.value = val
    localStorage.setItem('member-token', val)
  }

  async function login(username: string, password: string, captchaId: string, captchaCode: string) {
    const res = await authApi.login({ username, password, captcha_id: captchaId, captcha_code: captchaCode })
    setToken(res.data.token)
    userInfo.value = res.data.member
    return res.data
  }

  async function fetchUserInfo() {
    const res = await authApi.getProfile()
    userInfo.value = res.data
  }

  function clearSession() {
    token.value = ''
    userInfo.value = null
    menus.value = []
    // 只清本应用自己的键：member/portal/application/base 同域子路径部署，
    // localStorage.clear() 会连带把其它系统的登录态一起清掉
    localStorage.removeItem('member-token')
    sessionStorage.removeItem('member-token')
  }

  // 本地强制登出：Token 已失效（401）时使用；不再请求服务端，避免二次 401 递归
  function forceLogout() {
    clearSession()
    // 跳登录页必须带上部署子路径（BASE_URL = vite base），否则子路径部署下会跳到不存在的 /login；
    // 同时带上当前页面，重新登录后回到原位置（与路由守卫同一套 returnUrl 口径）
    window.location.href = loginPathWithReturnUrl(currentReturnUrl())
  }

  // 主动退出登录（用户点「退出登录」/改密后）：先让服务端把当前 Token 拉黑
  // （TTL = 剩余有效期），再清理本地会话。服务端失败不阻塞本地登出。
  async function logout() {
    if (token.value) {
      try {
        await authApi.logout()
      } catch {
        // 忽略：服务端登出失败也要完成本地登出
      }
    }
    forceLogout()
  }

  return { token, userInfo, menus, isLoggedIn, isAdmin, setToken, login, fetchUserInfo, clearSession, forceLogout, logout }
})
