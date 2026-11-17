import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/api": { target: "https://127.0.0.1:8443", secure: false, changeOrigin: true },
      "/ws": { target: "wss://127.0.0.1:8443", secure: false, ws: true },
    },
  },
});
