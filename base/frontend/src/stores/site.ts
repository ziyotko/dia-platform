import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getSiteInfo } from '@/api/auth'

const DEFAULT_NAME = 'Base 底座平台'

/**
 * 站点信息（系统设置 → 基础配置）：平台名称 / Logo / 版权 + 登录验证码开关。
 * 数据来自公开接口 /site-info，登录页与已登录页面（浏览器标题）都会消费，
 * 否则「平台名称」这类设置只是写进库、看不到任何效果。
 */
export const useSiteStore = defineStore('site', () => {
  const platformName = ref(DEFAULT_NAME)
  const logo = ref('')
  const copyright = ref('')
  const captchaEnabled = ref(true)
  const loaded = ref(false)

  async function fetchSiteInfo() {
    try {
      const res: any = await getSiteInfo()
      const data = res?.data || {}
      if (data.platformName) platformName.value = data.platformName
      logo.value = data.logo || ''
      copyright.value = data.copyright || ''
      if (data.captchaEnabled !== undefined) captchaEnabled.value = data.captchaEnabled !== false
      document.title = platformName.value
    } catch (error) {
      // 取不到时保留默认值，不影响登录/使用
      console.error('[base] 获取站点信息失败', error)
    } finally {
      loaded.value = true
    }
  }

  return { platformName, logo, copyright, captchaEnabled, loaded, fetchSiteInfo }
})
