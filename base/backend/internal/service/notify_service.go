package service

import "base/pkg/notifier"

// ReloadNotifiers 按当前系统设置重新注册所有消息发送器。
//
// 取值来源：
//   - category = email ：SMTP 配置（host/port/username/password/from/ssl）
//   - category = notify：企业微信 webhook、短信网关
//
// 启动时与「系统设置」保存成功后都会调用，使渠道配置立即生效（未配置的渠道会被注销，
// 前端渠道下拉也就不会再出现该选项）。
func ReloadNotifiers() {
	email, _ := (SettingsService{}).GetByCategory("email")
	notifier.LoadEmailSender(func(key string) string { return email[key] })

	notify, _ := (SettingsService{}).GetByCategory("notify")
	get := func(key string) string { return notify[key] }
	notifier.LoadWeChatSender(get)
	notifier.LoadSmsSender(get)
}
