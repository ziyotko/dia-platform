/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module '@wangeditor/editor-for-vue' {
  import { DefineComponent } from 'vue'
  export const Editor: DefineComponent<{}, {}, any>
  export const Toolbar: DefineComponent<{}, {}, any>
}

declare module 'echarts-wordcloud' {}