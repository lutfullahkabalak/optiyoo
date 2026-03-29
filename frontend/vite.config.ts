import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    // Cloudflare quick tunnel (trycloudflare.com) farklı Host gönderir; aksi halde Vite isteği reddeder.
    allowedHosts: true,
  },
})
