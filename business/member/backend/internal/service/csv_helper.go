package service

import (
	"strings"

	"member/internal/models"
)

// csvField 按 RFC4180 转义单个 CSV 字段：含逗号、引号、换行时加引号并把内部引号翻倍。
// 会籍记录的公司名称/变更原因等字段可能带逗号，不转义会导致列错位。
func csvField(value string) string {
	if !strings.ContainsAny(value, ",\"\r\n") {
		return value
	}
	return "\"" + strings.ReplaceAll(value, "\"", "\"\"") + "\""
}

// newCSVBuilder 创建带 UTF-8 BOM 的 CSV 构建器：Excel 直接打开中文不乱码。
func newCSVBuilder() *strings.Builder {
	sb := &strings.Builder{}
	sb.WriteString("\uFEFF")
	return sb
}

// csvLine 把若干字段拼成一行 CSV（各字段自动转义）。
func csvLine(fields ...string) string {
	escaped := make([]string, 0, len(fields))
	for _, f := range fields {
		escaped = append(escaped, csvField(f))
	}
	return strings.Join(escaped, ",") + "\n"
}

// memberTypeLabel 会员类型的中文标签（导出用）。
func memberTypeLabel(t string) string {
	switch t {
	case models.MemberTypeUnit:
		return "单位会员"
	case models.MemberTypePersonal:
		return "个人会员"
	default:
		return t
	}
}
