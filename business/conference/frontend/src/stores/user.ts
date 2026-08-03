import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('conference-member-token') || '')
  const userInfo = ref<any>(null)

  const isLoggedIn = computed(() => !!token.value)

  function setToken(val: string) {
    token.value = val
    localStorage.setItem('conference-member-token', val)
  }

  async function login(username: string, password: string, captchaId: string, captchaCode: string) {
    const res = await authApi.memberLogin({ username, password, captchaId, captchaCode })
    setToken(res.data.token)
    userInfo.value = res.data.user
    return res.data
  }

  async function fetchUserInfo() {
    const res = await authApi.getMemberProfile()
    userInfo.value = res.data
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('conference-member-token')
    window.location.href = '/conference/login'
  }

  return { token, userInfo, isLoggedIn, setToken, login, fetchUserInfo, logout }
})
