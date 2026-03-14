# SCOUTER Frontend

React 19 + TypeScript + Vite frontend for SCOUTER Universal — personal spending intelligence.

## Dev setup

```bash
npm install
npm run dev      # http://localhost:5173
npm run build
npm run lint
```

Requires the backend running at `http://localhost:8080` (or set `VITE_API_BASE` in `.env.local`).

## Structure

```
src/
  api/            typed fetch + Zod schemas (missions, options, shopping)
  components/
    scouter/      design system: Card, Badge, BudgetBar, StatusBadge, Topnav, LoadingPulse, ScouterGrid
    mission/      MissionCard, MissionForm, ConstraintEditor, CategoryTemplate
    options/      OptionCard, AttributeRenderer, ComparisonTable, ConstraintChecker, RadarChart
    shopping/     ShoppingList, MerchantGroup, ShoppingItemRow, PriceHistoryChart, CostBreakdown
  pages/          HQDashboard, MissionOverview, OptionsExplorer, ShoppingTracker
  hooks/          useMission, useOptions, useShopping, useResearch, usePriceIntel
  styles/
    theme.css     CSS custom properties (all SCOUTER tokens)
    global.css    reset + base typography
  i18n/           index.ts, en.json, fr.json (i18next)
  types/          TypeScript types mirrored from Zod schemas
  main.tsx
```

## Design system

All tokens are in `src/styles/theme.css`. Import it once at the root; components use `var(--*)`.

Key tokens:
- `--surface` — card background
- `--border-radius` — 16px, used on all cards
- `--color-accent` — primary action color

Status badge variants (pass as `variant` prop):
`buy` · `flash-sale` · `preorder` · `defer` · `watch` · `crisis` · `recommended` · `rejected`

## i18n

```ts
import { useTranslation } from 'react-i18next'
const { t } = useTranslation()
t('missions.create')
```

Locale files: `src/i18n/en.json`, `src/i18n/fr.json`. Default locale: `en`.

## API layer

All API calls go through `src/api/`. Each module exports typed query functions and Zod schemas.
Zod validates every response at the network boundary — callers get typed data or a thrown `ZodError`.

```ts
import { getMission, listMissions } from '@/api/missions'
```

## State management

Tanstack Query v5 for server state. Custom hooks in `src/hooks/` wrap `useQuery`/`useMutation`.

```ts
const { mission, isLoading } = useMission(slug)
const { options } = useOptions(missionId)
const { triggerResearch, isPending } = useResearch(missionId)
```
