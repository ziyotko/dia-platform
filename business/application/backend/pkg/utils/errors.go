package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// LogError 安全写错误日志：Logger 未初始化（脚本/验证程序）时退化为标准输出，避免 nil panic。
func LogError(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Errorf(format, args...)
		return
	}
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

// dbErrorPattern 匹配数据库/驱动/系统层的底层错误文本。
// 这类文本可能含表名、列名、SQL 片段、磁盘路径，一律不得回传前端。
var dbErrorPattern = regexp.MustCompile(`(?i)(error \d{4}|duplicate entry|sql:|sqlstate|unknown column|data too long|incorrect .* value|out of range|foreign key constraint|no such file|permission denied|dial tcp|connection refused|EOF)`)

// SafeMessage 把错误文案转成可以安全返回给前端的文案。
//
// 约定：service 层用 errors.New("中文提示") 表达「业务错误」，这类信息需要原样透传给用户；
// 其余（数据库 / 网络 / 文件系统，通常为英文，或中文前缀里拼了 SQL 文本）一律替换成 fallback，
// 同时把原文写进服务端日志，避免向客户端泄露内部细节。
func SafeMessage(message string, fallback string) string {
	msg := strings.TrimSpace(message)
	if msg == "" {
		return fallback
	}
	if dbErrorPattern.MatchString(msg) || !containsHan(msg) {
		LogError("internal error not exposed to client: %s", msg)
		return fallback
	}
	return msg
}

// SafeErrMessage 同 SafeMessage，接受 error。
func SafeErrMessage(err error, fallback string) string {
	if err == nil {
		return fallback
	}
	return SafeMessage(err.Error(), fallback)
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
