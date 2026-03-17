/**
 * Shared API mock fixtures for E2E tests.
 * All requests go to /api/** (VITE_API_BASE="" so calls are same-origin).
 *
 * Primary fixture mission: "Home Server Build" (slug: home-server-2026)
 * This mirrors the seed data in backend/cmd/seed/main.go so smoke tests
 * and mocked tests use the same realistic dataset.
 */
import type { Page, Route } from '@playwright/test'

// ─── Mission fixtures ────────────────────────────────────────────────────────

/** Primary E2E fixture: Build a home server, budget ~2000€ */
export const MISSION_1 = {
  id: '11111111-1111-4111-8111-111111111111',
  slug: 'home-server-2026',
  name: 'Home Server Build',
  icon: '🖥️',
  category: 'computing',
  autotagCategory: 'computing',
  budget: 2000,
  currency: 'EUR',
  locale: 'fr-FR',
  phase: 'comparing',
  constraints: [
    { key: 'noise_level', label: 'Noise level', value: 'Silent or near-silent (living space)', type: 'hard' },
    { key: 'idle_power', label: 'Idle power draw', value: 'Under 50W at idle', type: 'hard' },
    { key: 'form_factor', label: 'Form factor', value: 'Mini-ITX or NAS enclosure preferred', type: 'soft' },
    { key: 'storage_bays', label: 'Storage bays', value: 'At least 4 drive bays for future expansion', type: 'soft' },
  ],
  costCategories: ['cpu_motherboard', 'memory', 'storage', 'case_psu', 'networking'],
  timeline: [
    { date: '2026-04-15', label: 'Components research deadline', urgent: false },
    { date: '2026-05-01', label: 'Purchase target (spring sales)', urgent: true },
    { date: '2026-06-01', label: 'Setup and configuration done', urgent: false },
  ],
  weightProfile: { price: 0.4, quality: 0.35, feature: 0.25 },
  lessons: null,
  envelopeId: null,
  archivedAt: null,
  createdAt: '2026-03-01T10:00:00Z',
  updatedAt: '2026-03-10T14:30:00Z',
}

/** Secondary fixture mission: Summer holiday for contrast */
export const MISSION_2 = {
  id: '22222222-2222-4222-8222-222222222222',
  slug: 'summer-holiday-2026',
  name: 'Summer Holiday 2026',
  icon: '🏖️',
  category: 'travel',
  autotagCategory: 'travel',
  budget: 4000,
  currency: 'EUR',
  locale: 'fr-FR',
  phase: 'researching',
  constraints: [
    { key: 'max_flight_duration', label: 'Max flight duration', value: '8h', type: 'hard' },
    { key: 'travel_style', label: 'Travel style', value: 'Mix of beach and culture', type: 'soft' },
    { key: 'accommodation', label: 'Accommodation', value: '4-star hotel or quality Airbnb', type: 'soft' },
  ],
  costCategories: ['flights', 'accommodation', 'transport', 'activities', 'food'],
  timeline: [
    { date: '2026-06-01', label: 'Booking deadline (best fares)', urgent: true },
    { date: '2026-07-15', label: 'Departure target', urgent: false },
    { date: '2026-07-29', label: 'Return date', urgent: false },
  ],
  weightProfile: { price: 0.5, quality: 0.3, feature: 0.2 },
  lessons: null,
  envelopeId: null,
  archivedAt: null,
  createdAt: '2026-02-15T09:00:00Z',
  updatedAt: '2026-02-15T09:00:00Z',
}

// ─── Options fixtures (home server) ─────────────────────────────────────────

