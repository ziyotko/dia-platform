package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/utils"
)

// 管理员角色 ID（与前端 userStore.userInfo.roleIds.includes(1) 保持一致）
const AdminRoleID = 1

// HasAdminRole 判断逗号分隔的角色 ID 字符串中是否包含管理员角色。
func HasAdminRole(roleIdsStr string) bool {
	for _, part := range strings.Split(roleIdsStr, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && id == AdminRoleID {
			return true
		}
	}
	return false
}

// IsAdminUser 查询用户是否具备管理员角色。
func IsAdminUser(userID uint) bool {
	var user models.User
	if err := utils.DB.First(&user, userID).Error; err != nil {
		return false
	}
	return HasAdminRole(user.RoleIds)
}

// AdminMiddleware 要求当前登录用户必须是管理员（角色 ID=1）。
// 仅校验角色，不校验具体权限点，用于保护后台管理接口。
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		if !IsAdminUser(userID) {
			c.JSON(200, utils.Error(1, "无权限执行该操作"))
			c.Abort()
			return
		}
		c.Next()
	}
}
