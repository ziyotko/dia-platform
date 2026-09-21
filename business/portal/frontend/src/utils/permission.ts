import router from '@/router'

/**
 * 权限判定工具（全前端唯一口径，须与后端 models/role.go 保持一致）
 *
 * 管理员角色 ID：仅 1=管理员（code `super_admin`，2026-09-21 由「超级管理员」更名；原 2=普通管理员 角色已下线移除）。
 * 后端 `AdminMiddleware` 与各控制器的管理员校验（`models.HasAdminRoleIDs`）同样只认角色 1，
 * 前端按钮显示必须使用同一判定，避免出现「按钮可见但接口拒绝」或反之。
 */
export const ADMIN_ROLE_IDS: number[] = [1]

/** 角色 ID 列表是否包含管理员角色（仅 1）。 */
export function hasAdminRole(roleIds?: number[] | null): boolean {
  if (!Array.isArray(roleIds) || roleIds.length === 0) return false
  return roleIds.some((id) => ADMIN_ROLE_IDS.includes(Number(id)))
}

/**
 * 当前用户是否可访问指定路径。
 * 后台的动态路由由已授权菜单生成，未授权的菜单不会注册对应路由，
 * 因此跳转前用它做判断可以避免跳到 404 页面（"菜单可见 = 路由可用"）。
 */
export function canAccessPath(path: string): boolean {
  if (!path) return false
  return router.resolve(path).matched.length > 0
}

/** 采集菜单树中的全部路径（与仪表盘快捷入口过滤同口径）。 */
export function collectMenuPaths(menus?: any[] | null): string[] {
  const paths: string[] = []
  const walk = (items: any[]) => {
    items.forEach((menu) => {
      if (menu.path) paths.push(menu.path)
      if (menu.children && menu.children.length > 0) walk(menu.children)
    })
  }
  walk(menus || [])
  return paths
}

/**
 * 首页路径：优先「管理首页」(/dashboard)；未授权该菜单时退回第一个已授权菜单路径。
 * 动态路由由已授权菜单生成，硬编码 /dashboard 会让无该菜单的用户（例如只授「待审核」的审核角色）跳到 404。
 */
export function resolveHomePath(menus?: any[] | null): string {
  const paths = collectMenuPaths(menus)
  return paths.find((path) => path === '/dashboard') || paths[0] || '/dashboard'
}
