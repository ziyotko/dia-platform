package utils

import (
	"fmt"
	"strings"
	"unicode"
)

// LogError 安全写错误日志：Logger 未初始化（测试/脚本环境）时退化为标准输出，避免 nil panic。
func LogError(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Errorf(format, args...)
		return
	}
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

// SafeErrMessage 把 service 层错误转成可以安全返回给前端的文案。
//
// 约定：service 层用 errors.New("中文提示") 表达「业务错误」，这类信息需要原样透传给用户；
// 其余多为数据库 / 网络 / 文件系统等底层错误（英文，可能含表名、SQL 片段、磁盘路径），
// 一律替换成 fallback，同时把原文写进服务端日志，避免向客户端泄露内部细节。
func SafeErrMessage(err error, fallback string) string {
	if err == nil {
		return fallback
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return fallback
	}
	if containsHan(msg) {
		return msg
	}
	LogError("internal error not exposed to client: %v", err)
	return fallback
}

// containsHan 判断字符串是否含汉字（含汉字即视为面向用户的业务提示）。
func containsHan(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
