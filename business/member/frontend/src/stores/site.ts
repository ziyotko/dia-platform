import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'

// Centralized site config loaded from GET /business_member/api/site-info
// (backed by the member_system_configs table). Used for global page titles.
export const useSiteStore = defineStore('site', () => {
  const siteInfo = ref<Record<string, string>>({})
  const loaded = ref(false)

  const site_name = computed(() => siteInfo.value.site_name || '会员系统')
  const site_description = computed(() => siteInfo.value.site_description || '')
  const copyright_name = computed(() => siteInfo.value.copyright_name || '')
  const icp_no = computed(() => siteInfo.value.icp_no || '')
  const beian_no = computed(() => siteInfo.value.beian_no || '')

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

  return { siteInfo, loaded, site_name, site_description, copyright_name, icp_no, beian_no, load }
})
