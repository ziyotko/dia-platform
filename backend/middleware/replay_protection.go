package middleware

import (
	"bytes"
	"fmt"
	"io"
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
		signature := c.GetHeader("X-Request-Signature")

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

		// 已登录接口必须携带请求签名，未登录接口（如 /login、/captcha）不强制签名
		userID, hasAuth := c.Get("userID")
		if hasAuth {
			if signature == "" {
				c.JSON(http.StatusOK, utils.Error(1, "缺少请求签名"))
				c.Abort()
				return
			}

			signKeyKey := fmt.Sprintf("signkey:%d", userID.(uint))
			signKey, err := utils.Redis.Get(utils.Ctx, signKeyKey).Result()
			if err != nil {
				c.JSON(http.StatusOK, utils.Error(1, "签名密钥无效或已过期"))
				c.Abort()
				return
			}

			bodyBytes, err := c.GetRawData()
			if err != nil {
				c.JSON(http.StatusOK, utils.Error(1, "读取请求体失败"))
				c.Abort()
				return
			}
			// 恢复请求体，供后续中间件和 handler 读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			bodyHash := utils.Sha256(string(bodyBytes))
			expected := utils.SignRequest(signKey, c.Request.Method, c.Request.URL.Path, timestampStr, nonce, bodyHash)
			utils.Logger.Infof("[签名调试] method=%s path=%s timestamp=%s nonce=%s bodyHash=%s signKey=%s expected=%s received=%s",
				c.Request.Method, c.Request.URL.Path, timestampStr, nonce, bodyHash, signKey, expected, signature)
			if !utils.VerifyRequest(signKey, c.Request.Method, c.Request.URL.Path, timestampStr, nonce, bodyHash, signature) {
				c.JSON(http.StatusOK, utils.Error(1, "请求签名无效"))
				c.Abort()
				return
			}
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
