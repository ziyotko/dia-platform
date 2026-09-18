import DOMPurify from 'dompurify'

/**
 * 富文本渲染前的统一消毒（所有 v-html 场景：会员文章、公告、协会章程等）。
 *
 * 后端在写入时已用 bluemonday 消毒，这里作为兜底：
 *   - 覆盖后端消毒策略上线之前就已入库的历史数据；
 *   - 防御后端策略被绕过或后续新增写入点遗漏消毒。
 *
 * 允许的标签/属性与后端策略保持一致（HTML profile + 内联 style + 位图 data URI），
 * 因此不会影响编辑器产出的排版效果。
 */

// 允许内联的图片 data URI 前缀：仅位图格式。
// 与后端 utils.SanitizeRichText 保持一致 —— 显式排除 image/svg+xml（SVG 可内嵌脚本）。
const BITMAP_DATA_URI = /^data:image\/(gif|jpeg|png|webp);base64,/i

// DOMPurify 默认会放行 <img src="data:image/svg+xml;...">（它认为 <img> 里的 SVG 不执行脚本），
// 这里按后端策略收紧，移除所有非位图的 data URI 图片。
DOMPurify.addHook('afterSanitizeAttributes', (node) => {
  if (node instanceof Element && node.tagName === 'IMG') {
    const src = node.getAttribute('src') || ''
    if (/^data:/i.test(src) && !BITMAP_DATA_URI.test(src)) {
      node.removeAttribute('src')
    }
  }
})

export function sanitizeHtml(html?: string | null): string {
  if (!html) return ''
  return DOMPurify.sanitize(String(html), {
    USE_PROFILES: { html: true },
    // 编辑器产出的 <a target="_blank"> 需要保留
    ADD_ATTR: ['target']
  })
}
