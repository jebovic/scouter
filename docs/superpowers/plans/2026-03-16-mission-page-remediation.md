# Mission Page Remediation Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `/missions/:slug` fully accessible by fixing a React #310 crash, treating missing sub-resource 404s as empty state, and returning a graceful degraded response instead of HTTP 502 when the LLM is unavailable.

**Architecture:** The fix has two layers — frontend (5 files) and backend (1 file). Frontend changes prevent TanStack Query's batch notifications from crashing React 19's concurrent renderer, and convert "no data yet" 404 API responses from hard errors to null data. Backend change makes the summary handler return a degraded-but-valid DTO instead of 502 when the LLM fails.

**Tech Stack:** React 19, TanStack Query v5, Vitest, Go 1.23+, chi router, net/http/httptest

**Spec:** `docs/superpowers/specs/2026-03-16-mission-page-remediation-design.md`

---

## File Map

| Action | File | Change |
|--------|------|--------|
| Modify | `frontend/src/main.tsx` | Add `notifyManager.setScheduler(queueMicrotask)` before `createRoot` |
| Modify | `frontend/src/api/forecast.ts` | `fetchForecast` returns `Promise<ForecastResult \| null>`, 404 → null |
| Create | `frontend/src/api/forecast.test.ts` | Unit tests for null-on-404 and error propagation |
| Modify | `frontend/src/api/decision.ts` | `getDecision` returns `Promise<Decision \| null>`, 404 → null |
| Create | `frontend/src/api/decision.test.ts` | Unit tests for null-on-404 and error propagation |
| Modify | `frontend/src/hooks/usePurchase.ts` | Add `retry: false` to `usePurchaseRecord` |
| Modify | `frontend/src/hooks/useSummary.ts` | Change `retry: 1` → `retry: false` on `useMissionSummary` |
| Modify | `backend/internal/summary/handler_test.go` | Replace 502 test; add degraded DTO test; add not-cached test |
| Modify | `backend/internal/summary/handler.go` | Return degraded 200 DTO instead of 502 on LLM failure |

---

## Chunk 1: Frontend notifyManager fix

### Task 1: Add notifyManager scheduler to main.tsx

**Files:**
- Modify: `frontend/src/main.tsx`

Current content of `main.tsx` (lines 1-27):
```ts
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from 'react-router-dom'
import { router } from './router'
import { ToastProvider } from './components/scouter'
import './i18n'
import './styles/global.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <RouterProvider router={router} />
      </ToastProvider>
    </QueryClientProvider>
  </StrictMode>,
)
```

- [ ] **Step 1: Apply the notifyManager fix**

Replace the content of `frontend/src/main.tsx` with:

```ts
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider, notifyManager } from '@tanstack/react-query'
import { RouterProvider } from 'react-router-dom'
import { router } from './router'
import { ToastProvider } from './components/scouter'
import './i18n'
import './styles/global.css'

// Defer TanStack Query batch notifications to the microtask queue.
// Without this, React 19's concurrent renderer throws error #310 when
// multiple queries error simultaneously (e.g. 3x 404 at page mount).
notifyManager.setScheduler(queueMicrotask)

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <RouterProvider router={router} />
      </ToastProvider>
    </QueryClientProvider>
  </StrictMode>,
)
```

- [ ] **Step 2: Verify TypeScript compiles**

Run from `frontend/`:
```bash
npm run typecheck
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/main.tsx
git commit -m "fix(react): defer TanStack Query notifyManager to queueMicrotask for React 19"
```

---

## Chunk 2: null-on-404 for fetchForecast and getDecision

### Task 2: null-on-404 for fetchForecast

**Files:**
- Modify: `frontend/src/api/forecast.ts`
- Create: `frontend/src/api/forecast.test.ts`

- [ ] **Step 1: Write the failing tests**

Create `frontend/src/api/forecast.test.ts`:

```ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ApiError } from './client'
import { fetchForecast } from './forecast'

vi.mock('./client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('./client')>()
  return { ...actual, apiFetch: vi.fn() }
})

import { apiFetch } from './client'
const mockApiFetch = vi.mocked(apiFetch)

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('fetchForecast', () => {
  it('returns null when the API responds 404 (no forecast yet)', async () => {
    mockApiFetch.mockRejectedValueOnce(new ApiError(404, 'not found'))
    const result = await fetchForecast('some-mission-id')
    expect(result).toBeNull()
  })

  it('throws when the API responds 500 (server error should propagate)', async () => {
    mockApiFetch.mockRejectedValueOnce(new ApiError(500, 'internal error'))
    await expect(fetchForecast('some-mission-id')).rejects.toBeInstanceOf(ApiError)
  })

  it('returns parsed DTO on success', async () => {
    const validForecast = {
      id: 'f1',
      missionId: 'm1',
      estimatedTotal: 299.99,
      confidence: 'high',
      recommendations: ['Buy now'],
      riskFactors: [],
      monthsToSave: null,
      optimalBuyTime: null,
      createdAt: '2026-03-16T00:00:00Z',
    }
    mockApiFetch.mockResolvedValueOnce(validForecast)
    const result = await fetchForecast('m1')
    expect(result).not.toBeNull()
    expect(result!.confidence).toBe('high')
  })
})
```

- [ ] **Step 2: Run tests — expect FAIL**

```bash
cd frontend && npm test -- src/api/forecast.test.ts
```
Expected: `fetchForecast` returns 404 → test fails because current implementation throws instead of returning null.

- [ ] **Step 3: Update fetchForecast to handle 404 → null**

In `frontend/src/api/forecast.ts`, replace:
```ts
export async function fetchForecast(missionId: string): Promise<ForecastResult> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/forecast`)
  return ForecastResultSchema.parse(data)
}
```

With:
```ts
export async function fetchForecast(missionId: string): Promise<ForecastResult | null> {
  try {
    const data = await apiFetch<unknown>(`/api/missions/${missionId}/forecast`)
    return ForecastResultSchema.parse(data)
  } catch (e: unknown) {
    if (e instanceof Error && 'status' in e && (e as { status: number }).status === 404) return null
    throw e
  }
}
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
cd frontend && npm test -- src/api/forecast.test.ts
```
Expected: all 3 tests pass.

- [ ] **Step 5: Typecheck**

```bash
npm run typecheck
```
Expected: no errors. (The return type change from `ForecastResult` to `ForecastResult | null` propagates to `useForecast` — callers use optional chaining so this is safe.)

- [ ] **Step 6: Commit**

```bash
git add frontend/src/api/forecast.ts frontend/src/api/forecast.test.ts
git commit -m "fix(api): fetchForecast returns null on 404 instead of throwing"
```

### Task 3: null-on-404 for getDecision

**Files:**
- Modify: `frontend/src/api/decision.ts`
- Create: `frontend/src/api/decision.test.ts`

- [ ] **Step 1: Write the failing tests**

Create `frontend/src/api/decision.test.ts`:

```ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ApiError } from './client'
import { getDecision } from './decision'

vi.mock('./client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('./client')>()
  return { ...actual, apiFetch: vi.fn() }
})

import { apiFetch } from './client'
const mockApiFetch = vi.mocked(apiFetch)

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('getDecision', () => {
  it('returns null when the API responds 404 (no decision yet)', async () => {
    mockApiFetch.mockRejectedValueOnce(new ApiError(404, 'not found'))
    const result = await getDecision('some-mission-id')
    expect(result).toBeNull()
  })

  it('throws when the API responds 500 (server error should propagate)', async () => {
    mockApiFetch.mockRejectedValueOnce(new ApiError(500, 'internal error'))
    await expect(getDecision('some-mission-id')).rejects.toBeInstanceOf(ApiError)
  })

  it('returns parsed DTO on success', async () => {
    const validDecision = {
      id: '00000000-0000-0000-0000-000000000001',
      missionId: '00000000-0000-0000-0000-000000000002',
      scores: [],
      summary: 'Aucune option',
      createdAt: '2026-03-16T00:00:00Z',
    }
    mockApiFetch.mockResolvedValueOnce(validDecision)
    const result = await getDecision('some-mission-id')
    expect(result).not.toBeNull()
    expect(result!.summary).toBe('Aucune option')
  })
})
```

- [ ] **Step 2: Run tests — expect FAIL**

```bash
cd frontend && npm test -- src/api/decision.test.ts
```
Expected: 404 test fails because current `getDecision` throws instead of returning null.

- [ ] **Step 3: Update getDecision to handle 404 → null**

In `frontend/src/api/decision.ts`, replace:
```ts
export async function getDecision(missionId: string): Promise<Decision> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/decision`)
  return DecisionSchema.parse(data) as Decision
}
```

