# Mission Page Remediation — Design Spec

**Date:** 2026-03-16
**Status:** Approved
**Approach:** C — Full resilience (notifyManager + null-on-404 + graceful LLM degradation)

---

## Problem Summary

The mission page (`/missions/:slug`) is inaccessible. A Playwright debug session and container log capture identified four distinct root causes:

1. **React error #310** — "Cannot update a component while rendering a different component". Three simultaneous 404 API responses (forecast + decision + purchase at ~502ms) trigger TanStack Query's `notifyManager` batch notification system. In React 19 concurrent mode, these synchronous batch updates conflict with the renderer and throw #310 as a hard error (was a warning in React 18). The `ErrorBoundary` wrapping `MissionOverview` catches it, replacing the page with an error UI.

2. **`fetchForecast` and `getDecision` treat 404 as hard error** — Both call `apiFetch` which throws `ApiError` on any non-2xx response. These 404s are legitimate "no data yet" states (no forecast/decision created for the mission), not routing errors. TanStack Query puts the query into `error` state, amplifying the #310 cascade. (`getPurchaseRecord` already handles 404 → null correctly.)

3. **`usePurchaseRecord` retries on 404** — Missing `retry: false` causes TanStack Query to retry the 404 once (default `retry: 1` from QueryClient config), adding an unnecessary second failing request. (`useForecast` and `useDecision` already have `retry: false`.)

4. **502 on "Brief IA"** — Ollama is not running on the host. SmartRouter cascades (qwen3:14b → qwen3:4b → deepseek-v3.2:cloud, ~26s). When the cloud fallback also fails, the backend returns HTTP 502. `MissionSummaryCard` has no graceful empty state for 502. `useMissionSummary` has `retry: 1` which would double the latency (up to ~52s) before surfacing an error.

---

## Design

### Frontend Changes (4 files)

#### 1. `frontend/src/main.tsx` — notifyManager scheduler fix

Add the following immediately before `createRoot`. This defers TanStack Query's batch notifications to the microtask queue, giving React 19's concurrent renderer a clean commit phase before state updates fire.

```ts
import { notifyManager } from '@tanstack/react-query'
notifyManager.setScheduler(queueMicrotask)
```

This is the documented TanStack Query fix for React 18+ concurrent rendering.

#### 2. `frontend/src/api/forecast.ts` and `frontend/src/api/decision.ts` — null-on-404

Each query function wraps `apiFetch` in a `try/catch` that catches `ApiError` with `status === 404`, returning `null` instead of re-throwing. This converts "no data yet" into an empty state that TanStack Query stores as `data: null, status: 'success'` — no error path, no cascade.

Pattern (matches existing `getPurchaseRecord` in `api/purchase.ts`):
```ts
} catch (e: unknown) {
  if (e instanceof Error && 'status' in e && (e as { status: number }).status === 404) return null
  throw e
}
```

Return types change from `Promise<ForecastDTO>` / `Promise<DecisionDTO>` to `Promise<ForecastDTO | null>` / `Promise<DecisionDTO | null>`. All callers already handle `null` data via conditional rendering — no cascading UI changes needed.

No changes to `api/purchase.ts` — null-on-404 is already implemented there.

#### 3. `frontend/src/hooks/usePurchase.ts` — retry: false on usePurchaseRecord

Add `retry: false` to the `usePurchaseRecord` query options. This is the only mission sub-resource hook missing it.

#### 4. `frontend/src/hooks/useSummary.ts` — retry: false on useMissionSummary

Change `retry: 1` to `retry: false` on `useMissionSummary`. With the backend now returning 200 on LLM failure, retrying is unnecessary. Without this fix, if the backend degradation is ever reverted, the hook would make two ~26s LLM calls before surfacing an error.

### Backend Change (1 file)

#### `backend/internal/summary/handler.go` — graceful LLM degradation

When `briefer.Brief()` returns an error, return HTTP 200 with a degraded `MissionSummaryDTO` instead of HTTP 502:

```json
{
  "missionSlug": "<slug>",
  "bullets": ["Service LLM temporairement indisponible"],
  "verdict": "research_more",
  "cachedAt": <time.Now().Unix()>
}
```

