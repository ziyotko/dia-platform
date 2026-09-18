import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'url'

export default defineConfig(({ mode }) => {
  // 与 portal / member 一致：部署子路径与 API 前缀统一从 .env 读取（VITE_BASE_PATH / VITE_API_BASE_URL），
  // 保证 build 产物的资源路径与 request.ts 的 baseURL 同源同值。
  const env = loadEnv(mode, process.cwd(), 'VITE_')
  const basePath = env.VITE_BASE_PATH || '/application/'
  const apiBase = env.VITE_API_BASE_URL || '/application/api'
  // 上传目录挂在同一个部署子路径下（后端 server.upload_dir_prefix = VITE_BASE_PATH 去尾斜杠）
  const uploadsBase = `${basePath.replace(/\/+$/, '')}/uploads`

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
      port: 3003,
      open: false,
      proxy: {
        // 代理键需与 request.ts 的 baseURL 一致，否则 dev 下请求不会被转发到后端
        [apiBase]: {
          target: 'http://127.0.0.1:8094',
          changeOrigin: true
        },
        [uploadsBase]: {
          target: 'http://127.0.0.1:8094',
          changeOrigin: true
        }
      }
    }
  }
})
