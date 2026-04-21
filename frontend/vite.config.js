import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    // Proxy /api requests to the Go server during local development.
    // This means the frontend never needs to know the API URL — it always uses /api/...
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: "dist", // Go embeds this directory
    emptyOutDir: true, // clean the directory before each build
  },
});
