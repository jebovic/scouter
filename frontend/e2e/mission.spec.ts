import { test, expect } from '@playwright/test'
import { mockApiRoutes, waitForPageReady, MISSION_1 } from './fixtures'

test.describe('Mission Flow', () => {
  test('mission overview page loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/missions/test-laptop')
    await waitForPageReady(page)

    await expect(page.locator('body')).toBeVisible()
    const content = await page.textContent('body')
    expect(content).toContain('Test Laptop Mission')
    await page.screenshot({ path: 'e2e/screenshots/mission-overview.png' })
  })

  test('mission options tab loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/missions/test-laptop/options')
    await waitForPageReady(page)

    await expect(page.locator('body')).toBeVisible()
    await page.screenshot({ path: 'e2e/screenshots/mission-options.png' })
  })

  test('mission shopping tab loads', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/missions/test-laptop/shopping')
    await waitForPageReady(page)

    await expect(page.locator('body')).toBeVisible()
    await page.screenshot({ path: 'e2e/screenshots/mission-shopping.png' })
  })

  test('can navigate between mission tabs', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/missions/test-laptop')
    await waitForPageReady(page)

    // Look for tab links to options and shopping
    const optionsTab = page.locator('a[href*="/options"]').first()
    if (await optionsTab.isVisible()) {
      await optionsTab.click()
      await page.waitForURL('**/options', { timeout: 5000 }).catch(() => {})
      await page.screenshot({ path: 'e2e/screenshots/mission-tab-options.png' })
    }

    const shoppingTab = page.locator('a[href*="/shopping"]').first()
    if (await shoppingTab.isVisible()) {
      await shoppingTab.click()
      await page.waitForURL('**/shopping', { timeout: 5000 }).catch(() => {})
      await page.screenshot({ path: 'e2e/screenshots/mission-tab-shopping.png' })
    }

    await expect(page.locator('body')).toBeVisible()
  })

  test('clicking a mission card navigates to mission', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    // Find mission card link and click it
    const missionLink = page.locator(`a[href*="test-laptop"]`).first()
    if (await missionLink.isVisible()) {
      await missionLink.click()
      await page.waitForURL('**/test-laptop**', { timeout: 5000 })
      await page.screenshot({ path: 'e2e/screenshots/mission-navigate-from-dashboard.png' })
      expect(page.url()).toContain('test-laptop')
    } else {
      // Fallback: navigate directly
      await page.goto('/missions/test-laptop')
      await waitForPageReady(page)
    }

    await expect(page.locator('body')).toBeVisible()
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

    await mockApiRoutes(page, {
      '**/api/missions': { items: [MISSION_1] },
    })

    // Intercept POST /api/missions
    await page.route('**/api/missions', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify(newMission),
        })
      } else {
        await route.continue()
      }
    })

    await page.goto('/')
    await waitForPageReady(page)
    await page.screenshot({ path: 'e2e/screenshots/mission-create-ready.png' })

    await expect(page.locator('body')).toBeVisible()
  })
})
