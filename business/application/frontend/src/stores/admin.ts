import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'

export const useAdminStore = defineStore('admin', () => {
  const token = ref<string>(localStorage.getItem('application-admin-token') || '')
  const adminInfo = ref<any>(null)

  const isLoggedIn = computed(() => !!token.value)
  const roleCode = computed(() => adminInfo.value?.roleCode || '')

  function setToken(val: string) {
    token.value = val
    localStorage.setItem('application-admin-token', val)
  }

  async function login(username: string, password: string, captchaId: string, captchaCode: string) {
    const res = await authApi.adminLogin({ username, password, captchaId, captchaCode })
    setToken(res.data.token)
    adminInfo.value = res.data.admin
    return res.data
  }

  async function fetchAdminInfo() {
    const res = await authApi.getAdminProfile()
    adminInfo.value = res.data
  }

  function logout() {
    token.value = ''
    adminInfo.value = null
    localStorage.removeItem('application-admin-token')
    window.location.href = '/application/admin/login'
  }

  return { token, adminInfo, isLoggedIn, roleCode, setToken, login, fetchAdminInfo, logout }
})
