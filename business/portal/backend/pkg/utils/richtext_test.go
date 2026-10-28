package utils

import (
	"strings"
	"testing"
)

func TestSanitizeRichTextImageLink(t *testing.T) {
	tests := []struct {
		name     string
		href     string
		preserve bool
	}{
		{name: "https link", href: "https://example.com/article/1", preserve: true},
		{name: "site link", href: "/article/1.html", preserve: true},
		{name: "script link", href: "javascript:alert(1)", preserve: false},
		{name: "data link", href: "data:text/html,unsafe", preserve: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := `<p><img src="https://example.com/image.png" data-href="` + tt.href + `" alt="image"></p>`
			got := SanitizeRichText(input)
			hasLink := strings.Contains(got, `data-href="`+tt.href+`"`)
			if hasLink != tt.preserve {
				t.Fatalf("image link preserved = %t, want %t; output: %s", hasLink, tt.preserve, got)
			}
		})
	}
}
