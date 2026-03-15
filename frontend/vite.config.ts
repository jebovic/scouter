import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg}'],
        runtimeCaching: [
          {
            urlPattern: /^\/api\//,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-cache',
              networkTimeoutSeconds: 10,
              expiration: { maxEntries: 100, maxAgeSeconds: 300 },
            },
          },
        ],
      },
      manifest: {
        name: 'Scouter — Mission Intel',
        short_name: 'Scouter',
        description: 'Personal spending intelligence — research, compare, and budget any purchase',
        theme_color: '#00d4ff',
        background_color: '#0a0a0f',
        display: 'standalone',
        start_url: '/',
        icons: [{ src: '/icons/icon.svg', sizes: 'any', type: 'image/svg+xml' }],
      },
    }),
  ],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
