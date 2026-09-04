import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getUserMenus, type MenuItem } from '@/api/menus'

export interface UserInfo {
  id: number
  username: string
  avatar: string
  roleIds: number[]
}

const USER_INFO_KEY = 'user_info'

// 兼容 roleIds 可能为逗号分隔字符串（如 "4,5,2"）或数字数组的两种情况，统一转为 number[]
function normalizeRoleIds(roleIds: any): number[] {
  if (Array.isArray(roleIds)) {
    return roleIds.map((id) => Number(id)).filter((n) => !Number.isNaN(n))
  }
  if (typeof roleIds === 'string' && roleIds.trim() !== '') {
    return roleIds
      .split(',')
      .map((s) => Number(s.trim()))
      .filter((n) => !Number.isNaN(n))
  }
  return []
}

function getStoredUserInfo(): UserInfo | null {
  const stored = localStorage.getItem(USER_INFO_KEY)
  if (!stored) return null
  try {
    const data = JSON.parse(stored) as any
    // 兼容后端旧数据：ID 大写转小写
    if (data.ID !== undefined && data.id === undefined) {
      data.id = data.ID
    }
    data.roleIds = normalizeRoleIds(data.roleIds)
    return data as UserInfo
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
    userInfo.value = { ...info, roleIds: normalizeRoleIds(info.roleIds) }
    localStorage.setItem(USER_INFO_KEY, JSON.stringify(userInfo.value))
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
    localStorage.removeItem('app-security')
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
