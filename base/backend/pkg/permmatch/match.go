// Package permmatch 接口权限的路径匹配工具。
//
// 底座有两处需要「按 method + path 判断权限」：base 自身的接口（middleware.PermissionAuth）
// 与子应用代理入口（子应用上报的权限点，app_code = 子应用编码）。
// 匹配规则必须一致，因此统一放在这里。
package permmatch

import (
	"strings"

	"base/config"
)

// APIPrefix 返回底座 HTTP 路由前缀。唯一来源是 config.yaml 的 server.api_prefix
// （解析与兜底见 config.applyDefaults）：路由注册、权限白名单匹配、审计脱敏判定、文件 URL、
// 路径匹配都引用它，因此改前缀只需改配置，不会漏改某处导致 404 或权限误判。
func APIPrefix() string {
	if config.Cfg != nil && config.Cfg.Server.APIPrefix != "" {
		return strings.TrimSuffix(strings.TrimSpace(config.Cfg.Server.APIPrefix), "/")
	}
	return config.DefaultAPIPrefix
}

// Normalize 规整路径：去除首尾空格与结尾斜杠，保证以 / 开头。
func Normalize(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if len(p) > 1 {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// Candidates 返回请求路径的候选集合：完整路径 +（若有）去掉 APIPrefix 前缀后的相对路径，
// 这样权限点里写完整路径或相对路径都能命中。
func Candidates(requestPath string) []string {
	normalized := Normalize(requestPath)
	paths := []string{normalized}
	prefix := APIPrefix()
	if prefix != "" && strings.HasPrefix(normalized, prefix) {
		relative := strings.TrimPrefix(normalized, prefix)
		if relative == "" {
			relative = "/"
		}
		paths = append(paths, relative)
	}
	return paths
}

// Match 判断权限路径是否匹配请求路径，支持 :param 与 * 通配段。
func Match(permPath, requestPath string) bool {
	permParts := strings.Split(Normalize(permPath), "/")
	reqParts := strings.Split(Normalize(requestPath), "/")
	if len(permParts) != len(reqParts) {
		return false
	}
	for i := range permParts {
		if strings.HasPrefix(permParts[i], ":") || strings.HasPrefix(permParts[i], "*") {
			continue
		}
		if permParts[i] != reqParts[i] {
			return false
		}
	}
	return true
}
