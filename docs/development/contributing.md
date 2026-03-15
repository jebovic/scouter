# Contributing

Development workflow, conventions, and processes for SCOUTER Universal.

---

## Phase Workflow

Every feature or fix follows this 11-step cycle:

```
1.  /everything-claude-code:plan + architect review
2.  Detailed plan: PRD, architecture, tech_doc, task_list
3.  TDD: write tests first (RED)
4.  Implement to pass tests (GREEN)
5.  Refactor (IMPROVE)
6.  Backend: go test ./... (verify coverage ≥ 80%)
7.  ECC go-reviewer agent
8.  Frontend: npm run build + npm run typecheck
9.  Frontend-design review
10. Commit + push (conventional commits)
11. Close phase in ROADMAP.md
```

---

## Repository Layout

```
scouter/
├── backend/
│   ├── cmd/server/
│   │   ├── main.go          # entrypoint, env validation, startup
│   │   └── routes.go        # route registration (routeDeps + registerRoutes)
│   ├── internal/            # all Go packages (132 total)
│   │   ├── config/          # Config + Load() from env
│   │   ├── db/              # pgx pool + golang-migrate runner
│   │   ├── httputil/        # WriteJSON / WriteError (buffer-first)
│   │   ├── llm/             # Provider interface + SmartRouter
│   │   ├── mission/         # model, repository, service, handler
│   │   └── ...              # other domain packages
│   ├── migrations/          # NNN_description.up/down.sql
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── api/             # typed fetch wrappers + Zod schemas
│   │   ├── components/      # UI components (CSS Modules)
│   │   ├── pages/           # route-level components
│   │   ├── hooks/           # TanStack Query hooks
│   │   └── styles/          # theme.css + global.css
│   └── Dockerfile
├── monitoring/              # Prometheus + Grafana config
├── docs/                    # this documentation
└── docker-compose.yml
```

---

## Coding Style

### Immutability (critical)

Never mutate existing objects. Always return new values.

```go
// WRONG
func updateMission(m *Mission) {
    m.Status = "done"     // mutates in place
}

// CORRECT
func updateMission(m Mission, status string) Mission {
    m.Status = status     // copy-on-write
    return m
}
```

```ts
// WRONG
options.push(newOption);

// CORRECT
const updated = [...options, newOption];
```

### File Size Limits

| Threshold | Action |
|-----------|--------|
| < 200 lines | Fine |
| 200–400 lines | Typical maximum |
| 400–800 lines | Extract utilities |
| > 800 lines | Mandatory split |

### Function Size

- Maximum 50 lines per function
- No nesting deeper than 4 levels
- One responsibility per function

---

## Go Conventions

### Package Structure

Each domain package has four files:

```
internal/<domain>/
├── model.go       # struct definitions only
├── repository.go  # pgx queries (Repository interface + pgxRepository)
├── service.go     # business logic (Service interface + service)
└── handler.go     # chi handlers (Handler struct, ServeHTTP methods)
```

### Error Handling

```go
// Always check pgx.ErrNoRows specifically
row, err := repo.GetByID(ctx, id)
if errors.Is(err, pgx.ErrNoRows) {
    httputil.WriteError(w, http.StatusNotFound, "mission not found")
    return
}
if err != nil {
    httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
    return
}
```

### HTTP Handlers

```go
// Always buffer first — WriteJSON handles header+body atomically
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    result, err := h.svc.Create(r.Context(), req)
    if err != nil {
        httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
        return
    }
    httputil.WriteJSON(w, http.StatusCreated, result)
}
```

### Migrations

Add new migrations in `backend/migrations/`:

```
NNN_description.up.sql    # forward migration
NNN_description.down.sql  # rollback
```

Migrations auto-apply at server startup. Test rollback before committing:

```bash
make migrate-down
make migrate-up
```

---

## TypeScript / React Conventions

### API Layer

Every API resource needs a Zod schema and typed fetch wrapper:

```ts
// src/api/example.ts
import { z } from 'zod';
import { apiFetch } from './client';

export const ExampleSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
});
export type Example = z.infer<typeof ExampleSchema>;

export async function getExample(id: string): Promise<Example> {
  return apiFetch(`/api/examples/${id}`, ExampleSchema);
}
```

