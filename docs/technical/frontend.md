# Frontend Deep Dive

React 19 + TypeScript + Vite + Tanstack Query v5 + React Router v7.

---

## Architecture Overview

![Frontend Architecture](../assets/frontend-architecture.svg)

---

## Routing

React Router v7 with the Outlet pattern:

```
/ → Layout (sidebar + onboarding)
├── / → HQDashboard
├── /missions/:slug → MissionLayout (topnav + breadcrumb)
│   ├── / → MissionOverview
│   ├── /options → OptionsExplorer
│   └── /shopping → ShoppingTracker
├── /search → SearchPage
├── /wishlist → WishListPage
├── /history → HistoryPage
├── /stats → StatsPage
├── /settings → SettingsPage
├── /notifications → NotificationsPage
├── /performance → PerformancePage
├── /deal-calendar → DealCalendarPage
├── /cashback → CashbackPage
├── /envelopes → EnvelopesPage
├── /insights → InsightsPage
├── /digest → DigestPage
├── /loyalty → LoyaltyPage
├── /shared/:token → SharedWishlistPage
└── /join/:token → JoinPage
```

### Layout.tsx

Root shell — renders:
- `SidebarContext` provider
- Collapsible mission list sidebar
- `OnboardingOverlay` (3-step, localStorage dismissed)
- `<Outlet />` for child pages

### MissionLayout.tsx

Mission-scoped shell — renders:
- `Topnav` with mission name, breadcrumb, search, notification bell
- `LLMStatus` dot (60s poll of `/api/health/llm`)
- `<Outlet />` for mission sub-pages

---

## State Management

### Server State: Tanstack Query v5

All API data is managed by Tanstack Query:

```ts
// src/hooks/useMission.ts
export function useMission(slug: string) {
  return useQuery({
    queryKey: missionKeys.detail(slug),
    queryFn: () => api.missions.get(slug),
    staleTime: 30_000,
  })
}

export function useRunResearch(missionId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => api.research.run(missionId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: optionKeys.list(missionId) })
    },
  })
}
```

Query key conventions (in `src/api/keys.ts`):

```ts
export const missionKeys = {
  all: ['missions'] as const,
  list: () => [...missionKeys.all, 'list'] as const,
  detail: (slug: string) => [...missionKeys.all, 'detail', slug] as const,
}
```

### Local State: Contexts

| Context | Location | Purpose |
|---------|----------|---------|
| `SidebarContext` | `src/contexts/sidebar.tsx` | Sidebar open/closed |
| React Router | Built-in | URL state, navigation |

---

## API Layer (`src/api/`)

Every resource has a typed module:

```ts
// src/api/missions.ts
import { z } from 'zod'

export const MissionSchema = z.object({
  id: z.string().uuid(),
  slug: z.string(),
  name: z.string(),
  budget: z.number(),
  status: z.enum(['active', 'done', 'archived']),
  // ...
})
export type Mission = z.infer<typeof MissionSchema>

export const missions = {
  list: async (): Promise<Mission[]> => {
    const res = await fetch('/api/missions')
    return MissionSchema.array().parse(await res.json())
  },
  get: async (slug: string): Promise<Mission> => {
    const res = await fetch(`/api/missions/${slug}`)
    return MissionSchema.parse(await res.json())
  },
}
```

Zod validates **every** API response at the boundary. Schema mismatches throw immediately (visible in Tanstack Query error state).

---

## Component Architecture

### Design System (`src/components/scouter/`)

| Component | Purpose |
|-----------|---------|
| `Card` | Base card with surface background, 16px radius |
| `Badge` | Status + deal badges |
| `BudgetBar` | Budget burn progress bar |
| `StatusBadge` | Color-coded deal status (buy/watch/flash-sale…) |
| `Topnav` | Mission nav bar with LLM status dot |
| `LoadingPulse` | Spinner for async operations |
| `ScouterGrid` | Responsive grid with skeleton support |
| `Skeleton` | Card/row/chart skeleton variants |
| `EmptyState` | Icon + title + description + CTA |
| `Toast` | Notification toasts |
| `SearchDropdown` | 5-result instant search dropdown |
| `OnboardingOverlay` | 3-step onboarding modal |

### CSS Modules Pattern

Every component co-locates a `.module.css` file:

