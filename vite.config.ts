import { defineConfig } from 'vite'
import createVuePlugin from '@vitejs/plugin-vue'
import { visualizer } from 'rollup-plugin-visualizer'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    createVuePlugin(),
    visualizer({
      gzipSize: true,
      brotliSize: true,
      emitFile: false,
      filename: "tmp/analyzer.html",
      open: true,
    }),
  ],
  server: {
    proxy: {
      '/api/ws/shell': {
        target: 'http://localhost:5000',
        ws: true,
        changeOrigin: true,
        secure: false
      },
      '/api/ws/judge': {
        target: 'http://localhost:5000',
        ws: true,
        changeOrigin: true,
        secure: false
      },
      '^/api/.*': {
        target: 'http://localhost:5000',
        ws: true,
        changeOrigin: true,
        secure: false
      }
    }
  }
})