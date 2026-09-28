import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'

export const useAdminStore = defineStore('admin', () => {
  const token = ref<string>(localStorage.getItem('application-admin-token') || '')
  const roleCode = ref<string>(localStorage.getItem('application-admin-role') || '')
  const adminInfo = ref<any>(null)

  const isLoggedIn = computed(() => !!token.value)

  function setToken(val: string) {
    token.value = val
    localStorage.setItem('application-admin-token', val)
  }

  function setRole(val: string) {
    roleCode.value = val || ''
    if (val) localStorage.setItem('application-admin-role', val)
    else localStorage.removeItem('application-admin-role')
  }

  function setAdminInfo(info: any) {
    adminInfo.value = info
    setRole(info?.roleCode || '')
  }

  async function login(username: string, password: string, captchaId: string, captchaCode: string) {
    const res = await authApi.adminLogin({ username, password, captcha_id: captchaId, captcha_code: captchaCode })
    setToken(res.data.token)
    setAdminInfo(res.data.admin)
    return res.data
  }

  async function fetchAdminInfo() {
    const res = await authApi.getAdminProfile()
    setAdminInfo(res.data)
  }

  function clearSession() {
    token.value = ''
    adminInfo.value = null
    roleCode.value = ''
    localStorage.removeItem('application-admin-token')
    localStorage.removeItem('application-admin-role')
  }

  // 本地强制登出：Token 已失效（401）时使用（request.ts 的 401 分支也会走同一条清理路径）
  function forceLogout() {
    clearSession()
    window.location.href = `${import.meta.env.BASE_URL || '/'}admin/login`
  }

  // 主动退出登录：先让服务端把当前 Token 拉黑（TTL = 剩余有效期），再清理本地会话
  async function logout() {
    if (token.value) {
      try {
        await authApi.adminLogout()
      } catch {
        // 忽略：服务端登出失败也要完成本地登出
      }
    }
    forceLogout()
  }

  return { token, adminInfo, isLoggedIn, roleCode, setToken, login, fetchAdminInfo, clearSession, forceLogout, logout }
})
