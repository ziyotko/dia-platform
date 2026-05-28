package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"server/services"
	"server/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供token"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token格式错误"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": services.ErrTokenExpired.Error()})
			c.Abort()
			return
		}

		_, err = utils.Redis.Get(utils.Ctx, "blacklist:"+claims.ID).Result()
		if err == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": services.ErrTokenBlacklisted.Error()})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
