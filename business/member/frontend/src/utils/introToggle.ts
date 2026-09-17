import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import type { ComponentPublicInstance } from 'vue'

/**
 * 长文本「展开 / 收起」的通用逻辑。
 *
 * 折叠样式由调用方的 CSS 提供（通常是 `display:-webkit-box; -webkit-line-clamp:N; overflow:hidden;`，
 * 展开态用 `.expanded` 覆盖）。只有真正被截断的文本才会显示切换按钮，短文本不显示。
 *
 * 用法：
 * ```ts
 * const { setIntroRef, introExpanded, introOverflowed, toggleIntro, measureIntro } = useIntroToggle()
 * ```
 * ```html
 * <div :ref="setIntroRef(key)" :class="{ expanded: introExpanded(key) }">{{ text }}</div>
 * <button v-if="introOverflowed(key)" @click="toggleIntro(key)">
 *   {{ introExpanded(key) ? '收起' : '展开' }}
 * </button>
 * ```
 * 文本内容变化后调用 `measureIntro()`（内部已 await nextTick）；窗口尺寸变化会自行重新测量。
 *
 * 实现说明：折叠态下 `scrollHeight > clientHeight` 即代表内容被截断（Chromium 实测可靠）。
 * 已展开的元素此时两者相等，所以测量时跳过它们并保留原有结论 ——
 * 否则点开「展开」后会被误判为「未溢出」，按钮随之消失。
 * 因此 key 必须在整个生命周期内保持稳定（如 `'org-' + item.id`）。
 */
export function useIntroToggle() {
  /** 当前展开的 key */
  const expandedKeys = ref<Set<string>>(new Set())
  /** 折叠态下确实溢出的 key（决定是否渲染切换按钮） */
  const overflowKeys = ref<Set<string>>(new Set())
  /** key -> 元素，由 :ref 回调维护 */
  const elMap = new Map<string, HTMLElement>()

  /**
   * 生成 :ref 回调。函数 ref 在元素卸载或 ref 变化时会以 null 回调，
   * 此时删除映射，避免残留失效元素。
   */
  function setIntroRef(key: string) {
    return (el: Element | ComponentPublicInstance | null) => {
      if (el instanceof HTMLElement) elMap.set(key, el)
      else elMap.delete(key)
    }
  }

  function introExpanded(key: string) {
    return expandedKeys.value.has(key)
  }

  function introOverflowed(key: string) {
    return overflowKeys.value.has(key)
  }

  function toggleIntro(key: string) {
    const next = new Set(expandedKeys.value)
    if (next.has(key)) next.delete(key)
    else next.add(key)
    expandedKeys.value = next
  }

  /** 重新测量所有已注册元素是否溢出（已展开的跳过，保留结论） */
  async function measureIntro() {
    await nextTick()
    const next = new Set(overflowKeys.value)
    elMap.forEach((el, key) => {
      if (expandedKeys.value.has(key)) return
      if (el.scrollHeight > el.clientHeight + 1) next.add(key)
      else next.delete(key)
    })
    overflowKeys.value = next
  }

  function handleResize() {
    void measureIntro()
  }

  onMounted(() => window.addEventListener('resize', handleResize))
  onBeforeUnmount(() => window.removeEventListener('resize', handleResize))

  return { setIntroRef, introExpanded, introOverflowed, toggleIntro, measureIntro }
}
