package service

import (
	"strings"
	"time"
	"unicode"
)

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
		for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.Parse(layout, s); err == nil {
				updates[f] = t
				break
			}
		}
	}
}