With:
```ts
export async function getDecision(missionId: string): Promise<Decision | null> {
  try {
    const data = await apiFetch<unknown>(`/api/missions/${missionId}/decision`)
    return DecisionSchema.parse(data) as Decision
  } catch (e: unknown) {
    if (e instanceof Error && 'status' in e && (e as { status: number }).status === 404) return null
    throw e
  }
}
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
cd frontend && npm test -- src/api/decision.test.ts
```
Expected: all 3 tests pass.

- [ ] **Step 5: Typecheck**

```bash
npm run typecheck
```
Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/api/decision.ts frontend/src/api/decision.test.ts
git commit -m "fix(api): getDecision returns null on 404 instead of throwing"
```

---

## Chunk 3: Hook retry fixes

### Task 4: retry: false on usePurchaseRecord and useMissionSummary

**Files:**
- Modify: `frontend/src/hooks/usePurchase.ts`
- Modify: `frontend/src/hooks/useSummary.ts`
- Create: `frontend/src/hooks/usePurchaseRecord.test.ts`

- [ ] **Step 1: Write failing test for retry behavior**

Create `frontend/src/hooks/usePurchaseRecord.test.ts`:

```ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React from 'react'
import { usePurchaseRecord } from './usePurchase'

vi.mock('../api/purchase', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/purchase')>()
  return { ...actual, getPurchaseRecord: vi.fn() }
})

import { getPurchaseRecord } from '../api/purchase'
const mockGet = vi.mocked(getPurchaseRecord)

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return React.createElement(QueryClientProvider, { client: qc }, children)
}

beforeEach(() => {
  mockGet.mockReset()
})

describe('usePurchaseRecord', () => {
  it('returns null data when getPurchaseRecord resolves null (no record yet)', async () => {
    mockGet.mockResolvedValue(null)
    const { result } = renderHook(() => usePurchaseRecord('mission-123'), { wrapper })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toBeNull()
    expect(mockGet).toHaveBeenCalledTimes(1)
  })

  it('calls getPurchaseRecord exactly once on failure (retry: false)', async () => {
    const { ApiError } = await import('../api/client')
    mockGet.mockRejectedValue(new ApiError(500, 'server error'))
    const { result } = renderHook(() => usePurchaseRecord('mission-456'), { wrapper })
    await waitFor(() => expect(result.current.isError).toBe(true))
    // With retry: false, the mock should be called exactly once, not retried.
    expect(mockGet).toHaveBeenCalledTimes(1)
  })
})
```

- [ ] **Step 2: Run tests — expect FAIL on retry count assertion**

```bash
cd frontend && npm test -- src/hooks/usePurchaseRecord.test.ts
```
Expected: second test fails because `usePurchaseRecord` currently has no `retry` option, inheriting QueryClient default of `retry: 1` → mock called twice.

(Note: the wrapper overrides retry to `false` at the QueryClient level, so the test is relying on the hook's own `retry` option. Without `retry: false` in the hook, the hook inherits `retry: 1`. We need to set `retry` in the hook itself, not rely solely on the QueryClient wrapper.)

Actually, to properly test the hook's own `retry: false` option, we give the wrapper QueryClient a `retry: 2` default (not 0), so the hook's `retry: false` will override it:

Update the wrapper in the test to use `retry: 2` as default:

```ts
function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: 2 } } })
  return React.createElement(QueryClientProvider, { client: qc }, children)
}
```

Re-run:
```bash
cd frontend && npm test -- src/hooks/usePurchaseRecord.test.ts
```
Expected: second test fails — mock called 3 times (1 + 2 retries from QueryClient default), not 1.

- [ ] **Step 3: Add retry: false to usePurchaseRecord**

In `frontend/src/hooks/usePurchase.ts`, replace:
```ts
export function usePurchaseRecord(missionId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.purchase(missionId ?? ''),
    queryFn: () => getPurchaseRecord(missionId!),
    enabled: !!missionId,
  })
}
```

With:
```ts
export function usePurchaseRecord(missionId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.purchase(missionId ?? ''),
    queryFn: () => getPurchaseRecord(missionId!),
    enabled: !!missionId,
    retry: false,
  })
}
```

- [ ] **Step 4: Change retry: 1 to retry: false on useMissionSummary**

In `frontend/src/hooks/useSummary.ts`, replace:
```ts
    retry: 1,
```

With:
```ts
    retry: false,
