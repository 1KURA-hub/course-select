import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/register": "http://43.136.63.219:8080",
      "/login": "http://43.136.63.219:8080",
      "/courses": "http://43.136.63.219:8080",
      "/auth": "http://43.136.63.219:8080",
      "/healthz": "http://43.136.63.219:8080",
    },
  },
})
