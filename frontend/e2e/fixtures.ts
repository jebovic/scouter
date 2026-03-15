/**
 * Shared API mock fixtures for E2E tests.
 * All requests go to /api/** (VITE_API_BASE="" so calls are same-origin).
 */
import type { Page, Route } from '@playwright/test'

// ─── Mock data ─────────────────────────────────────────────────────────────

export const MISSION_1 = {
  id: '11111111-1111-4111-8111-111111111111',
  slug: 'test-laptop',
  name: 'Test Laptop Mission',
  icon: '💻',
  category: 'electronics',
  autotagCategory: null,
  budget: 1500,
  currency: 'EUR',
  locale: 'fr-FR',
  phase: 'researching',
  constraints: [],
  costCategories: [],
  timeline: [],
  weightProfile: { price: 0.5, quality: 0.3, feature: 0.2 },
  lessons: null,
  envelopeId: null,
  archivedAt: null,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}

export const MISSION_2 = {
  id: '22222222-2222-4222-8222-222222222222',
  slug: 'test-phone',
  name: 'Test Phone Mission',
  icon: '📱',
  category: 'electronics',
  autotagCategory: null,
  budget: 800,
  currency: 'EUR',
  locale: 'fr-FR',
  phase: 'comparing',
  constraints: [],
  costCategories: [],
  timeline: [],
  weightProfile: { price: 0.4, quality: 0.4, feature: 0.2 },
  lessons: null,
  envelopeId: null,
  archivedAt: null,
  createdAt: '2026-01-02T00:00:00Z',
  updatedAt: '2026-01-02T00:00:00Z',
}

export const SETTINGS = {
  currency: 'EUR',
  locale: 'fr-FR',
  llm_provider: 'ollama',
}

export const HEALTH = { status: 'ok', db: 'ok' }

export const LLM_HEALTH = {
  status: 'healthy',
  models: [],
}

export const USAGE = {
  missions_count: 2,
  options_count: 5,
  research_runs: 3,
}

export const TEMPLATES = {
  items: [
    {
      id: 'tmpl-electronics',
      slug: 'electronics',
      name: 'Electronics Purchase',
      icon: '🔌',
      description: 'Research and compare electronics',
      category: 'electronics',
      constraints: [],
      costCategories: [],
      timeline: [],
    },
  ],
}

export const DEAL_CALENDAR = { events: [] }

export const NOTIFICATIONS_EMPTY = []

export const UNREAD_COUNT = { count: 0 }

export const WISHLIST_EMPTY = { items: [] }

export const ENVELOPES_EMPTY = { items: [] }

export const STATS = {
  total_spent: 0,
  purchase_count: 0,
  avg_savings: 0,
  categories: [],
}

// ─── Route helpers ─────────────────────────────────────────────────────────

type MockRoutes = Record<string, unknown>

/** Apply a map of URL patterns → JSON responses to the page. */
export async function mockApiRoutes(page: Page, overrides: MockRoutes = {}): Promise<void> {
  // Dismiss the onboarding overlay so it doesn't block test assertions
  await page.addInitScript(() => {
    localStorage.setItem('scouter_onboarding_dismissed', 'true')
  })
  const defaults: MockRoutes = {
    '**/api/health': HEALTH,
    '**/api/health/llm': LLM_HEALTH,
    '**/api/missions': { items: [MISSION_1, MISSION_2] },
    '**/api/missions/test-laptop': MISSION_1,
    '**/api/missions/test-phone': MISSION_2,
    '**/api/settings': SETTINGS,
    '**/api/usage': USAGE,
    '**/api/templates': TEMPLATES,
    '**/api/deal-calendar': DEAL_CALENDAR,
    '**/api/notifications': NOTIFICATIONS_EMPTY,
    '**/api/notifications/unread-count': UNREAD_COUNT,
    '**/api/wishlist': WISHLIST_EMPTY,
    '**/api/wishlist/prioritized': WISHLIST_EMPTY,
    '**/api/envelopes': ENVELOPES_EMPTY,
    '**/api/stats': STATS,
    '**/api/stats/monthly': { months: [] },
    '**/api/analytics/spending': { data: [] },
    '**/api/analytics/budget-heatmap': { data: [] },
    '**/api/analytics/cross-mission': { data: [] },
    '**/api/search': { results: [] },
    '**/api/loyalty/programs': { items: [] },
    '**/api/currency/rates': { rates: {} },
    '**/api/seasonal': { events: [] },
    '**/api/persona': { type: 'balanced', score: 0.5 },
    '**/api/kanban/columns': { columns: [] },
    '**/api/cashback/programs': { items: [] },
    '**/api/missions/*/options': { items: [], total: 0, page: 1, limit: 20 },
    '**/api/missions/*/shopping': { items: [], total: 0, page: 1, limit: 20 },
    '**/api/missions/*/agent-runs': { items: [] },
    '**/api/missions/*/coach': { suggestions: [] },
    '**/api/missions/*/scorecard': { grade: 'A', score: 95, achievements: [] },
    '**/api/missions/*/french-benchmark': { median: 1200, verdict: 'bon_prix' },
    '**/api/missions/*/purchase-timeline': { weeks: [] },
    '**/api/missions/*/purchase': null,
    '**/api/missions/*/invites': { items: [] },
    '**/api/missions/*/collaborators': { items: [] },
    '**/api/missions/*/smart-alerts': { alerts: [] },
    '**/api/missions/*/timing-advice': { advice: null },
    '**/api/missions/*/budget-plan': { phases: [] },
    '**/api/missions/*/receipts': { items: [] },
  }

  const routes = { ...defaults, ...overrides }

  await page.route('**/api/**', async (route: Route) => {
    const url = route.request().url()
    const pathname = new URL(url).pathname

    // Do not intercept Vite source module requests (e.g. /src/api/usage.ts)
    if (pathname.startsWith('/src/') || pathname.startsWith('/@')) {
      await route.continue()
      return
    }

    // Find matching pattern — sort by length descending so specific routes win
    const sortedRoutes = Object.entries(routes).sort(([a], [b]) => b.length - a.length)
    for (const [pattern, body] of sortedRoutes) {
      // Match against pathname only, with anchors, to avoid /api/missions matching /api/missions/slug
      const patternPath = pattern
        .replace(/^\*\*\//, '/') // strip leading **/
        .replace(/\*\*/g, '.*')
        .replace(/\*/g, '[^/]*')
      const regex = new RegExp(`^${patternPath}$`)
      if (regex.test(pathname)) {
        if (body === null) {
          await route.fulfill({ status: 404, contentType: 'application/json', body: '{"error":"not found"}' })
        } else {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify(body),
          })
        }
        return
      }
    }

    // Fallback: 500 so missing mock definitions surface immediately
    const pathname2 = new URL(route.request().url()).pathname
    await route.fulfill({
      status: 500,
      contentType: 'application/json',
      body: JSON.stringify({ error: `unmatched mock route: ${pathname2}` }),
    })
  })
}

/** Wait for the page to finish loading (network idle + React mounted + data resolved). */
export async function waitForPageReady(page: Page): Promise<void> {
  await page.waitForLoadState('load')
  // Wait until React has mounted something into #root (state: 'attached' bypasses visibility check)
  await page.waitForSelector('#root > *', { state: 'attached', timeout: 15000 })
  // Wait for all mocked API responses to settle (network idle = React Query flushed)
  await page.waitForLoadState('networkidle')
}
