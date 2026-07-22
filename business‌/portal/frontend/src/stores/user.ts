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
    const data = JSON.parse(stored) as any
    // 兼容后端旧数据：ID 大写转小写
    if (data.ID !== undefined && data.id === undefined) {
      data.id = data.ID
    }
    return data as UserInfo
  } catch {
    return null
  }
}

const SIGN_KEY_KEY = 'sign_key'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const signKey = ref<string>(localStorage.getItem(SIGN_KEY_KEY) || '')
  const userInfo = ref<UserInfo | null>(getStoredUserInfo())
  const menuList = ref<MenuItem[]>([])
  const hasFetchedMenus = ref(false)

  const isLoggedIn = computed(() => !!token.value)

  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  function setSignKey(newSignKey: string) {
    signKey.value = newSignKey
    localStorage.setItem(SIGN_KEY_KEY, newSignKey)
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
    signKey.value = ''
    userInfo.value = null
    menuList.value = []
    hasFetchedMenus.value = false
    localStorage.removeItem('token')
    localStorage.removeItem(SIGN_KEY_KEY)
    localStorage.removeItem(USER_INFO_KEY)
    localStorage.removeItem('app-theme')
    localStorage.removeItem('app-security')
  }

  return {
    token,
    signKey,
    userInfo,
    menuList,
    hasFetchedMenus,
    isLoggedIn,
    setToken,
    setSignKey,
    setUserInfo,
    fetchUserMenusAndGenerateRoutes,
    resetDynamicRoutes,
    logout
  }
})
