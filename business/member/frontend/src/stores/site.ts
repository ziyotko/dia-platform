import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'

// Centralized site config loaded from GET /member/api/site-info
// (backed by the member_system_configs table). Used for global page titles.
export const useSiteStore = defineStore('site', () => {
  const siteInfo = ref<Record<string, string>>({})
  const loaded = ref(false)

  const site_name = computed(() => siteInfo.value.site_name || '会员系统')

  async function load() {
    if (loaded.value) return
    try {
      const res = await authApi.getSiteInfo()
      siteInfo.value = res.data || {}
      loaded.value = true
    } catch {
      // Keep defaults on network failure
    }
  }

  return { siteInfo, loaded, site_name, load }
})
