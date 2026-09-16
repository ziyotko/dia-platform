import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'url'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_')

  return {
    base: env.VITE_BASE_PATH,
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
        '/portal/api': {
          target: 'http://127.0.0.1:8084',
          changeOrigin: true,
          secure: false
        },
        '/portal/uploads': {
          target: 'http://127.0.0.1:8084',
          changeOrigin: true,
          secure: false
        }
      }
    }
  }
})