**i18n note:** The bullet string is in French to match the project's primary display language. The `MissionSummaryDTO.bullets` field is a plain string array (not a translation key), so the string must be localized at source. French is used because `CLAUDE.md` specifies the app is multilingual (EN/FR) but the backend has no access to the user's locale preference. Using French here is consistent with the existing hardcoded French strings in `MissionSummaryCard.tsx`.

The degraded response is **NOT written to the in-memory cache** — the next user click retries the LLM.

`cachedAt` is set to `time.Now().Unix()` (not `0` or a sentinel). `MissionSummaryCard` renders `relativeTime(brief.cachedAt)` as a human-readable timestamp — using the current time means it will display "à l'instant" (just now), which is correct and unambiguous for a freshly-generated degraded response.

The frontend `MissionSummaryCard` already renders all three verdict states; `research_more` with a single informational bullet is a clean degraded state that passes `MissionSummarySchema` Zod validation (`bullets: z.array(z.string()).min(1)`, `verdict: z.enum([...])`, `cachedAt: z.number().int()`).

---

## Error Handling

| Scenario | Before | After |
|---|---|---|
| No forecast data | `ApiError(404)` → TanStack error → React #310 crash | `null` → `data: null, status: 'success'` → empty state UI |
| No decision data | `ApiError(404)` → TanStack error → React #310 crash | `null` → empty state UI |
| No purchase data | Already returns `null` (no change) | No change |
| Simultaneous forecast+decision 404s | React #310 → ErrorBoundary → blank page | No errors → all panels render in empty state |
| `usePurchaseRecord` 404 | Retried once → 2 failing requests | `retry: false` → 1 request → empty state |
| LLM unavailable on "Brief IA" | HTTP 502 → Zod parse error → "Réessayer" shown | HTTP 200 degraded DTO → `research_more` + French bullet |
| LLM slow but successful (~26s) | Success after 26s, `retry: 1` would double on failure | `retry: false` → surfaces immediately if fails |

---

## Testing

### Frontend (Vitest)

**`api/forecast.ts`**
- `ApiError` with `status: 404` → resolves to `null`
- `ApiError` with `status: 500` → still throws
- Successful response → returns parsed DTO

**`api/decision.ts`**
- Same pattern as forecast

**`hooks/usePurchase.ts`** (integration behavior check)
- Mock `getPurchaseRecord` returning `null` → `usePurchaseRecord` data is `null`, status is `'success'`
- Mock `getPurchaseRecord` throwing `ApiError(404)` → assert the mock is called exactly once (no retry). This validates the `retry: false` addition.

### Backend (Go)

**`internal/summary/handler_test.go`** — add two test cases:

1. **Degraded response shape**: `fakeBriefer.err = errors.New("llm unavailable")` → GET → assert HTTP 200, response body matches degraded DTO schema (`verdict: "research_more"`, `bullets` non-empty, `cachedAt` > 0).

2. **Not cached**: Make two sequential GET requests with a failing briefer. Assert `fakeBriefer.calls == 2` (briefer called on both requests) AND assert both responses are HTTP 200 with valid degraded DTO shape. This proves degraded responses are never written to the cache — if they were cached, the second call would return from cache and `calls` would be 1.

### E2E (Playwright)

- Navigate to `https://scouter.dev.local/missions/home-server-2026`
- Assert no browser console errors containing `#310`
- Assert MissionOverview main content renders (mission title visible in DOM)
- Assert forecast panel renders in empty/loading state (not error boundary)
- Assert decision panel renders in empty/loading state (not error boundary)
- Assert purchase panel renders (data: null → empty state)
- Click "Brief IA" button → assert degraded state message visible ("Service LLM temporairement indisponible" text or `research_more` verdict badge)

---

## Out of Scope

- Starting Ollama locally (operational, not a code fix)
- Adding a Traefik timeout for LLM requests (separate infrastructure task)
- Fixing remaining hardcoded French strings in `MissionSummaryCard.tsx` (i18n debt, separate task)
- The `routes.go` persona fix commit (separate cleanup, already done)
