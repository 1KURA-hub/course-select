import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "../web/dist",
    emptyOutDir: true
  },
  server: {
    port: 5173,
    proxy: {
      "/login": "http://127.0.0.1:8080",
      "/register": "http://127.0.0.1:8080",
      "/courses": "http://127.0.0.1:8080",
      "/auth": "http://127.0.0.1:8080",
      "/healthz": "http://127.0.0.1:8080"
    }
  }
});
