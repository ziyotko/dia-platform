package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/utils"
)

// 管理员角色 ID（与前端 userStore.userInfo.roleIds.includes(1/2) 保持一致）
// 1=超级管理员，2=普通管理员
const (
	AdminRoleID  = 1
	AdminRoleID2 = 2
)

// HasAdminRole 判断逗号分隔的角色 ID 字符串中是否包含管理员角色。
func HasAdminRole(roleIdsStr string) bool {
	for _, part := range strings.Split(roleIdsStr, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && IsAdminRoleID(id) {
			return true
		}
	}
	return false
}

// HasAdminRoleIDs 判断角色 ID 列表是否包含管理员角色（1=超级管理员，2=普通管理员）。
// 供控制器与中间件共用，避免各处重复判断且判定口径不一致。
func HasAdminRoleIDs(roleIds []int) bool {
	for _, id := range roleIds {
		if IsAdminRoleID(id) {
			return true
		}
	}
	return false
}

// IsAdminRoleID 判断单个角色 ID 是否属于管理员角色。
func IsAdminRoleID(id int) bool {
	return id == AdminRoleID || id == AdminRoleID2
}

// IsAdminUser 查询用户是否具备管理员角色。
func IsAdminUser(userID uint) bool {
	var user models.User
	if err := utils.DB.First(&user, userID).Error; err != nil {
		return false
	}
	return HasAdminRole(user.RoleIds)
}

// AdminMiddleware 要求当前登录用户必须是管理员（角色 ID=1 或 2）。
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
