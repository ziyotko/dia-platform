import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import router from '@/router'
import { authApi } from '@/api/auth'

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

  function logout() {
    token.value = ''
    userInfo.value = null
    menus.value = []
    localStorage.clear()
    sessionStorage.clear()
    window.location.href = '/member/login'
  }

  return { token, userInfo, menus, isLoggedIn, isAdmin, setToken, login, fetchUserInfo, logout }
})
