import { test, expect } from '@playwright/test'
import { mockApiRoutes, waitForPageReady, MISSION_1 } from './fixtures'

test.describe('Mission Flow', () => {
  test('mission overview page loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/missions/test-laptop')
    await waitForPageReady(page)

    const content = await page.textContent('body')
    expect(content).toContain('Test Laptop Mission')
    await page.screenshot({ path: 'e2e/screenshots/mission-overview.png' })
  })

  test('mission options tab loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/missions/test-laptop/options')
    await waitForPageReady(page)

    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/mission-options.png' })
  })

  test('mission shopping tab loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/missions/test-laptop/shopping')
    await waitForPageReady(page)

    await expect(page.locator('#root')).not.toBeEmpty()
    await page.screenshot({ path: 'e2e/screenshots/mission-shopping.png' })
  })

  test('can navigate between mission tabs', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/missions/test-laptop')
    await waitForPageReady(page)

    const optionsTab = page.locator('a[href*="/options"]').first()
    await expect(optionsTab).toBeVisible()
    await optionsTab.click()
    await page.waitForURL('**/options', { timeout: 5000 })
    await page.screenshot({ path: 'e2e/screenshots/mission-tab-options.png' })

    const shoppingTab = page.locator('a[href*="/shopping"]').first()
    await expect(shoppingTab).toBeVisible()
    await shoppingTab.click()
    await page.waitForURL('**/shopping', { timeout: 5000 })
    await page.screenshot({ path: 'e2e/screenshots/mission-tab-shopping.png' })
  })

  test('clicking a mission card navigates to mission', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    const missionLink = page.locator('a[href*="test-laptop"]').first()
    await expect(missionLink).toBeVisible()
    await missionLink.click()
    await page.waitForURL('**/test-laptop**', { timeout: 5000 })
    await page.screenshot({ path: 'e2e/screenshots/mission-navigate-from-dashboard.png' })
    expect(page.url()).toContain('test-laptop')
  })

  test('mission create and mock response', async ({ page }) => {
    const newMission = {
      ...MISSION_1,
      id: '33333333-3333-3333-3333-333333333333',
      slug: 'new-test-mission',
      name: 'New Test Mission',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }

    // Handle both GET and POST for /api/missions to avoid route.continue() hitting real server
    await mockApiRoutes(page, {
      '**/api/missions': { items: [MISSION_1] },
    })

    await page.route('**/api/missions', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify(newMission),
        })
      } else {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ items: [MISSION_1] }),
        })
      }
    })

    await page.goto('/')
    await waitForPageReady(page)
    await page.screenshot({ path: 'e2e/screenshots/mission-create-ready.png' })

    await expect(page.locator('#root')).not.toBeEmpty()
  })
})
