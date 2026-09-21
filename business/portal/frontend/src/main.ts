import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn.mjs'
import 'element-plus/dist/index.css'
import '@/styles/global.scss'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import router from './router/index.ts'
import { useAppStore } from './stores/app.ts'

// 仅对「会阻塞滚动的」事件（wheel / touchstart / touchmove）默认补上 passive: true，
// 用于消除 Element Plus 等第三方组件在 devtools 中产生的非 passive 监听器告警。
// 与早期实现的两点差异（重要）：
//   1. 只处理上述三个事件，不再改写 click/keydown 等其它事件类型（否则可能出现
//      “passive 监听器里调 preventDefault 无效”这类难查的交互异常）；
//   2. 调用方显式传入的 options（含 passive: false）一律尊重，只在其缺省时补默认值。
const SCROLL_BLOCKING_EVENTS = ['wheel', 'touchstart', 'touchmove']
const originalAddEventListener = EventTarget.prototype.addEventListener
EventTarget.prototype.addEventListener = function (
  type: string,
  listener: EventListenerOrEventListenerObject | null,
  options?: boolean | AddEventListenerOptions
) {
  if (SCROLL_BLOCKING_EVENTS.includes(type)) {
    if (typeof options === 'boolean') {
      options = { capture: options, passive: true }
    } else if (options === undefined) {
      options = { passive: true }
    } else if (options.passive === undefined) {
      options = { ...options, passive: true }
    }
  }
  return originalAddEventListener.call(this, type, listener, options)
}

const app = createApp(App)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })

const appStore = useAppStore()
appStore.applyTheme()

app.mount('#app')                                                                                                                                                                                                                                                          