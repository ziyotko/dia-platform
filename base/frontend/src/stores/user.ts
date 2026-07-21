import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as authApi from '@/api/auth'
import type { Menu } from '@/api/menu'
import router from '@/router'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('base-token') || '')
  const userInfo = ref<any>(null)
  const menus = ref<Menu[]>([])
  const permissions = ref<string[]>([])
  const hasFetchedMenus = ref(false)

  const isLoggedIn = computed(() => !!token.value)

  function setToken(val: string) {
    token.value = val
    localStorage.setItem('base-token', val)
  }

  async function login(data: authApi.LoginReq) {
    const res: any = await authApi.login(data)
    setToken(res.data.token)
    userInfo.value = res.data.user
    return res
  }

  async function fetchUserInfo() {
    const res: any = await authApi.getUserInfo()
    userInfo.value = res.data
    return res.data
  }

  async function fetchUserMenusAndGenerateRoutes() {
    const res: any = await authApi.getUserMenus()
    menus.value = res.data || []
    hasFetchedMenus.value = true
    return res.data
  }

  async function fetchPermissions() {
    const res: any = await authApi.getUserPermissions()
    permissions.value = res.data || []
    return res.data
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    menus.value = []
    permissions.value = []
    hasFetchedMenus.value = false
    localStorage.removeItem('base-token')
    router.push('/login')
  }

  return {
    token,
    userInfo,
    menus,
    permissions,
    hasFetchedMenus,
    isLoggedIn,
    setToken,
    login,
    fetchUserInfo,
    fetchUserMenusAndGenerateRoutes,
    fetchPermissions,
    logout
  }
})
