import { defineConfig } from 'vitest/config'

export default defineConfig({
  build: {
    target: 'es2022',
    sourcemap: true,
    lib: {
      entry: 'src/index.ts',
      formats: ['es'],
      fileName: () => 'marku-client.js',
    },
  },
  test: {
    environment: 'jsdom',
    restoreMocks: true,
  },
})
