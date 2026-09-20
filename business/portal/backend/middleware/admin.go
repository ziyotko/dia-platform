package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/utils"
)

// 管理员判定口径统一在 models 包（models.IsAdminRoleID / HasAdminRoleIDs / HasAdminRoleStr）：
// 仅 1=管理员（2026-09-21 由「超级管理员」更名；原 2=普通管理员 已下线移除）；中间件、控制器与前端 hasAdminRole 均使用这一套判定。

// IsAdminUser 查询用户是否具备管理员角色。
func IsAdminUser(userID uint) bool {
	var user models.User
	if err := utils.DB.First(&user, userID).Error; err != nil {
		return false
	}
	return models.HasAdminRoleStr(user.RoleIds)
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
