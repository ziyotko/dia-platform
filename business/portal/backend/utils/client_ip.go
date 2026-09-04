package utils

import "github.com/gin-gonic/gin"

// RealIP 获取真实客户端 IP。
//
// 采用严谨策略：直接委托给 gin 的 ClientIP()。
// gin 仅当请求来自 trusted_proxies 中配置的可信代理时才信任
// X-Forwarded-For / X-Real-IP，并会从右向左跳过可信代理、返回第一个非可信 IP；
// 否则忽略转发头、返回实际对端 IP，从而避免伪造头绕过按 IP 的去重/限流。
//
// 因此部署在反向代理后时，请务必在配置 server.trusted_proxies 中填入代理真实地址。
func RealIP(c *gin.Context) string {
	return c.ClientIP()
}
