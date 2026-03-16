# Debug Session — 2026-03-16

## Executive Summary

- **Total issues:** 8 (1 critical, 4 high, 1 medium, 2 low)
- **Pages with errors:** `/missions/:slug` (MissionOverview — fatal crash)
- **Backend errors observed:** Ollama DNS failure on every LLM call; 4 missing API endpoints (404)
- **Phase 2 flows reproduced:** 0 (user chose direct remediation after Phase 1)
- **Status:** CRITICAL bug fixed in this session

---

## Findings

### [CRITICAL] MissionOverview Crashes on Page Load — `null.length` in CashbackSummaryPanel

- **Route:** `/missions/:slug`
- **Symptom:** `TypeError: Cannot read properties of null (reading 'length')` at `hu (index-_6c0Bw91.js:11:127247)` called from `q (MissionOverview-DuAyUi-5.js:1:11379)`. Page renders "[ SYSTEM ERROR ] SOMETHING WENT WRONG" via `ErrorBoundary`.
- **Triggered by:** Page load (Phase 1 automated sweep)
- **Root cause:** Two-layer bug:
  1. **Backend** (`internal/cashbacktracker/handler.go:125`): `var cashbacks []MerchantCashback` declares a nil slice. When the mission has no shopping items, the for loop is never entered and `cashbacks` remains nil. Go serializes nil slices as JSON `null`, so the response is `{"merchantCashback": null, ...}`.
  2. **Frontend** (`src/api/cashbacktracker.ts:23`): `getCashbackSummary` called `apiFetch<CashbackSummary>(...)` without Zod validation, bypassing schema enforcement. `data.merchantCashback` is therefore `null` at runtime despite the TypeScript type saying `MerchantCashback[]`.
  3. **Component** (`src/components/mission/CashbackSummaryPanel.tsx:78`): `data.merchantCashback.length > 0` — direct `.length` access on null crashes.
- **Fix applied:**
  - `backend/internal/cashbacktracker/handler.go:125`: `var cashbacks []MerchantCashback` → `cashbacks := make([]MerchantCashback, 0)` — empty slice serializes to `[]` not `null`.
  - `frontend/src/api/cashbacktracker.ts:14`: `merchantCashback: z.array(...).nullable().default([])` — Zod now coerces null → `[]`.
  - `frontend/src/api/cashbacktracker.ts:23-25`: Added `CashbackSummarySchema.parse(data)` call.

---

### [HIGH] 4 Backend API Endpoints Return 404

- **Routes:** All mission pages calling these endpoints
- **Symptom:** HTTP 404 on every call to `/api/missions/:id/decision`, `/api/missions/:id/forecast`, `/api/missions/:id/purchase`, `/api/missions/:id/summary`
- **Triggered by:** Page load (Phase 1 sweep, all mission routes)
- **Backend log:** No log entries — requests never reached backend handlers (routes not registered or handlers not wired)
- **Root cause hypothesis:** These endpoints were implemented in later phases but may not be registered in `cmd/server/routes.go`, or the handler files exist but are not imported. Check `registerRoutes` in `backend/cmd/server/routes.go`.
- **Fix:** Audit `backend/cmd/server/routes.go` for missing route registrations for `decisionHandler`, `forecastHandler`, `purchaseHandler`, and `summaryHandler`. Components have null guards and degrade gracefully — these are non-blocking UX issues.

---

### [HIGH] Ollama DNS Failure — LLM Agent Calls All Fail

- **Route:** Any page that triggers research or pricing agent
- **Symptom:** `dial tcp: lookup host.docker.internal on 127.0.0.11:53: no such host` → SmartRouter cascades to `deepseek-v3.2:cloud` → `context canceled`. Research and pricing agents fail silently.
- **Triggered by:** Any `/api/missions/:id/research` or `/api/missions/:id/pricing` call
- **Backend log:** `{"level":"ERROR","msg":"ollama call failed","error":"dial tcp: lookup host.docker.internal..."}` followed by `{"level":"WARN","msg":"cascading to cloud","provider":"deepseek-v3.2:cloud"}` then `{"level":"ERROR","msg":"cloud call failed","error":"context canceled"}`
- **Root cause hypothesis:** `host.docker.internal` DNS is a Docker Desktop–specific hostname not available in all Linux Docker environments. On WSL2/Linux the hostname resolves on the host but not inside the container without explicit Docker network configuration.
- **Fix:** In `deployment/docker-compose.yml`, add `extra_hosts: ["host.docker.internal:host-gateway"]` to the backend service, or set `OLLAMA_BASE_URL=http://172.17.0.1:11434` (Docker bridge IP) as a fallback. Alternatively, run Ollama as a Docker service in the same network.

---

### [MEDIUM] ServiceWorker SSL Error on Every Page Load

- **Route:** All pages
- **Symptom:** `Failed to register a ServiceWorker for scope 'https://scouter.dev.local/' with script 'https://scouter.dev.local/sw.js': An SSL certificate error occurred when fetching the script.`
- **Triggered by:** Page load (Phase 1 sweep, all routes)
- **Root cause hypothesis:** ServiceWorker registration uses the browser's network stack but does NOT inherit the OS CA trust store. The self-signed Traefik dev cert (`make certs`) is trusted by Chromium's TLS layer but rejected by the ServiceWorker registration API in some configurations.
- **Fix:** Either disable the ServiceWorker in dev (`if (import.meta.env.DEV) { /* skip registration */ }`), or use `localhost` as the dev domain for PWA testing (avoids the self-signed cert issue entirely). The Playwright `--ignore-certificate-errors` flag would also suppress this but is not used per spec.

