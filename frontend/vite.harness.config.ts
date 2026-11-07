import { defineConfig } from "vite"
import tailwindcss from "@tailwindcss/vite"
import path from "path"

// Isolated config for the parity harness: renders the real React components
// with fixture data so Fyne can be diffed against measured truth.
export default defineConfig({
  plugins: [tailwindcss()],
  root: __dirname,
  resolve: {
    alias: { "@": path.resolve(__dirname, "./src") },
  },
  server: { port: 5199, strictPort: true },
})
