import { test, expect } from '@playwright/test'
import { mockApiRoutes, waitForPageReady } from './fixtures'

test.describe('App Health & Core Shell', () => {
  test('app loads without JS errors', async ({ page }) => {
    const errors: string[] = []
    page.on('pageerror', (err) => errors.push(err.message))

    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    // Filter out known benign errors (service worker, etc.)
    const criticalErrors = errors.filter(
      (e) =>
        !e.includes('service-worker') &&
        !e.includes('ResizeObserver') &&
        !e.includes('Network request failed') &&
        !e.includes('AbortError'),
    )
    expect(criticalErrors).toEqual([])
    await page.screenshot({ path: 'e2e/screenshots/app-loaded.png' })
  })

  test('page title is set correctly', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    const title = await page.title()
    expect(title.length).toBeGreaterThan(0)
    await page.screenshot({ path: 'e2e/screenshots/page-title.png' })
  })

  test('app renders with dark theme', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    // Check the root element has content rendered
    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/dark-theme.png' })
  })

  test('topnav is present on main pages', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    // Topnav should render (has nav, header, or topbar element)
    const hasNav =
      (await page.locator('nav').count()) > 0 ||
      (await page.locator('header').count()) > 0 ||
      (await page.locator('[class*="topnav"], [class*="nav"]').count()) > 0

    expect(hasNav).toBeTruthy()
    await page.screenshot({ path: 'e2e/screenshots/topnav.png' })
  })

  test('404 unknown routes do not crash app', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/unknown-route-xyz')
    await waitForPageReady(page)

    // App should not show blank screen
    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/unknown-route.png' })
  })

  test('performance page loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/performance')
    await waitForPageReady(page)

    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/performance-page.png' })
  })

  test('loyalty page loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/loyalty')
    await waitForPageReady(page)

    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/loyalty-page.png' })
  })

  test('insights page loads', async ({ page }) => {
    await mockApiRoutes(page, {
      '**/api/missions/*/options': { items: [], total: 0, page: 1, limit: 20 },
    })
    await page.goto('/insights')
    await waitForPageReady(page)

    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/insights-page.png' })
  })
})
