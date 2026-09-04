package utils

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// RealIP 获取真实客户端 IP。
//
// 部署在反向代理后时，ctx.ClientIP() 可能返回代理自身的地址，
// 因此优先从可信代理写入的 X-Real-IP / X-Forwarded-For 中解析，
// 并对解析结果做合法性校验，避免攻击者伪造非法 IP 绕过按 IP 的去重/限流。
//
// 优先级：X-Real-IP（代理以 $remote_addr 写入，不可伪造） > X-Forwarded-For > ClientIP。
func RealIP(c *gin.Context) string {
	if ip := parseIPHeader(c.GetHeader("X-Real-IP")); ip != "" {
		return ip
	}
	if ip := parseIPHeader(c.GetHeader("X-Forwarded-For")); ip != "" {
		return ip
	}
	return c.ClientIP()
}

// parseIPHeader 从可能逗号分隔的头中提取第一个合法的 IP 地址。
// 同时兼容 "IP" 与 "IP:port" 两种形式。
func parseIPHeader(header string) string {
	if header == "" {
		return ""
	}
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if ip, _, err := net.SplitHostPort(part); err == nil {
			if parsed := net.ParseIP(ip); parsed != nil {
				return parsed.String()
			}
		}
		if parsed := net.ParseIP(part); parsed != nil {
			return parsed.String()
		}
	}
	return ""
}
