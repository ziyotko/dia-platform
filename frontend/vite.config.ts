import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 3000,
    open: false,
    host: true,
    proxy: {
      '/miicapi': {
        target: 'http://10.1.100.35:8080',
        changeOrigin: true,
        secure: false
      },
      '/uploads': {
        target: 'http://10.1.100.35:8080',
        changeOrigin: true, 
        secure: false
      }
    }
  }
})