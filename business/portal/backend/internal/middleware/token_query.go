package middleware

import (
	"github.com/gin-gonic/gin"
)

// TokenFromQueryMiddleware 允许用查询参数 `?token=<jwt>` 代替 `Authorization: Bearer <jwt>`。
//
// 用途：`<img src>` / `<video src>` / `<a href>` 这类由浏览器直接发起的请求**无法设置请求头**，
// 而会员专区的文件（封面图/附件/报刊文件/视频）必须登录才能下载。
//
// 使用约束（重要）：
//   - 必须挂在 AuthMiddleware **之前**（同一路由组内，组内中间件按注册顺序执行）；
//   - **只挂需要它的路由**（当前仅 GET /member-files/:name），不要全局挂载——
//     否则所有接口都接受 URL 里的 token，而 token 会进入访问日志/Referer，暴露面变大；
//   - 仅当请求未带 Authorization 头时才回填，标准头鉴权优先级更高。
func TokenFromQueryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			if token := c.Query("token"); token != "" {
				c.Request.Header.Set("Authorization", "Bearer "+token)
			}
		}
		c.Next()
	}
}
