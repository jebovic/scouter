/**
 * E2E tests for mission-option-management features:
 * - MissionCard overflow menu (archive / delete)
 * - OptionsExplorer "Unpin All" button
 * - ShortlistPanel in buying phase of MissionOverview
 * - HQDashboard "Show archived" toggle
 * - MissionEditModal (edit mission action)
 * - OptionEditModal (edit option action)
 */
import { test, expect } from '@playwright/test'
import { mockApiRoutes, waitForPageReady, MISSION_1, MISSION_1_OPTIONS } from './fixtures'

// ─── HQDashboard ─────────────────────────────────────────────────────────────

test.describe('HQDashboard — Show archived toggle', () => {
  test('shows "Show archived" toggle button with aria-pressed', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    const toggle = page.getByRole('button', { name: 'Show archived' })
    await expect(toggle).toBeVisible()
    await expect(toggle).toHaveAttribute('aria-pressed', 'false')
  })

  test('toggling "Show archived" flips aria-pressed state', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    const toggle = page.getByRole('button', { name: 'Show archived' })
    await toggle.click()
    await expect(toggle).toHaveAttribute('aria-pressed', 'true')

    await toggle.click()
    await expect(toggle).toHaveAttribute('aria-pressed', 'false')
  })
})

// ─── MissionCard — overflow menu ─────────────────────────────────────────────

test.describe('MissionCard — overflow menu', () => {
  test('overflow menu (⋯) opens and shows archive and delete options', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    // Click the first MissionCard's ⋯ menu button (aria-label="more options")
    const menuBtn = page.getByRole('button', { name: 'more options' }).first()
    await menuBtn.click()

    // Archive and Delete options should appear
    await expect(page.getByRole('button', { name: 'Archive' }).last()).toBeVisible()
    await expect(page.getByRole('button', { name: 'Delete' }).last()).toBeVisible()
  })

  test('overflow menu closes when clicking outside', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto('/')
    await waitForPageReady(page)

    const menuBtn = page.getByRole('button', { name: 'more options' }).first()
    await menuBtn.click()

    // Confirm the menu is open
    const deleteBtn = page.getByRole('button', { name: 'Delete' }).last()
    await expect(deleteBtn).toBeVisible()

    // Click somewhere outside the menu
    await page.mouse.click(10, 10)
    await expect(deleteBtn).not.toBeVisible()
  })
})

// ─── OptionsExplorer — Unpin All ─────────────────────────────────────────────

test.describe('OptionsExplorer — Unpin All button', () => {
  test('shows "Unpin all" button when pinned options exist', async ({ page }) => {
    // Fixtures already have one pinned option (Synology DS923+ — pinned: true)
    await mockApiRoutes(page)
    await page.goto(`/missions/${MISSION_1.slug}/options`)
    await waitForPageReady(page)

    // Button text: "Unpin all (1)" — matches t('option.actions.unpinAll', { count: 1 })
    await expect(page.getByRole('button', { name: /Unpin all/i })).toBeVisible()
  })

  test('does not show "Unpin all" when no options are pinned', async ({ page }) => {
    const unpinnedOptions = MISSION_1_OPTIONS.map((o) => ({ ...o, pinned: false }))
    await mockApiRoutes(page, {
      [`**/api/missions/${MISSION_1.id}/options`]: {
        items: unpinnedOptions,
        total: unpinnedOptions.length,
        page: 1,
        limit: 20,
      },
    })
    await page.goto(`/missions/${MISSION_1.slug}/options`)
    await waitForPageReady(page)

    await expect(page.getByRole('button', { name: /Unpin all/i })).not.toBeVisible()
  })
})

// ─── MissionOverview — MissionEditModal ──────────────────────────────────────

test.describe('MissionOverview — MissionEditModal', () => {
  test('clicking "Edit mission" opens the edit modal', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto(`/missions/${MISSION_1.slug}`)
    await waitForPageReady(page)

    // MissionActionBar renders an "Edit mission" button
    const editBtn = page.getByRole('button', { name: 'Edit mission' })
    await expect(editBtn).toBeVisible()
    await editBtn.click()

    // The modal appears with an h2 heading "Edit mission"
    await expect(page.getByRole('heading', { name: 'Edit mission' })).toBeVisible()
  })

  test('closing MissionEditModal via Cancel hides the modal', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto(`/missions/${MISSION_1.slug}`)
    await waitForPageReady(page)

    await page.getByRole('button', { name: 'Edit mission' }).click()
    const heading = page.getByRole('heading', { name: 'Edit mission' })
    await expect(heading).toBeVisible()

    // Close it via the Cancel button inside the form
    await page.getByRole('button', { name: 'Cancel' }).click()
    await expect(heading).not.toBeVisible()
  })
})

// ─── MissionOverview — ShortlistPanel in buying phase ────────────────────────

test.describe('MissionOverview — ShortlistPanel (buying phase)', () => {
  test('shows ShortlistPanel when mission is in buying phase', async ({ page }) => {
    const buyingMission = { ...MISSION_1, phase: 'buying' }
    await mockApiRoutes(page, {
      '**/api/missions/home-server-2026': buyingMission,
    })
    await page.goto(`/missions/${MISSION_1.slug}`)
    await waitForPageReady(page)

    // ShortlistPanel renders with heading "Your Shortlist"
    await expect(page.getByText('Your Shortlist')).toBeVisible()
  })

  test('does not show ShortlistPanel when mission is in comparing phase', async ({ page }) => {
    // Default fixture has phase: 'comparing'
    await mockApiRoutes(page)
    await page.goto(`/missions/${MISSION_1.slug}`)
    await waitForPageReady(page)

    await expect(page.getByText('Your Shortlist')).not.toBeVisible()
  })
})

// ─── OptionsExplorer — OptionEditModal ───────────────────────────────────────

test.describe('OptionsExplorer — OptionEditModal', () => {
  test('clicking "Edit option" on an option card opens the edit modal', async ({ page }) => {
    await mockApiRoutes(page)
    await page.goto(`/missions/${MISSION_1.slug}/options`)
    await waitForPageReady(page)

    // Each OptionCard has a ⋯ menu button with aria-label="more options"
    const menuBtn = page.getByLabel('more options').first()
    await menuBtn.click()

    // The "Edit option" menu item should appear
    const editOptionItem = page.getByRole('menuitem', { name: 'Edit option' })
    await expect(editOptionItem).toBeVisible()
    await editOptionItem.click()

    // OptionEditModal appears with an h2 heading "Edit option"
    await expect(page.getByRole('heading', { name: 'Edit option' })).toBeVisible()
  })
})
