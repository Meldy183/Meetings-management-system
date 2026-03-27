import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    // In dev mode the frontend runs on :5173 and the backend on :8080 (cross-origin).
    // SameSite=Lax cookies are not sent on cross-origin fetch, so the session cookie
    // would be silently dropped. This proxy makes all API calls appear same-origin.
    proxy: {
      '/auth': 'http://localhost:8080',
      '/people': 'http://localhost:8080',
      '/meetings': 'http://localhost:8080',
      '/health': 'http://localhost:8080',
    },
  },
})
