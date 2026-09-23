package service

import (
	"strings"
	"time"
	"unicode"
)

// toInt converts a JSON-decoded numeric value into an int (used for the 0/1 status
// columns, where encoding/json gives us float64).
func toInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case uint64:
		return int(n)
	}
	return 0
}

// isDuplicateKeyErr 判断是否为唯一键冲突（MySQL 1062）。
// 项目未开启 GORM 的 TranslateError，只能按错误文本判断。
func isDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "1062") || strings.Contains(msg, "duplicate entry")
}

// snakeKey converts a camelCase / PascalCase key (e.g. "projectBrief") into the
// snake_case column name used by the models ("project_brief"). Keys that are
// already snake_case are returned unchanged.
//
// This is required because GORM's `Updates(map[string]interface{})` resolves a
// key against the schema by its exact database name or Go field name only — a
// lowerCamel key such as "projectBrief" matches neither and would be emitted
// verbatim as a column (`SET projectBrief = ?`), producing a SQL error.
func snakeKey(k string) string {
	r := []rune(k)
	var b strings.Builder
	b.Grow(len(k) + 4)
	for i, c := range r {
		if unicode.IsUpper(c) {
			if i > 0 && (unicode.IsLower(r[i-1]) || unicode.IsDigit(r[i-1]) ||
				(i+1 < len(r) && unicode.IsLower(r[i+1]))) {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(c))
			continue
		}
		b.WriteRune(c)
	}
	return b.String()
}

// toUint64 converts a JSON-decoded numeric value into an unsigned integer.
// encoding/json decodes every number as float64, while internal callers may pass
// int / int64 / uint64. Negative values are clamped to 0.
func toUint64(v interface{}) uint64 {
	switch n := v.(type) {
	case uint64:
		return n
	case uint:
		return uint64(n)
	case int:
		if n > 0 {
			return uint64(n)
		}
	case int64:
		if n > 0 {
			return uint64(n)
		}
	case float64:
		if n > 0 {
			return uint64(n)
		}
	case float32:
		if n > 0 {
			return uint64(n)
		}
	}
	return 0
}

// keysOf returns the keys of a set-style map as a slice (used for IN queries).
func keysOf(set map[uint64]bool) []uint64 {
	out := make([]uint64, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}

// pickUpdates normalizes incoming JSON keys to snake_case and keeps only the
// allowed columns, so callers cannot mass-assign protected fields.
func pickUpdates(updates map[string]interface{}, allowed ...string) map[string]interface{} {
	set := make(map[string]bool, len(allowed))
	for _, a := range allowed {
		set[a] = true
	}
	clean := make(map[string]interface{}, len(updates))
	for k, v := range updates {
		if nk := snakeKey(k); set[nk] {
			clean[nk] = v
		}
	}
	return clean
}

// normalizeTimeFields converts string values of the listed (snake_case) columns
// into time.Time / nil, so they can be written to DATETIME columns safely.
// Accepts RFC3339 (what the Go API returns), "2006-01-02 15:04:05" and
// "2006-01-02"; an empty string clears the column.
func normalizeTimeFields(updates map[string]interface{}, fields ...string) {
	for _, f := range fields {
		v, ok := updates[f]
		if !ok {
			continue
		}
		s, isStr := v.(string)
		if !isStr {
			continue
		}
		s = strings.TrimSpace(s)
		if s == "" {
			updates[f] = nil
			continue
		}
		for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05"} {
			if t, err := time.Parse(layout, s); err == nil {
				updates[f] = t
				break
			}
		}
		if _, ok := updates[f].(time.Time); ok {
			continue
		}
		// 无时区串（前端 el-date-picker 发出的 "2006-01-02 15:04:05"）必须按服务器本地时区解析：
		// time.Parse 会当成 UTC，使申报起止时间整体偏移 8 小时，且批次发布后不可改。
		for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
				updates[f] = t
				break
			}
		}
	}
}

// stringOf reads a string value out of a decoded JSON object.
func stringOf(v interface{}) string {
	s, _ := v.(string)
	return s
}

// timeOf reads a *time.Time out of a normalized JSON object.
func timeOf(v interface{}) *time.Time {
	switch t := v.(type) {
	case time.Time:
		return &t
	case *time.Time:
		return t
	}
	return nil
}
