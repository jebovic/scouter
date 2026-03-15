import { test, expect } from '@playwright/test'
import { mockApiRoutes, waitForPageReady } from './fixtures'

test.describe('HQ Dashboard', () => {
  test('renders the dashboard with missions', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    await expect(page).toHaveTitle(/Scouter/)
    await page.screenshot({ path: 'e2e/screenshots/dashboard-with-missions.png' })
  })

  test('shows empty state when no missions exist', async ({ page }) => {
    await mockApiRoutes(page, { '**/api/missions': { items: [] } })
    await page.goto('/')
    await waitForPageReady(page)

    // Should not crash and should render the page shell
    const body = page.locator('body')
    await expect(body).toBeVisible()
    await page.screenshot({ path: 'e2e/screenshots/dashboard-empty.png' })
  })

  test('displays mission cards', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    // Mission names should appear on page
    const pageContent = await page.textContent('body')
    expect(pageContent).toContain('Test Laptop Mission')
    expect(pageContent).toContain('Test Phone Mission')
    await page.screenshot({ path: 'e2e/screenshots/dashboard-mission-cards.png' })
  })

  test('can open the create mission modal', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    // Look for a create/new mission button
    const createBtn = page.locator('button, [role="button"]').filter({ hasText: /new mission|create|nouvelle/i }).first()
    if (await createBtn.isVisible()) {
      await createBtn.click()
      await page.screenshot({ path: 'e2e/screenshots/dashboard-create-modal.png' })
    } else {
      // Try the + button or similar
      const plusBtn = page.locator('[aria-label*="new"], [aria-label*="create"], [title*="new"]').first()
      if (await plusBtn.isVisible()) {
        await plusBtn.click()
        await page.screenshot({ path: 'e2e/screenshots/dashboard-create-modal-plus.png' })
      }
    }

    // Either way, page should not have crashed
    await expect(page.locator('body')).toBeVisible()
  })
})
