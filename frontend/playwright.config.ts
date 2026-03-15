import { defineConfig, devices } from '@playwright/test'

const isDocker = !!process.env.E2E_DOCKER

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  workers: 2,
  reporter: [['html', { open: 'never' }]],
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'on',
    video: 'off',
  },
  projects: [
    {
      name: 'chrome',
      use: { ...devices['Desktop Chrome'], channel: 'chrome' },
      testIgnore: ['**/smoke.spec.ts'],
    },
    ...(isDocker
      ? [
          {
            name: 'smoke',
            use: { ...devices['Desktop Chrome'], channel: 'chrome' },
            testMatch: ['**/smoke.spec.ts'],
          },
        ]
      : []),
  ],
  webServer: isDocker
    ? undefined
    : {
        command: 'VITE_API_BASE="" npm run dev -- --port 5173',
        url: 'http://localhost:5173',
        reuseExistingServer: true,
        timeout: 30000,
      },
})