```tsx
// OptionCard.tsx
import styles from './OptionCard.module.css'

export function OptionCard({ option }: Props) {
  return (
    <div className={styles.card}>
      <span className={styles.title}>{option.name}</span>
      {/* Dynamic values via CSS custom properties */}
      <div
        className={styles.scoreBar}
        style={{ '--score': option.dealScore } as React.CSSProperties}
      />
    </div>
  )
}
```

```css
/* OptionCard.module.css */
.card {
  background: var(--surface);
  border-radius: 16px;
  padding: var(--space-4);
}
.scoreBar {
  width: calc(var(--score) * 1%);
  background: var(--color-buy);
}
```

### Theme Tokens (`src/styles/theme.css`)

All SCOUTER visual tokens live in one file. Import in any component:

```css
:root {
  --surface: #1e293b;
  --surface-raised: #263147;
  --color-buy: #22c55e;
  --color-watch: #a855f7;
  --color-flash-sale: #f97316;
  --color-crisis: #ef4444;
  --space-4: 16px;
  --radius-card: 16px;
  /* ... */
}
```

---

## Internationalization

Uses react-i18next with static JSON translation files:

```ts
// src/i18n/index.ts
i18n.use(initReactI18next).init({
  resources: { en: { translation: en }, fr: { translation: fr } },
  lng: localStorage.getItem('locale') ?? 'en',
  fallbackLng: 'en',
})
```

Usage in components:

```tsx
const { t } = useTranslation()
return <h1>{t('mission.create')}</h1>
```

Translation keys are in `src/i18n/en.json` and `src/i18n/fr.json`.

---

## Accessibility

- All interactive elements have `aria-label` or `aria-labelledby`
- Keyboard navigation: `Tab`, `Enter`, `Space` for all buttons and modals
- Custom keyboard shortcuts via `useKeyboardShortcuts` hook
- `StarRating` component is keyboard-accessible (arrow keys)
- `TemplatePreview` modal uses focus trap
- All status badges have `role="status"` or `aria-live`

---

## PWA

Configured via `vite-plugin-pwa`:

```ts
// vite.config.ts
VitePWA({
  registerType: 'autoUpdate',
  manifest: {
    name: 'SCOUTER Universal',
    short_name: 'SCOUTER',
    theme_color: '#0f172a',
    icons: [{ src: '/icon-192.png', sizes: '192x192' }],
  },
  workbox: {
    globPatterns: ['**/*.{js,css,html,svg}'],
    runtimeCaching: [{ urlPattern: /\/api\//, handler: 'NetworkFirst' }],
  },
})
```

- App shell cached for offline use
- API calls use NetworkFirst strategy (fresh data when online, cached when offline)
- Installable on desktop and mobile

---

## Semantic Search

```tsx
// SearchDropdown.tsx — in Topnav
const { data: results } = useSearch(query, { enabled: query.length >= 2 })
// 300ms debounce, 5 results in dropdown
// Enter → navigate to /search?q=<query>

// SearchPage.tsx — full results page
// URL-synced query param
// Displays all matching options with mission context
```

Backend: pgvector cosine ANN (`<=>` operator) via CTE:

```sql
WITH ranked AS (
  SELECT o.*, e.vec <=> $1::vector AS distance
  FROM options o
  JOIN embeddings e ON e.option_id = o.id
  ORDER BY distance
  LIMIT 20
)
SELECT * FROM ranked WHERE distance < 0.5
```

---

## Testing

```bash
npm run test          # Single run (Vitest)
npm run test:watch    # Watch mode
npm run test:coverage # Coverage report
npm run typecheck     # TypeScript strict check
```

Test files are co-located with components (`.test.tsx`):

```
src/
  components/
    options/
      OptionCard.tsx
      OptionCard.test.tsx   ← tests live here
  pages/
    PerformancePage.tsx
    PerformancePage.test.tsx
```

Testing stack:
- **Vitest** — fast test runner (Vite-native)
- **jsdom** — browser environment simulation
- **Testing Library** — `render`, `screen`, `fireEvent`, `userEvent`
- **MSW** (if configured) — API mocking

Coverage target: **80%+**

---

## Build

```bash
npm run build     # TypeScript check + Vite build → dist/
npm run preview   # Serve dist/ locally
```

Output: `frontend/dist/` — static files served by Nginx in Docker.

Dockerfile:
```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json .
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
```
