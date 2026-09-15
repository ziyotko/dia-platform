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
    try {
      const res: any = await authApi.getUserMenus()
      menus.value = res.data || []
      return menus.value
    } finally {
      // 无论成功与否都标记为已拉取，避免路由守卫因异常反复重试
      hasFetchedMenus.value = true
    }
  }

  async function fetchPermissions() {
    const res: any = await authApi.getUserPermissions()
    permissions.value = res.data || []
    return res.data
  }

  /**
   * 是否拥有某个权限标识（如 base:user:delete）。
   * 与后端 PermissionAuth 口径保持一致：平台超管 / 租户管理员直接放行，其余按授权码判断。
   */
  function can(code: string) {
    const info = userInfo.value
    if (!info) return false
    if (info.tenantId === 0 || info.isAdmin) return true
    return permissions.value.includes(code)
  }

  /** 仅清空本地会话，不做跳转（供路由守卫使用） */
  function clearSession() {
    token.value = ''
    userInfo.value = null
    menus.value = []
    permissions.value = []
    hasFetchedMenus.value = false
    localStorage.removeItem('base-token')
  }

  function logout() {
    clearSession()
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
    can,
    clearSession,
    logout
  }
})
