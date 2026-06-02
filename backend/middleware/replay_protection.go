package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"server/utils"
)

func ReplayProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		timestampStr := c.GetHeader("X-Request-Timestamp")
		nonce := c.GetHeader("X-Request-Nonce")

		if timestampStr == "" || nonce == "" {
			c.JSON(http.StatusOK, utils.Error(1, "缺少防重放攻击请求头"))
			c.Abort()
			return
		}

		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, utils.Error(1, "无效的时间戳"))
			c.Abort()
			return
		}

		now := time.Now().UnixMilli()
		const maxAge = 5 * 60 * 1000
		if now-timestamp > maxAge || timestamp-now > maxAge {
			c.JSON(http.StatusOK, utils.Error(1, "请求已过期，请重新发送"))
			c.Abort()
			return
		}

		nonceKey := fmt.Sprintf("replay:nonce:%s", nonce)
		ok, err := utils.Redis1.SetNX(utils.Ctx, nonceKey, "1", 5*time.Minute).Result()
		if err != nil {
			c.JSON(http.StatusOK, utils.Error(1, "请求验证失败"))
			c.Abort()
			return
		}
		if !ok {
			c.JSON(http.StatusOK, utils.Error(1, "检测到重放攻击，请求已被拒绝"))
			c.Abort()
			return
		}

		c.Next()
	}
}
