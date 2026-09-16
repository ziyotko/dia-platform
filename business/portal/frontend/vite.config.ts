import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'url'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_')
  const basePath = env.VITE_BASE_PATH || '/business_portal/'
  const apiBase = env.VITE_API_BASE_URL || '/business_portal/api'
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
      port: 3000,
      open: false,
      proxy: {
        // 代理键必须与 request.ts 的 baseURL / 后端实际路径一致，否则 dev 下请求不会被转发
        [apiBase]: {
          target: 'http://127.0.0.1:8084',
          changeOrigin: true,
          secure: false
        },
        [uploadsBase]: {
          target: 'http://127.0.0.1:8084',
          changeOrigin: true,
          secure: false
        }
      }
    }
  }
})
