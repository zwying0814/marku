import { fileURLToPath } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 36581,
    strictPort: true,
    proxy: {
      '/api/v1': {
        target: process.env.MARKU_API_PROXY_TARGET ?? 'http://localhost:36580',
        changeOrigin: true,
      },
    },
  },
  preview: {
    port: 36581,
    strictPort: true,
  },
})
