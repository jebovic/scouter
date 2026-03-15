# Testing

Test strategy, tooling, and coverage requirements for SCOUTER Universal.

---

## Overview

| Layer | Framework | Target Coverage |
|-------|-----------|-----------------|
| Backend (Go) | `go test` + `testify` | ≥ 80% |
| Frontend (TypeScript) | Vitest + Testing Library | ≥ 80% |
| E2E | Playwright | Critical flows |

---

## Test-Driven Development (Mandatory)

All new features and bug fixes follow the RED → GREEN → IMPROVE cycle:

```
1. Write failing test  (RED)
2. Run: test must FAIL
3. Write minimal implementation (GREEN)
4. Run: test must PASS
5. Refactor without breaking tests (IMPROVE)
6. Verify coverage ≥ 80%
```

Never write implementation code before the test. This is enforced by the `everything-claude-code:tdd-guide` agent.

---

## Backend Testing (Go)

### Running Tests

```bash
# All tests
cd backend
go test ./...

# Specific package
go test ./internal/mission/...

# With coverage report
go test -cover ./...

# Coverage by package
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Structure

Tests live alongside the code they test:

```
internal/mission/
├── model.go
├── repository.go
├── repository_test.go   # integration tests (real pgx)
├── service.go
├── service_test.go      # unit tests (mocked repository)
└── handler_test.go      # HTTP tests (httptest)
```

### Unit Tests

Unit tests mock external dependencies using interfaces:

```go
// service_test.go
type mockRepo struct {
    missions []Mission
}

func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*Mission, error) {
    for _, mis := range m.missions {
        if mis.ID == id {
            return &mis, nil
        }
    }
    return nil, pgx.ErrNoRows
}

func TestService_GetByID_NotFound(t *testing.T) {
    svc := NewService(&mockRepo{})
    _, err := svc.GetByID(context.Background(), uuid.New())
    require.ErrorIs(t, err, ErrNotFound)
}
```

### Table-Driven Tests

Prefer table-driven tests for multiple cases:

```go
func TestDealScore_Calculate(t *testing.T) {
    tests := []struct {
        name     string
        current  float64
        target   float64
        trend    string
        want     int
    }{
        {"below target down trend", 80, 100, "down", 95},
        {"at target flat",         100, 100, "flat", 70},
        {"above target up trend",  120, 100, "up",   20},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got := Calculate(tc.current, tc.target, tc.trend)
            assert.Equal(t, tc.want, got)
        })
    }
}
```

### HTTP Handler Tests

Use `net/http/httptest` for handler tests:

```go
func TestHandler_Create(t *testing.T) {
    body := `{"name":"Test Mission","budget":1000,"category":"electronics"}`
    req := httptest.NewRequest(http.MethodPost, "/api/missions", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    h := NewHandler(mockSvc)
    h.Create(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
    var result Mission
    require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
    assert.Equal(t, "Test Mission", result.Name)
}
```

### Integration Tests (Database)

Integration tests use a real PostgreSQL connection. Set `TEST_DATABASE_URL` in `.env.test`:

```bash
TEST_DATABASE_URL=postgres://scouter:scouter@localhost:5432/scouter_test
```

```go
//go:build integration

func TestRepository_Create_Integration(t *testing.T) {
    pool := testdb.Setup(t)  // sets up + tears down test DB
    repo := NewRepository(pool)

    m, err := repo.Create(context.Background(), CreateParams{
        Name:     "Integration Test Mission",
        Budget:   500,
        Category: "electronics",
    })
    require.NoError(t, err)
    assert.NotEmpty(t, m.ID)
    assert.NotEmpty(t, m.Slug)
}
```

Run integration tests explicitly:

```bash
go test -tags integration ./...
```

### Key Packages with Tests

| Package | Test Focus |
|---------|------------|
| `internal/dealintel` | Trend calculation, deal score, pure Go |
| `internal/mission` | CRUD, cursor pagination, slug generation |
| `internal/scheduler` | Price alert trigger logic |
| `internal/wishlistprioritizer` | FNV-32a determinism, composite score |
| `internal/frenchbenchmark` | Market median, verdict thresholds |
| `internal/scorecard` | Grade A/B/C/D boundary conditions |
| `internal/quantityoptimizer` | Tier price curve accuracy |
| `internal/timelineplanner` | Status-based week distribution |

---

## Frontend Testing (Vitest + Testing Library)

### Running Tests

```bash
cd frontend

# All tests (single run)
npm test

# Watch mode
npm run test:watch

# Coverage report
npm run test:coverage
```

### Test Structure

Tests live alongside components:

```
src/components/mission/
├── MissionCard.tsx
├── MissionCard.module.css
└── MissionCard.test.tsx
```

### Component Tests

```tsx
// MissionCard.test.tsx
import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { MissionCard } from './MissionCard';

const mockMission = {
  id: '123e4567-e89b-12d3-a456-426614174000',
  slug: 'work-laptop-2026',
  name: 'Work Laptop Upgrade',
  budget: 1500,
  status: 'active' as const,
  category: 'electronics',
  created_at: '2026-03-15T10:00:00Z',
};

describe('MissionCard', () => {
  it('renders mission name and budget', () => {
    render(<MissionCard mission={mockMission} />);
    expect(screen.getByText('Work Laptop Upgrade')).toBeInTheDocument();
    expect(screen.getByText(/1\s*500/)).toBeInTheDocument();
  });

  it('shows active status badge', () => {
    render(<MissionCard mission={mockMission} />);
    expect(screen.getByRole('status')).toHaveTextContent('active');
  });
});
```

### Hook Tests

```tsx
// useNotifications.test.tsx
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';
import { useNotifications } from './useNotifications';
import { vi } from 'vitest';

vi.mock('../api/notifications', () => ({
  listNotifications: vi.fn().mockResolvedValue({ data: [], total: 0 }),
}));

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <QueryClientProvider client={new QueryClient()}>
    {children}
  </QueryClientProvider>
);

it('returns empty list initially', async () => {
  const { result } = renderHook(() => useNotifications(), { wrapper });
  await waitFor(() => expect(result.current.isSuccess).toBe(true));
  expect(result.current.data?.data).toHaveLength(0);
});
```

### Test Utilities

Common test helpers live in `src/test/`:

```
src/test/
├── setup.ts         # vitest global setup (@testing-library/jest-dom)
├── render.tsx       # custom render with providers (QueryClient, Router)
└── factories.ts     # test data factories (createMission, createOption, etc.)
```

Custom render wrapper:

```tsx
// src/test/render.tsx
export function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>{ui}</MemoryRouter>
    </QueryClientProvider>
  );
}
```

### Current Test Files (71+ tests)

| File | Count | Focus |
|------|-------|-------|
| `MissionCard.test.tsx` | 8 | Rendering, status badges, budget display |
| `OptionCard.test.tsx` | 7 | Attributes, deal score, status |
| `ShoppingItemRow.test.tsx` | 9 | Price display, trend badge, target price edit |
| `useSearch.test.ts` | 6 | Debounce, min-length guard, result mapping |
| `BudgetBar.test.tsx` | 5 | Percentage calculation, overflow state |
| `DealScoreBadge.test.tsx` | 4 | Score ranges, color classes |
| `PriceHistoryChart.test.tsx` | 6 | Data points, empty state |
| `StarRating.test.tsx` | 8 | Click, keyboard, aria-label |
| `useNotifications.test.ts` | 5 | Polling, unread count |
| `WishlistPriorityCard.test.tsx` | 7 | Score display, gold highlight for top pick |
| Additional files | 16+ | — |

---

## E2E Testing (Playwright)

### Setup

```bash
cd frontend
npx playwright install
```

### Running E2E Tests

```bash
# All E2E tests
npm run test:e2e

