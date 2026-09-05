import { defineConfig } from 'vite'

export default defineConfig({
  server: {
    port: 3006,
    host: true,
    proxy: {
      '/v1': {
        target: 'http://127.0.0.1:8000',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    target: 'es2020',
    sourcemap: false
  }
})
