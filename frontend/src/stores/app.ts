import { defineStore } from 'pinia'
import { ref } from 'vue'

export type SidebarStyle = 'light' | 'dark'

const storedTheme = localStorage.getItem('app-theme')
const parsedTheme = storedTheme
  ? (() => {
      try {
        return JSON.parse(storedTheme)
      } catch {
        return null
      }
    })()
  : null

const storedSecurity = localStorage.getItem('app-security')
const parsedSecurity = storedSecurity
  ? (() => {
      try {
        return JSON.parse(storedSecurity)
      } catch {
        return null
      }
    })()
  : null

export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(false)
  const themeColor = ref(parsedTheme?.themeColor || '#409eff')
  const sidebarStyle = ref<SidebarStyle>(parsedTheme?.sidebarStyle || 'light')
  const tagsView = ref(parsedTheme?.tagsView ?? true)
  const breadcrumb = ref(parsedTheme?.breadcrumb ?? true)
  const minPasswordLength = ref(parsedSecurity?.minPasswordLength ?? 8)

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  const refreshKey = ref(0)

  function triggerRefresh() {
    refreshKey.value++
  }

  function applyTheme() {
    const root = document.documentElement
    root.style.setProperty('--el-color-primary', themeColor.value)
    root.style.setProperty('--primary-color', themeColor.value)
  }

  function setThemeSettings(settings: {
    themeColor?: string
    sidebarStyle?: SidebarStyle
    tagsView?: boolean
    breadcrumb?: boolean
  }) {
    if (settings.themeColor !== undefined) themeColor.value = settings.themeColor
    if (settings.sidebarStyle !== undefined) sidebarStyle.value = settings.sidebarStyle
    if (settings.tagsView !== undefined) tagsView.value = settings.tagsView
    if (settings.breadcrumb !== undefined) breadcrumb.value = settings.breadcrumb

    localStorage.setItem(
      'app-theme',
      JSON.stringify({
        themeColor: themeColor.value,
        sidebarStyle: sidebarStyle.value,
        tagsView: tagsView.value,
        breadcrumb: breadcrumb.value
      })
    )

    applyTheme()
  }

  function setSecuritySettings(settings: {
    minPasswordLength?: number
  }) {
    if (settings.minPasswordLength !== undefined) minPasswordLength.value = settings.minPasswordLength

    localStorage.setItem(
      'app-security',
      JSON.stringify({
        minPasswordLength: minPasswordLength.value
      })
    )
  }

  return {
    sidebarCollapsed,
    themeColor,
    sidebarStyle,
    tagsView,
    breadcrumb,
    minPasswordLength,
    refreshKey,
    toggleSidebar,
    triggerRefresh,
    applyTheme,
    setThemeSettings,
    setSecuritySettings
  }
})