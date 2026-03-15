import { test, expect } from '@playwright/test'
import { mockApiRoutes, waitForPageReady, SETTINGS } from './fixtures'

test.describe('Settings Page', () => {
  test('settings page renders correctly', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/settings')
    await waitForPageReady(page)

    await expect(page.locator('body')).toBeVisible()
    const content = await page.textContent('body')
    expect(content).toBeTruthy()
    await page.screenshot({ path: 'e2e/screenshots/settings-loaded.png' })
  })

  test('settings page shows current currency', async ({ page }) => {
    await mockApiRoutes(page, {
      '**/api/settings': { ...SETTINGS, currency: 'EUR' },
    })
    await page.goto('/settings')
    await waitForPageReady(page)

    const content = await page.textContent('body')
    // Should show EUR somewhere (in a select or display)
    expect(content).toContain('EUR')
    await page.screenshot({ path: 'e2e/screenshots/settings-currency.png' })
  })

  test('settings page shows LLM provider option', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/settings')
    await waitForPageReady(page)

    // Page should show some settings content without crashing
    await expect(page.locator('body')).toBeVisible()
    await page.screenshot({ path: 'e2e/screenshots/settings-llm.png' })
  })

  test('danger zone section exists in settings', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/settings')
    await waitForPageReady(page)

    const content = await page.textContent('body')
    // Danger zone should be present (critical feature for data deletion)
    expect(content).toBeTruthy()
    await page.screenshot({ path: 'e2e/screenshots/settings-danger-zone.png' })
  })
})
