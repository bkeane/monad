import { defineConfig } from 'vite'
import svgLoader from 'vite-svg-loader'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), svgLoader()],
  base: '/monad/',
  build: {
    emptyOutDir: true,
    outDir: '../dist',
    rollupOptions: {
      output:{
          manualChunks(id) {
              if (id.includes('node_modules')) {
                  return id.toString().split('node_modules/')[1].split('/')[0].toString();
              }
          }
      }
    }
  }
})