export const MISSION_1_OPTIONS = [
  {
    id: 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',
    missionId: '11111111-1111-4111-8111-111111111111',
    name: 'Synology DS923+',
    category: 'NAS',
    badge: 'recommended',
    pinned: true,
    rejected: false,
    rejectReason: null,
    attributes: [
      { key: 'cpu', label: 'CPU', value: 'AMD Ryzen R1600 dual-core 2.6 GHz', type: 'text' },
      { key: 'ram', label: 'RAM', value: '4 GB DDR4 ECC (expandable to 32 GB)', type: 'text' },
      { key: 'bays', label: 'Drive bays', value: '4 (expandable to 9 with DX517)', type: 'text' },
      { key: 'idle_power', label: 'Idle power', value: '~19W (HDD off)', type: 'text' },
      { key: 'os', label: 'OS', value: 'Synology DSM 7.2', type: 'text' },
      { key: 'noise', label: 'Noise', value: '21.7 dB(A) — near silent', type: 'text' },
      { key: 'price_score', label: 'Value score', value: 78, type: 'score', max: 100 },
    ],
    priceRange: { min: 580, max: 650, best: 609 },
    notes: 'Best all-in-one NAS for home use. Excellent DSM ecosystem, apps, and mobile integrations. Drive cost extra (~€320 for 2×4TB WD Red Plus). Total budget: ~€930.',
    warnings: ['Drives sold separately', 'Proprietary OS limits custom software'],
    url: 'https://www.synology.com/fr-fr/products/DS923+',
    createdAt: '2026-03-02T10:00:00Z',
  },
  {
    id: 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb',
    missionId: '11111111-1111-4111-8111-111111111111',
    name: 'Custom mini-ITX — Intel N100',
    category: 'Custom build',
    badge: 'alternative',
    pinned: false,
    rejected: false,
    rejectReason: null,
    attributes: [
      { key: 'cpu', label: 'CPU', value: 'Intel N100 (4C/4T, 3.4 GHz burst, 6W TDP)', type: 'text' },
      { key: 'ram', label: 'RAM', value: '32 GB DDR4 SO-DIMM', type: 'text' },
      { key: 'bays', label: 'Drive bays', value: '6 (Fractal Node 304)', type: 'text' },
      { key: 'idle_power', label: 'Idle power', value: '~8–12W with SSD, ~18W with HDDs', type: 'text' },
      { key: 'os', label: 'OS', value: 'TrueNAS SCALE / Proxmox + TrueNAS VM', type: 'text' },
      { key: 'noise', label: 'Noise', value: 'Near-silent with Noctua fan swap', type: 'text' },
      { key: 'price_score', label: 'Value score', value: 88, type: 'score', max: 100 },
    ],
    priceRange: { min: 380, max: 520, best: 430 },
    notes: 'Maximum flexibility at lowest cost. Run VMs, containers, and Plex on the same box. Full Linux ecosystem. Requires more setup time vs Synology.',
    warnings: ['Requires manual OS setup and maintenance', 'No vendor support'],
    url: null,
    createdAt: '2026-03-02T10:05:00Z',
  },
  {
    id: 'cccccccc-cccc-4ccc-8ccc-cccccccccccc',
    missionId: '11111111-1111-4111-8111-111111111111',
    name: 'Raspberry Pi 5 — 8 GB',
    category: 'SBC',
    badge: 'watch',
    pinned: false,
    rejected: false,
    rejectReason: null,
    attributes: [
      { key: 'cpu', label: 'CPU', value: 'Broadcom BCM2712 quad-core A76 @ 2.4 GHz', type: 'text' },
      { key: 'ram', label: 'RAM', value: '8 GB LPDDR4X', type: 'text' },
      { key: 'bays', label: 'Drive bays', value: '0 (external USB / NVMe HAT required)', type: 'text' },
      { key: 'idle_power', label: 'Idle power', value: '~3–5W', type: 'text' },
      { key: 'os', label: 'OS', value: 'Raspberry Pi OS / DietPi', type: 'text' },
      { key: 'noise', label: 'Noise', value: 'Silent (no fan needed with heatsink case)', type: 'text' },
      { key: 'price_score', label: 'Value score', value: 60, type: 'score', max: 100 },
    ],
    priceRange: { min: 80, max: 160, best: 95 },
    notes: 'Ultra-low power. Best for lightweight file serving and Pihole. Not suited for heavy transcoding or large RAID arrays.',
    warnings: [
      'Limited PCIe bandwidth for multi-drive setups',
      'No ECC memory',
      'USB 3.0 bus shared with networking on older HATs',
    ],
    url: 'https://www.raspberrypi.com/products/raspberry-pi-5/',
    createdAt: '2026-03-03T09:00:00Z',
  },
  {
    id: 'dddddddd-dddd-4ddd-8ddd-dddddddddddd',
    missionId: '11111111-1111-4111-8111-111111111111',
    name: 'Dell PowerEdge T150',
    category: 'Tower server',
    badge: 'rejected',
    pinned: false,
    rejected: true,
    rejectReason: 'Too loud (45 dB+) and power-hungry (80–200W) for a home living space. Overkill for personal NAS use case.',
    attributes: [
      { key: 'cpu', label: 'CPU', value: 'Intel Xeon E-2300 series', type: 'text' },
      { key: 'ram', label: 'RAM', value: 'Up to 128 GB ECC DDR4', type: 'text' },
      { key: 'bays', label: 'Drive bays', value: '4 × 3.5" LFF', type: 'text' },
      { key: 'idle_power', label: 'Idle power', value: '~80–120W', type: 'text' },
      { key: 'noise', label: 'Noise', value: '45+ dB(A) — too loud', type: 'text' },
      { key: 'price_score', label: 'Value score', value: 35, type: 'score', max: 100 },
    ],
    priceRange: { min: 650, max: 950, best: 720 },
    notes: 'Enterprise-grade reliability but wrong tool for a quiet home server. High power cost (~€150/year extra vs N100 build).',
    warnings: ['Very loud fans', 'High electricity cost', 'No iGPU for media transcoding'],
    url: null,
    createdAt: '2026-03-03T09:10:00Z',
  },
]

