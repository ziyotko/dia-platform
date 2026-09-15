import { ref } from 'vue'
import { getTenantList } from '@/api/tenant'
import type { Tenant } from '@/api/tenant'

// 租户下拉选项缓存：/tenants 仅超管可访问，因此只在确认是超管时调用。
export const tenantOptions = ref<Tenant[]>([])
export const tenantOptionsLoaded = ref(false)

let loading: Promise<void> | null = null

export function ensureTenants(force = false): Promise<void> {
  if (loading && !force) return loading
  if (tenantOptionsLoaded.value && !force) return Promise.resolve()

  loading = getTenantList({ page: 1, size: 1000 })
    .then((res: any) => {
      tenantOptions.value = res.data?.list || []
      tenantOptionsLoaded.value = true
    })
    .catch(() => {
      tenantOptions.value = []
    })
    .finally(() => {
      loading = null
    })
  return loading
}

export function tenantName(tenantId?: number): string {
  if (!tenantId) return '平台级'
  const hit = tenantOptions.value.find((t) => t.id === tenantId)
  return hit ? hit.name : `租户#${tenantId}`
}
