import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn.mjs'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import router from './router/index.ts'
import { useAppStore } from './stores/app.ts'

// 消除 Element Plus 内部组件的非 passive 事件监听器警告
const originalAddEventListener = EventTarget.prototype.addEventListener
EventTarget.prototype.addEventListener = function (
  type: string,
  listener: EventListenerOrEventListenerObject | null,
  options?: boolean | AddEventListenerOptions
) {
  if ((type === 'wheel' || type === 'touchstart') && typeof options === 'boolean') {
    options = { capture: options, passive: true }
  } else if ((type === 'wheel' || type === 'touchstart') && options === undefined) {
    options = { passive: true }
  } else if (typeof options === 'object' && options !== null && options.passive === undefined) {
    options = { ...options, passive: true }
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