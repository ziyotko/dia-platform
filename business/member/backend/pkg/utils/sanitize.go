package utils

import (
	"encoding/base64"
	"net/url"
	"strings"
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

// dataURIBitmapPrefixes 允许的内联图片 data URI 前缀（仅位图格式）。
//
// 注意：bluemonday v1.0.27 自带的 AllowDataURIImages() 会放行 image/svg+xml，
// 而 SVG 可以内嵌脚本（直接打开该 URL 时按同源文档执行），因此这里不用它，
// 改为自定义策略，只放行 gif/jpeg/png/webp。
var dataURIBitmapPrefixes = []string{
	"image/gif;base64,",
	"image/jpeg;base64,",
	"image/png;base64,",
	"image/webp;base64,",
}

var (
	richTextPolicyOnce sync.Once
	richTextPolicy     *bluemonday.Policy
)

// RichTextPolicy 返回富文本（会员文章 / 公告 / 协会章程）统一的消毒策略。
//
// 设计取舍：
//   - 保留 wangEditor 产出的常见排版标签（标题/段落/列表/表格/引用/代码/图片等），
//     否则清洗后正文会明显掉格式；
//   - 仅放行与排版相关的内联样式属性，缩略值交由 bluemonday 内置的 CSS 解析器清洗，
//     因此 style 里不可能出现 expression()/url(javascript:) 之类的可执行内容；
//   - 允许内联 Base64 位图图片（gif/jpeg/png/webp，必须是合法 base64，显式排除可携带
//     脚本的 svg+xml），与编辑器直接粘贴图片的既有行为保持一致；
//   - script/iframe/object/embed/svg/form 等可执行或可交互标签、on* 事件属性、
//     javascript:/vbscript: 等危险协议一律剔除（UGCPolicy 默认行为）。
func RichTextPolicy() *bluemonday.Policy {
	richTextPolicyOnce.Do(func() {
		p := bluemonday.UGCPolicy()
		// 内联 Base64 图片：仅位图格式（显式排除 svg+xml）
		p.AllowURLSchemeWithCustomPolicy("data", allowBitmapDataURI)
		p.AllowStyles(
			"color", "background-color",
			"text-align", "text-indent", "text-decoration",
			"font-size", "font-family", "font-weight", "font-style",
			"line-height", "letter-spacing",
			"margin", "margin-top", "margin-right", "margin-bottom", "margin-left",
			"padding", "padding-top", "padding-right", "padding-bottom", "padding-left",
			"border", "border-width", "border-style", "border-color", "border-collapse",
			"width", "height", "max-width", "min-width",
			"vertical-align", "white-space", "word-break", "list-style-type",
		).Globally()
		// 外链统一新窗口打开（bluemonday 会同时补 noopener/noreferrer）
		p.AddTargetBlankToFullyQualifiedLinks(true)
		richTextPolicy = p
	})
	return richTextPolicy
}

// allowBitmapDataURI 校验 data: URI 是否为受支持的位图图片：
// 必须是 gif/jpeg/png/webp + base64、不含 query/fragment、且 base64 可解析。
func allowBitmapDataURI(u *url.URL) bool {
	if u == nil || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	opaque := u.Opaque
	prefix := ""
	for _, p := range dataURIBitmapPrefixes {
		if strings.HasPrefix(opaque, p) {
			prefix = p
			break
		}
	}
	if prefix == "" {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(strings.TrimSpace(opaque[len(prefix):]))
	return err == nil
}

// SanitizeRichText 清洗富文本 HTML，返回可安全用于 v-html 渲染的内容。
// 空串直接返回，避免无谓开销。
func SanitizeRichText(html string) string {
	if strings.TrimSpace(html) == "" {
		return ""
	}
	return RichTextPolicy().Sanitize(html)
}

// StripAllTags 去除全部 HTML 标签，用于只允许纯文本的字段（如摘要）。
func StripAllTags(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return bluemonday.StrictPolicy().Sanitize(s)
}