# Headed (see the browser)
npm run test:e2e -- --headed

# Single test file
npm run test:e2e -- e2e/mission-flow.spec.ts
```

### Critical Flows Covered

| Flow | File |
|------|------|
| Create mission → run research → see options | `e2e/research-flow.spec.ts` |
| Add to shopping list → track price | `e2e/shopping-flow.spec.ts` |
| Record purchase → view history | `e2e/purchase-flow.spec.ts` |
| Search options semantically | `e2e/search-flow.spec.ts` |

### Example E2E Test

```ts
// e2e/mission-flow.spec.ts
import { test, expect } from '@playwright/test';

test('create mission and run research', async ({ page }) => {
  await page.goto('http://localhost:5173');

  // Create mission
  await page.click('[data-testid="new-mission-btn"]');
  await page.fill('[name="name"]', 'Gaming Laptop Test');
  await page.fill('[name="budget"]', '1200');
  await page.selectOption('[name="category"]', 'electronics');
  await page.click('[type="submit"]');

  // Verify created
  await expect(page.locator('h1')).toContainText('Gaming Laptop Test');

  // Run research (mocked in test environment)
  await page.click('[data-testid="research-btn"]');
  await expect(page.locator('[data-testid="options-list"]')).toBeVisible();
});
```

---

## CI Pipeline

Tests run automatically on every push:

```yaml
# .github/workflows/test.yml (conceptual)
jobs:
  backend:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: pgvector/pgvector:pg16
    steps:
      - go test -cover ./...
      - go vet ./...

  frontend:
    runs-on: ubuntu-latest
    steps:
      - npm ci
      - npm run typecheck
      - npm test
      - npm run build
```

---

## Coverage Enforcement

### Backend

Minimum 80% coverage per package. Check with:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total
```

### Frontend

```bash
npm run test:coverage
# Opens HTML report at coverage/index.html
```

Vitest config enforces thresholds in `vitest.config.ts`:

```ts
coverage: {
  thresholds: {
    lines: 80,
    functions: 80,
    branches: 75,
  },
}
```

---

## Testing Anti-Patterns to Avoid

| Anti-pattern | Why | Instead |
|-------------|-----|---------|
| Testing implementation details | Tests break on refactor | Test behavior/output |
| Mocking the database in integration tests | Masks real divergence | Use test DB |
| `time.Sleep` in tests | Flaky | Use channels or `waitFor` |
| Large test functions | Hard to debug | One assertion per test |
| Skipping edge cases | Bugs hide there | Table-driven with boundary values |
| `t.Parallel()` without isolation | Race conditions | Isolate shared state first |
