import { fileURLToPath } from 'node:url'

import { defineConfig } from 'vitest/config'

export default defineConfig({
  resolve: {
    alias: {
      '@marku/client': fileURLToPath(
        new URL('../client/src/index.ts', import.meta.url),
      ),
    },
  },
  build: {
    target: 'es2022',
    sourcemap: true,
    lib: {
      entry: 'src/index.ts',
      name: 'MarkuEmbed',
      formats: ['es', 'iife'],
      fileName: (format) =>
        format === 'iife' ? 'marku-embed.iife.js' : 'marku-embed.js',
    },
  },
  test: {
    environment: 'jsdom',
    restoreMocks: true,
  },
})