```

- [ ] **Step 5: Run tests — expect PASS**

```bash
cd frontend && npm test -- src/hooks/usePurchaseRecord.test.ts
```
Expected: both tests pass. Mock called once in both cases.

- [ ] **Step 6: Run full frontend test suite**

```bash
cd frontend && npm test
```
Expected: all existing tests continue to pass.

- [ ] **Step 7: Typecheck**

```bash
npm run typecheck
```
Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add frontend/src/hooks/usePurchase.ts frontend/src/hooks/useSummary.ts frontend/src/hooks/usePurchaseRecord.test.ts
git commit -m "fix(hooks): retry: false on usePurchaseRecord and useMissionSummary"
```

---

## Chunk 4: Backend graceful LLM degradation

### Task 5: Degraded DTO on LLM failure in summary handler

**Files:**
- Modify: `backend/internal/summary/handler_test.go`
- Modify: `backend/internal/summary/handler.go`

**Background:** The existing test `TestHandler_Get_BrieferError_Returns502` (line 261) asserts that a briefer error returns HTTP 502. After this change, it should return HTTP 200 with a degraded DTO. We must update that test first (it becomes the new specification), then update the handler.

- [ ] **Step 1: Update the existing 502 test and add the not-cached test**

In `backend/internal/summary/handler_test.go`, replace the block starting at line 261:
```go
func TestHandler_Get_BrieferError_Returns502(t *testing.T) {
	br := &fakeBriefer{err: errors.New("llm unavailable")}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(time.Hour)
	h := buildHandler(mget, olist, slist, br, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/summary", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", w.Code, w.Body.String())
	}
}
```

With:
```go
func TestHandler_Get_BrieferError_ReturnsDegradedDTO(t *testing.T) {
	br := &fakeBriefer{err: errors.New("llm unavailable")}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(time.Hour)
	h := buildHandler(mget, olist, slist, br, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/summary", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 degraded DTO, got %d: %s", w.Code, w.Body.String())
	}
	var got summary.MissionSummary
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode degraded DTO: %v", err)
	}
	if got.Verdict != "research_more" {
		t.Errorf("expected verdict 'research_more', got %q", got.Verdict)
	}
	if len(got.Bullets) == 0 {
		t.Error("expected at least one bullet in degraded DTO")
	}
	if got.CachedAt <= 0 {
		t.Error("expected cachedAt > 0 in degraded DTO")
	}
	if got.MissionSlug != testSlug {
		t.Errorf("expected missionSlug %q, got %q", testSlug, got.MissionSlug)
	}
}

func TestHandler_Get_BrieferError_DegradedResponseNotCached(t *testing.T) {
	br := &fakeBriefer{err: errors.New("llm unavailable")}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(time.Hour)
	h := buildHandler(mget, olist, slist, br, cache)
	router := mountRouter(h)

	// First GET: briefer fails, returns degraded DTO.
	r1 := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/summary", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first GET: expected 200, got %d", w1.Code)
	}

	// Second GET: briefer should be called again (degraded response was not cached).
	r2 := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/summary", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second GET: expected 200, got %d", w2.Code)
	}

	// Both responses must be valid degraded DTOs.
	var got1, got2 summary.MissionSummary
	_ = json.NewDecoder(w1.Body).Decode(&got1)
	_ = json.NewDecoder(w2.Body).Decode(&got2)
	if got1.Verdict != "research_more" || got2.Verdict != "research_more" {
		t.Errorf("expected both responses to be degraded: got %q, %q", got1.Verdict, got2.Verdict)
	}

	// Briefer must have been called twice — degraded DTO was NOT served from cache.
	if br.calls != 2 {
		t.Errorf("expected briefer called 2 times (not cached), got %d", br.calls)
	}
}
```

- [ ] **Step 2: Run Go tests — expect FAIL**

```bash
cd backend && go test ./internal/summary/... -v -run "TestHandler_Get_BrieferError"
```
Expected: `TestHandler_Get_BrieferError_ReturnsDegradedDTO` FAILS (handler currently returns 502), `TestHandler_Get_BrieferError_DegradedResponseNotCached` also FAILS.

- [ ] **Step 3: Update handler.go to return degraded DTO**

In `backend/internal/summary/handler.go`, replace:
```go
	// Generate brief via LLM agent.
	brief, err := h.briefer.Brief(r.Context(), *m, opts, items)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "summary service unavailable")
		return
	}

	h.cache.Set(slug, brief)
	httputil.WriteJSON(w, http.StatusOK, brief)
```

