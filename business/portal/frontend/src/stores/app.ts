import { defineStore } from 'pinia'
import { ref } from 'vue'

// 品牌主题色（固定，不再支持后台配置）
// 需要导出：Canvas 场景（ECharts）无法使用 CSS 变量 var(--el-color-primary)，只能取 JS 常量
export const BRAND_THEME_COLOR = '#002fa7'

// 按 Element Plus 算法混合两种颜色（weight 为 color2 占比 0~1）
const mixColor = (color1: string, color2: string, weight: number): string => {
  const c1 = parseInt(color1.slice(1), 16)
  const c2 = parseInt(color2.slice(1), 16)
  const mix = (v1: number, v2: number) => Math.round(v1 * (1 - weight) + v2 * weight)
  const r = mix((c1 >> 16) & 255, (c2 >> 16) & 255)
  const g = mix((c1 >> 8) & 255, (c2 >> 8) & 255)
  const b = mix(c1 & 255, c2 & 255)
  return '#' + [r, g, b].map(v => v.toString(16).padStart(2, '0')).join('')
}

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

const storedCollapsed = localStorage.getItem('sidebar-collapsed')

export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(storedCollapsed === 'true')
  const themeColor = ref(BRAND_THEME_COLOR)
  const minPasswordLength = ref(parsedSecurity?.minPasswordLength ?? 8)

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
    localStorage.setItem('sidebar-collapsed', String(sidebarCollapsed.value))
  }

  const refreshKey = ref(0)

  function triggerRefresh() {
    refreshKey.value++
  }

  function applyTheme() {
    const root = document.documentElement
    const primary = themeColor.value
    root.style.setProperty('--el-color-primary', primary)
    root.style.setProperty('--primary-color', primary)
    // 同步 Element Plus 全套色阶，保证按钮/图标/聚焦态/浅色背景统一跟随主题色
    root.style.setProperty('--el-color-primary-light-3', mixColor(primary, '#ffffff', 0.3))
    root.style.setProperty('--el-color-primary-light-5', mixColor(primary, '#ffffff', 0.5))
    root.style.setProperty('--el-color-primary-light-7', mixColor(primary, '#ffffff', 0.7))
    root.style.setProperty('--el-color-primary-light-8', mixColor(primary, '#ffffff', 0.8))
    root.style.setProperty('--el-color-primary-light-9', mixColor(primary, '#ffffff', 0.9))
    root.style.setProperty('--el-color-primary-dark-2', mixColor(primary, '#000000', 0.2))
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

  applyTheme()

  return {
    sidebarCollapsed,
    themeColor,
    minPasswordLength,
    refreshKey,
    toggleSidebar,
    triggerRefresh,
    applyTheme,
    setSecuritySettings
  }
})