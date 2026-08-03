import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  base: '/conference/',
  resolve: {
    alias: { '@': resolve(__dirname, 'src') }
  },
  server: {
    host: true,
    port: 3002,
    open: false,
    proxy: {
      '/conference/api': {
        target: 'http://127.0.0.1:8086',
        changeOrigin: true
      },
      '/uploads': {
        target: 'http://127.0.0.1:8086',
        changeOrigin: true
      }
    }
  }
})
