package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"server/utils"
)

// RateLimitMiddleware 基于 Redis 的固定窗口限流，按真实客户端 IP 计数。
// 用于公开写接口（如站点分析），防止被脚本刷量/灌脏数据。
//
// 窗口 key 以 IP + 当前时间窗口段生成，便于分布式场景下多实例共享计数。
// 返回 limit 个请求/窗口 之外的请求被拒绝。
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	return func(ctx *gin.Context) {
		ip := utils.RealIP(ctx)
		windowSec := int64(window.Seconds())
		now := time.Now().Unix()
		key := fmt.Sprintf("ratelimit:%s:%d", ip, now/windowSec)

		count, err := utils.Redis1.Incr(utils.Ctx, key).Result()
		if err != nil {
			// Redis 不可用时放行，避免限流组件故障导致业务中断（fail-open）
			ctx.Next()
			return
		}
		if count == 1 {
			utils.Redis1.Expire(utils.Ctx, key, window)
		}
		if count > int64(limit) {
			ctx.JSON(http.StatusOK, utils.Error(1, "请求过于频繁，请稍后再试"))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
