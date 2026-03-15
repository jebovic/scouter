/**
 * Smoke tests — run only when E2E_DOCKER=1 (docker-compose stack is up).
 * No API mocking; hits the real backend at http://localhost:8080.
 * Tests are read-only to avoid polluting the database.
 */
import { test, expect } from '@playwright/test'

const isDocker = !!process.env.E2E_DOCKER

test.describe('Docker Stack Smoke', () => {
  test.skip(!isDocker, 'Smoke tests require E2E_DOCKER=1 (docker-compose stack)')

  test('GET /api/health returns ok', async ({ request }) => {
    const res = await request.get('http://localhost:8080/api/health')
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(body.status).toBe('ok')
    expect(body.db).toBe('ok')
  })

  test('GET /api/templates returns items array', async ({ request }) => {
    const res = await request.get('http://localhost:8080/api/templates')
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(Array.isArray(body.items)).toBe(true)
    expect(body.items.length).toBeGreaterThan(0)
  })

  test('GET /api/settings returns valid config', async ({ request }) => {
    const res = await request.get('http://localhost:8080/api/settings')
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(body).toHaveProperty('currency')
    expect(body).toHaveProperty('locale')
  })

  test('frontend serves app shell', async ({ page }) => {
    await page.goto('http://localhost:5173/')
    await page.waitForSelector('#root > *', { state: 'attached', timeout: 15000 })
    await expect(page.locator('#root')).not.toBeEmpty()
    // Topnav should be present (proves nginx proxy_pass to backend works on initial load)
    const nav = page.locator('nav, header, [class*="topnav"]').first()
    await expect(nav).toBeVisible()
  })
})
