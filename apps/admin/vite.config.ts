import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

const packageJson = JSON.parse(
  readFileSync(new URL('./package.json', import.meta.url), 'utf8'),
) as { version: string }

export default defineConfig({
  plugins: [react(), tailwindcss()],
  define: {
    __MARKU_VERSION__: JSON.stringify(packageJson.version),
  },
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
