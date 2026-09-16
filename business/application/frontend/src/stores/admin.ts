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

  function logout() {
    token.value = ''
    adminInfo.value = null
    localStorage.removeItem('application-admin-token')
    localStorage.removeItem('application-admin-role')
    window.location.href = `${import.meta.env.BASE_URL || '/'}admin/login`
  }

  return { token, adminInfo, isLoggedIn, roleCode, setToken, login, fetchAdminInfo, logout }
})
