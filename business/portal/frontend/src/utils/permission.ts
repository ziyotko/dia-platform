import router from '@/router'

/**
 * 权限判定工具（全前端唯一口径，须与后端 models/role.go 保持一致）
 *
 * 管理员角色 ID：1=超级管理员，2=普通管理员。
 * 后端 `AdminMiddleware` 与各控制器的管理员校验（`models.HasAdminRoleIDs`）都是「1 或 2」，
 * 前端按钮显示必须使用同一判定，避免出现「按钮可见但接口拒绝」或反之。
 */
export const ADMIN_ROLE_IDS: number[] = [1, 2]

/** 角色 ID 列表是否包含管理员角色（1 或 2）。 */
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
