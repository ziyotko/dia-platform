import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { App } from '@/api/app'
import * as appApi from '@/api/app'

export const useAppStore = defineStore('app', () => {
  const collapsed = ref(false)
  const myApps = ref<App[]>([])

  function toggleCollapse() {
    collapsed.value = !collapsed.value
  }

  async function fetchMyApps() {
    const res: any = await appApi.getMyApps()
    myApps.value = res.data || []
    return res.data
  }

  return {
    collapsed,
    myApps,
    toggleCollapse,
    fetchMyApps
  }
})