// ─── Shopping items fixtures (home server custom build) ──────────────────────

export const MISSION_1_SHOPPING = [
  {
    id: 'ee000001-eeee-4eee-8eee-eeeeeeeeeeee',
    missionId: '11111111-1111-4111-8111-111111111111',
    name: 'ASRock N100DC-ITX (board + N100 combo)',
    merchant: 'LDLC',
    costCategory: 'cpu_motherboard',
    price: 149.90,
    originalEstimate: 140.0,
    targetPrice: 130.0,
    status: 'buy',
    note: 'Fanless DC-in board. Pairs with a PicoPSU for ultra-quiet build. In stock.',
    url: 'https://www.ldlc.com/fiche/PB00565862.html',
    pinned: true,
    createdAt: '2026-03-05T11:00:00Z',
  },
  {
    id: 'ee000002-eeee-4eee-8eee-eeeeeeeeeeee',
    missionId: '11111111-1111-4111-8111-111111111111',
    name: 'Kingston Fury Beast 32 GB DDR4-3200 SO-DIMM',
    merchant: 'Amazon FR',
    costCategory: 'memory',
    price: 72.99,
    originalEstimate: 80.0,
    targetPrice: 65.0,
    status: 'buy',
    note: '2×16 GB kit. N100 supports dual-channel DDR4 up to 3200 MT/s.',
    url: 'https://www.amazon.fr/dp/B09VDMG9PC',
    pinned: false,
    createdAt: '2026-03-05T11:05:00Z',
  },
  {
    id: 'ee000003-eeee-4eee-8eee-eeeeeeeeeeee',
    missionId: '11111111-1111-4111-8111-111111111111',
    name: 'Samsung 870 EVO 2 TB SATA SSD (×2)',
    merchant: 'Materiel.net',
    costCategory: 'storage',
    price: 209.98,
    originalEstimate: 240.0,
    targetPrice: 185.0,
    status: 'watch',
    note: 'Pair for ZFS mirror (RAID-1 equivalent). Price often drops during Amazon Prime Day.',
    url: 'https://www.materiel.net/produit/202104160001.html',
    pinned: false,
    createdAt: '2026-03-05T11:10:00Z',
  },
  {
    id: 'ee000004-eeee-4eee-8eee-eeeeeeeeeeee',
    missionId: '11111111-1111-4111-8111-111111111111',
    name: 'WD Red Plus 4 TB 3.5" CMR (×2)',
    merchant: 'Rueducommerce',
    costCategory: 'storage',
    price: 174.00,
    originalEstimate: 200.0,
    targetPrice: 150.0,
    status: 'watch',
    note: 'NAS-rated CMR drives for bulk cold storage. Avoid SMR WD Red (non-Plus) for ZFS.',
    url: null,
    pinned: false,
    createdAt: '2026-03-05T11:15:00Z',
  },
  {
    id: 'ee000005-eeee-4eee-8eee-eeeeeeeeeeee',
    missionId: '11111111-1111-4111-8111-111111111111',
    name: 'Fractal Design Node 304',
    merchant: 'LDLC',
    costCategory: 'case_psu',
    price: 74.90,
    originalEstimate: 75.0,
    targetPrice: 65.0,
    status: 'buy',
    note: '6 × 3.5" hot-swap bays, mini-ITX, near-silent fan. Best compact NAS case available.',
    url: 'https://www.ldlc.com/fiche/PB00142440.html',
    pinned: false,
    createdAt: '2026-03-05T11:20:00Z',
  },
  {
    id: 'ee000006-eeee-4eee-8eee-eeeeeeeeeeee',
    missionId: '11111111-1111-4111-8111-111111111111',
    name: 'Corsair SF450 SFX PSU 80+ Gold',
    merchant: 'Amazon FR',
    costCategory: 'case_psu',
    price: 89.99,
    originalEstimate: 95.0,
    targetPrice: 80.0,
    status: 'watch',
    note: 'Overkill for N100 (~70W peak) but excellent efficiency at low loads. Alternative: 120W PicoPSU.',
    url: 'https://www.amazon.fr/dp/B07VPHMS5W',
    pinned: false,
    createdAt: '2026-03-05T11:25:00Z',
  },
  {
    id: 'ee000007-eeee-4eee-8eee-eeeeeeeeeeee',
    missionId: '11111111-1111-4111-8111-111111111111',
    name: 'TP-Link TL-SG108E 8-port Managed Switch',
    merchant: 'Amazon FR',
    costCategory: 'networking',
    price: 29.99,
    originalEstimate: 35.0,
    targetPrice: 25.0,
    status: 'defer',
    note: 'Nice-to-have for VLAN segmentation. Not needed for initial setup — defer to phase 2.',
    url: 'https://www.amazon.fr/dp/B00K4DS5KU',
    pinned: false,
    createdAt: '2026-03-05T11:30:00Z',
  },
]

