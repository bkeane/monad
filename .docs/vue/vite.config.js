import { defineConfig } from 'vite'
import svgLoader from 'vite-svg-loader'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), svgLoader()],
  build: {
    emptyOutDir: true,
    outDir: '../docs',
  },
})
