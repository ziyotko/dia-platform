package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/services"
	"server/utils"
)

// 管理员判定口径统一在 models 包（models.IsAdminRoleID / HasAdminRoleIDs / HasAdminRoleStr）：
// 仅 1=管理员（2026-09-21 由「超级管理员」更名；原 2=普通管理员 已下线移除）；中间件、控制器与前端 hasAdminRole 均使用这一套判定。

// IsAdminUser 查询用户是否具备管理员角色。
// 角色状态即时生效：管理员角色被禁用后不再具备管理权限（与 MenuService.GetUserMenus 同一口径）。
// 注：内置角色（含管理员）在 RoleService.UpdateRole 中禁止修改，正常不会出现该状态，此处属防御性判断。
func IsAdminUser(userID uint) bool {
	// 与菜单前缀校验共用同一份短 TTL 缓存（services.GetUserAccess），避免每个请求重复查库
	access, err := services.GetUserAccess(userID)
	if err != nil {
		return false
	}
	return access.IsAdmin
}

// AdminMiddleware 要求当前登录用户必须是管理员（角色 ID=1）。
// 仅校验角色，不校验具体权限点，用于保护后台管理接口。
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		if !IsAdminUser(userID) {
			c.JSON(http.StatusOK, utils.Error(1, "无权限执行该操作"))
			c.Abort()
			return
		}
		c.Next()
	}
}
