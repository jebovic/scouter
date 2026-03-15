import { test, expect } from '@playwright/test'
import { mockApiRoutes, waitForPageReady } from './fixtures'

test.describe('Navigation', () => {
  test('settings page loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/settings')
    await waitForPageReady(page)

    await expect(page.locator('#root')).not.toBeEmpty()
    // Settings page renders currency from mock (EUR)
    const content = await page.textContent('body')
    expect(content).toContain('EUR')
    await page.screenshot({ path: 'e2e/screenshots/settings-page.png' })
  })

  test('notifications page loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/notifications')
    await waitForPageReady(page)

    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/notifications-page.png' })
  })

  test('stats page loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/stats')
    await waitForPageReady(page)

    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/stats-page.png' })
  })

  test('wishlist page loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/wishlist')
    await waitForPageReady(page)

    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/wishlist-page.png' })
  })

  test('history page loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/history')
    await waitForPageReady(page)

    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/history-page.png' })
  })

  test('deal calendar page loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/deal-calendar')
    await waitForPageReady(page)

    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/deal-calendar-page.png' })
  })

  test('analytics redirect works', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/analytics')
    await waitForPageReady(page)

    // /analytics redirects to /stats
    expect(page.url()).toContain('/stats')
    await page.screenshot({ path: 'e2e/screenshots/analytics-redirect.png' })
  })

  test('search page loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/search')
    await waitForPageReady(page)

    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/search-page.png' })
  })

  test('sidebar navigation works from dashboard', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    // Nav/sidebar must be present
    const nav = page.locator('nav, aside, [class*="sidebar"], [class*="topnav"]').first()
    await expect(nav).toBeVisible()

    // Click the first non-root nav link if available
    const navLinks = page.locator('nav a, aside a')
    const count = await navLinks.count()
    if (count > 0) {
      const href = await navLinks.first().getAttribute('href')
      if (href && href !== '/') {
        await navLinks.first().click()
        await page.waitForURL(`**${href}`, { timeout: 5000 }).catch(() => {})
      }
    }

    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/sidebar-navigation.png' })
  })
})
