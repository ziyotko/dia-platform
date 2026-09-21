package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"server/models"
	"server/utils"
)

// AuthErrorCode 鉴权失败时返回的业务码。前端 request.ts 会对该码自动登出并跳转登录页。
const AuthErrorCode = 401

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, utils.Error(AuthErrorCode, "未提供token"))
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusOK, utils.Error(AuthErrorCode, "token格式错误"))
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ParseToken(token)
		if err != nil || claims == nil {
			c.JSON(http.StatusOK, utils.Error(AuthErrorCode, "token已过期"))
			c.Abort()
			return
		}

		_, err = utils.Redis.Get(utils.Ctx, "blacklist:"+claims.ID).Result()
		if err == nil {
			c.JSON(http.StatusOK, utils.Error(AuthErrorCode, "token已被拉黑"))
			c.Abort()
			return
		}
		if err != redis.Nil {
			c.JSON(http.StatusOK, utils.Error(1, "服务异常，请稍后重试"))
			c.Abort()
			return
		}

		// 账号存活校验：被禁用/删除的账号立即失效，否则「禁用」要等 JWT 自然过期（默认 24h）才生效，
		// 期间持旧 Token 仍可访问全部 member/admin 接口。用户表极小且按主键查询，代价可忽略。
		var user models.User
		if err := utils.DB.Select("id", "status").First(&user, claims.UserID).Error; err != nil || user.Status != 1 {
			c.JSON(http.StatusOK, utils.Error(AuthErrorCode, "账号不存在或已被禁用"))
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
