import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getUserMenus, type MenuItem } from '@/api/menus'

export interface UserInfo {
  id: number
  username: string
  nickname: string
  avatar: string
  roleIds: number[]
}

const USER_INFO_KEY = 'user_info'

function getStoredUserInfo(): UserInfo | null {
  const stored = localStorage.getItem(USER_INFO_KEY)
  if (!stored) return null
  try {
    return JSON.parse(stored) as UserInfo
  } catch {
    return null
  }
}

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo | null>(getStoredUserInfo())
  const menuList = ref<MenuItem[]>([])
  const hasFetchedMenus = ref(false)

  const isLoggedIn = computed(() => !!token.value)

  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  function setUserInfo(info: UserInfo) {
    userInfo.value = info
    localStorage.setItem(USER_INFO_KEY, JSON.stringify(info))
  }

  /**
   * 获取用户菜单并动态生成路由
   */
  async function fetchUserMenusAndGenerateRoutes() {
    const { addDynamicRoutes } = await import('@/router')
    const res: any = await getUserMenus()
    const menus = res.data || []
    menuList.value = menus
    addDynamicRoutes(menus)
    hasFetchedMenus.value = true
  }

  /**
   * 重置动态路由状态（用于重新登录等场景）
   */
  function resetDynamicRoutes() {
    menuList.value = []
    hasFetchedMenus.value = false
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    menuList.value = []
    hasFetchedMenus.value = false
    localStorage.removeItem('token')
    localStorage.removeItem(USER_INFO_KEY)
    localStorage.removeItem('app-theme')
  }

  return {
    token,
    userInfo,
    menuList,
    hasFetchedMenus,
    isLoggedIn,
    setToken,
    setUserInfo,
    fetchUserMenusAndGenerateRoutes,
    resetDynamicRoutes,
    logout
  }
})
