import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'url'

export default defineConfig(({ mode }) => {
  // 与 portal / member / application 一致：部署子路径与 API 前缀统一从 .env 读取
  // （VITE_BASE_PATH / VITE_API_BASE_URL），保证 build 产物资源路径与 request.ts 的 baseURL 同源同值。
  const env = loadEnv(mode, process.cwd(), 'VITE_')
  const basePath = env.VITE_BASE_PATH || '/business_base/'
  const apiBase = env.VITE_API_BASE_URL || '/business_base/api'

  return {
    base: basePath,
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    },
    server: {
      host: true,
      port: 3000,
      open: false,
      proxy: {
        // 代理键需与 request.ts 的 baseURL 一致，否则 dev 下请求不会被转发到后端
        [apiBase]: {
          target: 'http://127.0.0.1:8080',
          changeOrigin: true
        }
      }
    }
  }
})