---

### [LOW] Stats Page — Raw Category Key Names Displayed

- **Route:** `/stats`
- **Symptom:** Category labels show raw backend key names (e.g., `elektrische_zahnbuerste`, `summer_holiday_2026`) instead of human-readable display names.
- **Triggered by:** Page load (Phase 1 sweep, `/stats`)
- **Root cause hypothesis:** The stats endpoint returns raw `costCategory` enum values from the DB. The frontend renders them directly without a display-name mapping.
- **Fix:** Add a `CATEGORY_DISPLAY_NAMES` map in `frontend/src/pages/StatsPage.tsx` (or reuse the existing one from `CategoryTemplate`) and apply it when rendering category labels.

---

### [LOW] i18n — Hardcoded French Strings in Several Components

- **Routes:** `/missions/:slug`, `/missions/:slug/shopping`
- **Symptom:** Text like "Rythme de dépenses", "Impossible de charger le rythme de dépenses", "Meilleure plateforme:", "Voir tout (N)", "Voir moins" are hardcoded French strings not going through `t()`.
- **Triggered by:** Page load
- **Affected files:**
  - `src/components/mission/BurnRateCard.tsx` — several hardcoded French labels
  - `src/components/mission/CashbackSummaryPanel.tsx` — "Meilleure plateforme:", "Cashback Estimé"
  - `src/components/mission/MissionTimeline.tsx` — "Voir tout (N)", "Voir moins"
- **Fix:** Replace hardcoded strings with `t()` calls and add corresponding keys to `src/i18n/en.json` and `src/i18n/fr.json`.

---

## Raw Logs

### Frontend Console Events (by route)

```
ROUTE: /
  CONSOLE_ERRORS: none
  CONSOLE_WARNS: none
  PAGE_ERRORS: none
  FAILED_REQUESTS: none
  SLOW_RESPONSES: none

ROUTE: /missions/summer-holiday-2026
  CONSOLE_ERRORS: [TypeError: Cannot read properties of null (reading 'length') at hu (index-_6c0Bw91.js:11:127247)]
  PAGE_ERRORS: [TypeError: Cannot read properties of null (reading 'length')]
  FAILED_REQUESTS: none (crash before requests complete)
  SLOW_RESPONSES: none

ROUTE: /missions/summer-holiday-2026/options
  CONSOLE_ERRORS: none
  FAILED_REQUESTS: none
  SLOW_RESPONSES: none

ROUTE: /missions/summer-holiday-2026/shopping
  CONSOLE_ERRORS: none
  FAILED_REQUESTS: none
  SLOW_RESPONSES: none

ROUTE: /search
  CONSOLE_ERRORS: none
  FAILED_REQUESTS: none
  SLOW_RESPONSES: none

ROUTE: /history
  CONSOLE_ERRORS: none
  FAILED_REQUESTS: none
  SLOW_RESPONSES: none

ROUTE: /stats
  CONSOLE_ERRORS: none (raw key names rendered — visual issue only)
  FAILED_REQUESTS: none
  SLOW_RESPONSES: none

ROUTE: /settings
  CONSOLE_ERRORS: none
  FAILED_REQUESTS: none
  SLOW_RESPONSES: none

ALL ROUTES (every page load):
  CONSOLE_WARNS: [ServiceWorker] Failed to register a ServiceWorker for scope
    'https://scouter.dev.local/' with script 'https://scouter.dev.local/sw.js':
    An SSL certificate error occurred when fetching the script.
```

### Backend Logs — Phase 1 Full Output

```
{"time":"2026-03-16T...","level":"ERROR","msg":"ollama call failed",
  "error":"dial tcp: lookup host.docker.internal on 127.0.0.11:53: no such host"}
{"time":"2026-03-16T...","level":"WARN","msg":"cascading to cloud provider",
  "provider":"deepseek-v3.2:cloud"}
{"time":"2026-03-16T...","level":"ERROR","msg":"cloud provider call failed",
  "error":"context canceled"}
```

404s observed on: `/api/missions/:id/decision`, `/api/missions/:id/forecast`,
`/api/missions/:id/purchase`, `/api/missions/:id/summary` — no backend log entries
(requests not reaching handlers — routes likely not registered).

### Backend Logs — Phase 2 Deltas

Phase 2 not executed — user proceeded directly to remediation.

---

## Fixes Applied in This Session

| Issue | Status | Files Changed |
|-------|--------|---------------|
| CRITICAL: `null.length` in CashbackSummaryPanel | ✅ Fixed | `backend/internal/cashbacktracker/handler.go`, `frontend/src/api/cashbacktracker.ts` |

## Remaining Work

| Issue | Priority | Estimated Effort |
|-------|----------|-----------------|
| 4 missing API endpoints (404) | HIGH | Audit `routes.go` — likely 1-2h |
| Ollama DNS in Docker | HIGH | 1 line in docker-compose.yml |
| ServiceWorker SSL in dev | MEDIUM | Disable in `import.meta.env.DEV` guard |
| Stats raw key names | LOW | Add display-name map in StatsPage |
| i18n hardcoded French strings | LOW | ~30 t() replacements across 3 files |
