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
