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

    await expect(page.locator('#root')).not.toBeEmpty()
    // Mission names must NOT appear when list is empty
    const content = await page.textContent('body')
    expect(content).not.toContain('Test Laptop Mission')
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

    const createBtn = page
      .locator('button, [role="button"]')
      .filter({ hasText: /new mission|create|nouvelle|nouveau/i })
      .first()
    await expect(createBtn).toBeVisible()
    await createBtn.click()
    await page.screenshot({ path: 'e2e/screenshots/dashboard-create-modal.png' })
    // Modal or form should have appeared — page must not crash
    await expect(page.locator('#root')).not.toBeEmpty()
  })
})