### Hooks

Use TanStack Query v5 patterns:

```ts
// src/hooks/useExample.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

export function useExample(id: string) {
  return useQuery({
    queryKey: ['example', id],
    queryFn: () => getExample(id),
  });
}
```

### CSS Modules

No raw `style={{}}` for layout or theming. Use co-located `.module.css`:

```tsx
// Component.tsx
import styles from './Component.module.css';

// Dynamic values via CSS custom properties
<div
  className={styles.card}
  style={{ '--progress': `${pct}%` } as React.CSSProperties}
/>
```

```css
/* Component.module.css */
.card {
  background: var(--surface);
  border-radius: 16px;
}
.bar {
  width: var(--progress);  /* consumes the dynamic property */
}
```

### i18n

All user-visible strings use the `useTranslation` hook:

```tsx
import { useTranslation } from '../i18n';

function MyComponent() {
  const { t } = useTranslation();
  return <h1>{t('missions.title')}</h1>;
}
```

Add new keys to both `src/i18n/en.json` and `src/i18n/fr.json`.

---

## Git Workflow

### Commit Format

```
<type>: <description>

<optional body>
```

**Types:** `feat` `fix` `refactor` `docs` `test` `chore` `perf` `ci`

### Examples

```
feat(backend): add quantity optimizer endpoint with FNV-32a discount curve

fix(frontend): correct cursor pagination on mission list scroll

test(backend): add service-level tests for scorecard grading logic

docs: update ROADMAP.md to mark phase 172 complete
```

### Pre-commit Checklist

- [ ] No hardcoded secrets (API keys, passwords, tokens)
- [ ] All user inputs validated at system boundaries
- [ ] SQL uses parameterized queries only
- [ ] Error messages do not leak internal detail to client
- [ ] New env vars documented in `.env.example` and `docs/deployment/environment.md`
- [ ] `go vet ./...` passes
- [ ] `npm run typecheck` passes

---

## Local Development Setup

### Prerequisites

- Go 1.23+
- Node.js 20+
- Docker + Docker Compose v2
- (Optional) Ollama for local LLM

### First-time Setup

```bash
git clone <repo-url>
cd scouter
cp .env.example .env
# Edit .env: set DATABASE_URL and LLM_PROVIDER

# Start all services
make dev
# or: docker compose up --build

# Load sample data
make seed
```

### Backend Only (hot reload)

```bash
cd backend
go run ./cmd/server
```

Set `ENV=development` in `.env` for permissive CORS.

### Frontend Only (Vite dev server)

```bash
cd frontend
npm install
npm run dev
```

Ensure backend is running on `:8080`.

### Useful Make Targets

| Command | Action |
|---------|--------|
| `make dev` | Start full stack (Docker Compose) |
| `make seed` | Load sample data |
| `make clean` | Stop + delete volumes |
| `make migrate-up` | Apply pending migrations |
| `make migrate-down` | Rollback one migration |
| `make test` | Run all backend tests |
| `make lint` | Run golangci-lint |

---

## Adding a New Backend Package

1. Create `internal/<name>/` with `model.go`, `repository.go`, `service.go`, `handler.go`
2. Add pgx queries using parameterized `$1`, `$2` placeholders
3. Register handler in `cmd/server/routes.go` via `routeDeps`
4. Write migration if schema changes needed
5. Add Prometheus instrumentation via `metrics.Recorder` if the package handles external calls

## Adding a New Frontend Page

1. Create `src/pages/MyPage.tsx` + `src/pages/MyPage.module.css`
2. Add route in `src/App.tsx` under the appropriate layout
3. Create `src/api/myResource.ts` with Zod schema + typed fetch
4. Create `src/hooks/useMyResource.ts` with TanStack Query hooks
5. Add i18n keys to `en.json` and `fr.json`
6. Add to Breadcrumb if nested under a mission

---

## Pull Request Process

1. Branch from `main`
2. Follow the Phase Workflow above
3. `git diff main...HEAD` to verify scope
4. PR title < 70 characters
5. PR body: Summary bullets + Test plan checklist
6. All checks must pass before merge
