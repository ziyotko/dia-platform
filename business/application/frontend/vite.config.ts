import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  base: '/application/',
  resolve: {
    alias: { '@': resolve(__dirname, 'src') }
  },
  server: {
    host: true,
    port: 3003,
    open: false,
    proxy: {
      '/application/api': {
        target: 'http://127.0.0.1:8087',
        changeOrigin: true
      },
      '/uploads': {
        target: 'http://127.0.0.1:8087',
        changeOrigin: true
      }
    }
  }
})