// ─── Other fixtures ──────────────────────────────────────────────────────────

export const SETTINGS = {
  currency: 'EUR',
  locale: 'en-US',
  llm_provider: 'ollama',
}

export const HEALTH = { status: 'ok', db: 'ok' }

export const LLM_HEALTH = {
  status: 'healthy',
  models: [],
}

export const USAGE = {
  missions_count: 2,
  options_count: 4,
  research_runs: 2,
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

  const missionOptionsResponse = {
    items: MISSION_1_OPTIONS,
    total: MISSION_1_OPTIONS.length,
    page: 1,
    limit: 20,
  }
  const missionShoppingResponse = {
    items: MISSION_1_SHOPPING,
    total: MISSION_1_SHOPPING.length,
    page: 1,
    limit: 20,
  }

  const defaults: MockRoutes = {
    '**/api/health': HEALTH,
    '**/api/health/llm': LLM_HEALTH,
    '**/api/missions': { items: [MISSION_1, MISSION_2] },
    '**/api/missions/home-server-2026': MISSION_1,
    '**/api/missions/summer-holiday-2026': MISSION_2,
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
    // Home server mission-specific routes — keyed by UUID (what the app actually fetches)
    // and by slug (for routes where slug is used, e.g. mission detail)
    [`**/api/missions/${MISSION_1.id}/options`]: missionOptionsResponse,
    [`**/api/missions/${MISSION_1.id}/shopping`]: missionShoppingResponse,
    // Slug-based fallbacks kept for any routes that still use slug
    '**/api/missions/home-server-2026/options': missionOptionsResponse,
    '**/api/missions/home-server-2026/shopping': missionShoppingResponse,
    // Research jobs — used by OptionsExplorer to show agent run status
    '**/api/missions/*/research/jobs': { jobs: [] },
    // Comparison weights / matrix — used by ComparisonMatrix on options page
    '**/api/missions/*/comparison-weights': { weights: [] },
    '**/api/missions/*/comparison-score': { missionId: MISSION_1.id, itemScores: [], topPick: '', summary: '' },
    '**/api/missions/*/matrix': { matrix: [] },
    // Duplicate detector — used by DuplicateAlert on shopping page
    '**/api/missions/*/duplicates': { missionId: '', groups: [], totalDupes: 0, generatedAt: new Date().toISOString() },
    // Sorted items — used by ShoppingList smart sort
    '**/api/missions/*/sorted-items': { items: [], sortedBy: 'price' },
    // Eco score and other panel routes
    '**/api/missions/*/eco-score': { score: 0, label: 'N/A', breakdown: [] },
    // Generic wildcard fallbacks (lower priority — specific routes above win)
    '**/api/missions/*/options': { items: [], total: 0, page: 1, limit: 20 },
    '**/api/missions/*/shopping': { items: [], total: 0, page: 1, limit: 20 },
    '**/api/missions/*/agent-runs': { items: [] },
    '**/api/missions/*/coach': { suggestions: [] },
    '**/api/missions/*/scorecard': { grade: 'B', score: 72, summary: 'Good progress', dimensions: { research: 80, pricing: 70, budgeting: 65, speed: 72 }, achievements: ['First options researched'] },
    '**/api/missions/*/french-benchmark': {
      missionId: MISSION_1.id,
      missionName: MISSION_1.name,
      medianPrice: 450,
      yourAvgPrice: 400,
      verdict: 'Bon prix',
      verdictCode: 'bon_prix',
      sampleSize: 12,
      priceRange: [300, 600],
      tips: [],
    },
    '**/api/missions/*/purchase-timeline': {
      missionId: MISSION_1.id,
      missionName: MISSION_1.name,
      buckets: [],
      summary: '',
      totalItems: 0,
    },
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
