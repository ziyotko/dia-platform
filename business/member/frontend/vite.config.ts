import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  base: '/member/',
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
        host: true,
      port: 3001,
      open: false,
    proxy: {
      '/member/api': {
        target: 'http://127.0.0.1:8085',
        changeOrigin: true
      },
      '/uploads': {
        target: 'http://127.0.0.1:8085',
        changeOrigin: true
      }
    }
  }
})
