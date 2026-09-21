package services

import "strings"

// SiteBaseURL 返回「站点地址」基址（系统设置-基础设置 siteUrl）。
// 仅接受 http(s) 开头的地址，并去掉末尾多余的「/」；未配置或非 http(s) 时返回空串。
func SiteBaseURL() string {
	settings, err := (&SettingsService{}).GetSettings()
	if err != nil || settings == nil {
		return ""
	}
	base := strings.TrimRight(strings.TrimSpace(settings.SiteUrl), "/")
	if !isAbsoluteHTTPURL(base) {
		return ""
	}
	return base
}

// BuildPageAccessURL 把站内访问路径拼成可直接访问的地址（静态化列表下发的 url 字段）。
// 规则：
//  1. accessPath 本身已是 http(s) 完整地址 → 原样返回；
//  2. baseURL（站点地址）已配置 → baseURL + 以 / 开头的 accessPath；
//  3. baseURL 为空 → 返回站内相对路径（以 / 开头），由浏览器按当前站点根解析
//     （静态文件与后台通常同域部署，此时相对路径即可访问）。
//
// 注意：「静态化输出路径」（setting.static_path，如 D:/static）是服务器上的文件输出目录，
// 不是可访问地址，故不参与拼接。
func BuildPageAccessURL(baseURL, accessPath string) string {
	path := strings.TrimSpace(accessPath)
	if path == "" {
		return ""
	}
	if isAbsoluteHTTPURL(path) {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if !isAbsoluteHTTPURL(base) {
		return path
	}
	return base + path
}

// isAbsoluteHTTPURL 判断是否为 http(s):// 开头的绝对地址
func isAbsoluteHTTPURL(s string) bool {
	lower := strings.ToLower(s)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}