With:
```go
	// Generate brief via LLM agent.
	brief, err := h.briefer.Brief(r.Context(), *m, opts, items)
	if err != nil {
		// LLM unavailable: return a degraded-but-valid DTO so the frontend
		// can render a useful state instead of an error. Do NOT cache this
		// response — the next request will retry the LLM.
		degraded := &MissionSummary{
			MissionSlug: slug,
			Bullets:     []string{"Service LLM temporairement indisponible"},
			Verdict:     "research_more",
			CachedAt:    time.Now().Unix(),
		}
		httputil.WriteJSON(w, http.StatusOK, degraded)
		return
	}

	h.cache.Set(slug, brief)
	httputil.WriteJSON(w, http.StatusOK, brief)
```

Also add `"time"` to the import block if not already present. The full import block should be:

```go
import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
)
```

Note: `"context"` was in the original import but is not used directly in `handler.go` — it was used through the interface methods. Remove it if `go vet` flags it; keep it if not.

- [ ] **Step 4: Run Go tests — expect PASS**

```bash
cd backend && go test ./internal/summary/... -v
```
Expected: all tests pass including the two new ones.

- [ ] **Step 5: Run full backend test suite**

```bash
cd backend && go test ./... 2>&1 | tail -20
```
Expected: no failures.

- [ ] **Step 6: Run go vet**

```bash
cd backend && go vet ./internal/summary/...
```
Expected: no issues.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/summary/handler.go backend/internal/summary/handler_test.go
git commit -m "fix(summary): return degraded 200 DTO instead of 502 when LLM unavailable"
```

---

## Chunk 5: Frontend build verification

### Task 6: Confirm full build passes

- [ ] **Step 1: Run full Vitest suite**

```bash
cd frontend && npm test
```
Expected: all tests pass (71+ existing + 6 new tests in forecast.test.ts, decision.test.ts, usePurchaseRecord.test.ts).

- [ ] **Step 2: Run Vite build**

```bash
cd frontend && npm run build
```
Expected: build succeeds with no TypeScript errors.

- [ ] **Step 3: Commit build verification (if any incidental fixes were needed)**

Only commit if step 1 or 2 revealed and required fixing something. Otherwise skip.

---

## Chunk 6: Smoke test in browser

### Task 7: Playwright smoke verification

These steps use the running dev environment (`https://scouter.dev.local`). Run after deploying the rebuilt containers (`make up` or `docker compose -f deployment/docker-compose.yml up -d --build`).

- [ ] **Step 1: Rebuild and deploy containers**

```bash
cd /home/jibei/projects/scouter
docker compose -f deployment/docker-compose.yml up -d --build backend frontend
```
Expected: containers restart with new code. Check logs for startup errors:
```bash
docker compose -f deployment/docker-compose.yml logs --tail=20 backend frontend
```

- [ ] **Step 2: Navigate to mission page and check for #310**

Use Playwright or open browser DevTools at `https://scouter.dev.local/missions/home-server-2026`.

Assert in console:
- **No** `Minified React error #310` error
- **No** `[ErrorBoundary] Error` message
- `Failed to load resource: 404` for forecast/decision/purchase is OK (expected — means no data yet)

Assert in DOM:
- Mission name/title is visible on the page
- Page is NOT showing an error boundary fallback

- [ ] **Step 3: Verify empty state panels**

On the mission page:
- Forecast panel: renders in "no forecast yet" state (not error state)
- Decision panel: renders in "no decision yet" state (not error state)
- Purchase panel: renders (no record) without crashing

- [ ] **Step 4: Click "Brief IA" and verify degraded state**

Click the "Brief IA" button in the `MissionSummaryCard` component.

Assert:
- No 502 response in the Network tab
- Response is HTTP 200
- The card shows the degraded state: `research_more` verdict and the bullet "Service LLM temporairement indisponible"

- [ ] **Step 5: Final commit if any adjustments were needed**

Only needed if the smoke test revealed a small issue that was fixed inline. Otherwise all commits are already done in Tasks 1-5.

---

## Summary of commits expected

1. `fix(react): defer TanStack Query notifyManager to queueMicrotask for React 19`
2. `fix(api): fetchForecast returns null on 404 instead of throwing`
3. `fix(api): getDecision returns null on 404 instead of throwing`
4. `fix(hooks): retry: false on usePurchaseRecord and useMissionSummary`
5. `fix(summary): return degraded 200 DTO instead of 502 when LLM unavailable`
