# API Reference

Base URL: `http://localhost:8080/api`

All responses use `Content-Type: application/json`. Errors return `{ "error": "message" }`.

---

## Health

```
GET  /api/health        # System health (DB ping, degraded status)
GET  /api/health/llm    # SmartRouter pool status
GET  /metrics           # Prometheus metrics (METRICS_ENABLED=true)
```

### GET /api/health

```json
{
  "status": "ok",           // "ok" | "degraded"
  "db": "connected",
  "version": "0.1.0"
}
```

### GET /api/health/llm

```json
{
  "healthy": true,
  "providers": [
    { "name": "ollama-heavy", "status": "ok", "circuit": "closed" },
    { "name": "anthropic",    "status": "ok", "circuit": "closed" }
  ]
}
```

---

## Missions

```
GET    /api/missions              # List (cursor-paginated)
POST   /api/missions              # Create
GET    /api/missions/:slug        # Get by slug
PATCH  /api/missions/:slug        # Update
DELETE /api/missions/:slug        # Delete
POST   /api/missions/:slug/duplicate   # Clone (shallow)
POST   /api/missions/:slug/clone       # Deep clone (with options)
POST   /api/missions/:id/archive       # Archive
POST   /api/missions/:id/unarchive     # Unarchive
POST   /api/missions/:id/share         # Generate share token
DELETE /api/missions/:id/share         # Revoke share token
```

### GET /api/missions

Query params: `limit` (default 20), `cursor` (UUID), `status` (active|done|archived)

```json
{
  "data": [
    {
      "id": "uuid",
      "slug": "work-laptop-2026",
      "name": "Work Laptop Upgrade",
      "budget": 1500.00,
      "status": "active",
      "category": "electronics",
      "constraints": { "min_ram_gb": 16, "max_weight_kg": 2 },
      "created_at": "2026-03-15T10:00:00Z"
    }
  ],
  "next_cursor": "uuid-or-null",
  "total": 42
}
```

### POST /api/missions

```json
{
  "name": "Work Laptop Upgrade",
  "budget": 1500.00,
  "category": "electronics",
  "constraints": { "min_ram_gb": 16, "max_weight_kg": 2 },
  "description": "Replacing my 2019 MacBook Pro"
}
```

### GET /api/shared/:token (CORS-open)

Public read-only access to a shared mission. No authentication required.

---

## Templates

```
GET  /api/templates       # List all 15 built-in templates
GET  /api/templates/:id   # Get template by ID
```

```json
{
  "id": "laptop",
  "name": "Laptop",
  "category": "electronics",
  "default_budget": 1200,
  "constraints": {
    "min_ram_gb": 8,
    "min_storage_gb": 256,
    "max_weight_kg": 2.5
  }
}
```

---

## AI Agents

### Research

```
POST /api/missions/:id/research    # Run ResearchAgent
```

Triggers ResearchAgent → discovers 5–10 options via LLM tool use → persists to DB.

Response: `{ "options": [...], "count": N }`

### Pricing

```
POST /api/missions/:id/pricing     # Run PricingAgent
```

Hunts prices across merchants → calculates TCO + deal score → updates shopping_items.

### Decision

```
POST /api/missions/:id/decision    # Run DecisionAgent
GET  /api/missions/:id/decision    # Get last decision
```

---

## Options

```
GET    /api/missions/:id/options          # List options for mission
POST   /api/missions/:id/options          # Add option manually
GET    /api/options/:id                   # Get option
PATCH  /api/options/:id                   # Update (status, attributes)
DELETE /api/options/:id                   # Delete
GET    /api/options/:id/similar           # Find similar options (pgvector)
GET    /api/options/:id/votes             # Get collaboration votes
POST   /api/options/:id/votes             # Cast vote
POST   /api/options/:id/comments          # Add comment
GET    /api/options/:id/comments          # List comments
```

---

## Shopping

```
GET    /api/missions/:id/shopping         # List shopping items
POST   /api/missions/:id/shopping         # Add item + price history
GET    /api/shopping/:id                  # Get item
PATCH  /api/shopping/:id                  # Update (target_price, status)
DELETE /api/shopping/:id                  # Delete
POST   /api/shopping/:id/price-history    # Add price point
GET    /api/shopping/:id/price-history    # Price history
GET    /api/missions/:id/deal-score       # Get mission deal score
```

---

## Specialized Intelligence Endpoints

### Phase 168 — Wishlist Prioritizer

```
GET /api/wishlist/prioritized
```

Returns wishlist items ranked by composite score (urgency × trend × budgetFit).
Cache: 15min · FNV-32a deterministic scoring.

### Phase 169 — French Market Benchmark

```
GET /api/missions/:id/french-benchmark
```

```json
{
  "market_median": 1150.00,
  "currency": "EUR",
  "verdict": "bon_prix",  // bon_prix | prix_moyen | au_dessus_du_marché
  "items": [
    { "name": "MacBook Air M3", "your_price": 999, "market_median": 1150, "verdict": "bon_prix" }
  ]
}
```

Cache: 20min.

### Phase 170 — Mission Scorecard

```
GET /api/missions/:id/scorecard
```

```json
{
  "grade": "A",
  "score": 87,
  "breakdown": {
    "price_efficiency": 92,
    "research_depth": 88,
    "time_to_decision": 85,
    "budget_discipline": 83
  },
  "achievements": ["Under budget", "Thorough research"],
  "lessons": "Should have checked more merchants"
}
```

Cache: 30min.

### Phase 171 — Quantity Optimizer

```
GET /api/missions/:id/items/:itemId/quantity-optimizer
```

```json
{
  "base_price": 49.99,
  "tiers": [
    { "qty": 1, "unit_price": 49.99, "total": 49.99,  "discount_pct": 0 },
    { "qty": 2, "unit_price": 47.49, "total": 94.98,  "discount_pct": 5 },
    { "qty": 3, "unit_price": 44.99, "total": 134.97, "discount_pct": 10 },
    { "qty": 5, "unit_price": 42.49, "total": 212.45, "discount_pct": 15 },
    { "qty": 10,"unit_price": 39.99, "total": 399.90, "discount_pct": 20 }
  ]
}
```

Cache: 10min.

### Phase 172 — Purchase Timeline Planner

```
GET /api/missions/:id/purchase-timeline
```

```json
{
  "weeks": [
    { "week": 1, "label": "Now",     "items": [...], "budget_pct": 40, "hint": "Check Black Friday deals" },
    { "week": 2, "label": "Soon",    "items": [...], "budget_pct": 30, "hint": null },
    { "week": 3, "label": "Later",   "items": [...], "budget_pct": 20, "hint": "Soldes start in 2 weeks" },
    { "week": 4, "label": "Defer",   "items": [...], "budget_pct": 10, "hint": null }
  ]
}
```

Cache: 20min.

---

## Notifications

```
GET   /api/notifications                 # List (paginated)
PATCH /api/notifications/:id/read        # Mark as read
POST  /api/notifications/mark-all-read   # Mark all as read
GET   /api/notifications/unread-count    # { "count": N }
```

---

## Purchase & Stats

```
GET    /api/missions/:id/purchase   # Get purchase records
POST   /api/missions/:id/purchase   # Record a purchase (advances mission to "done")
PATCH  /api/missions/:id/purchase   # Update lessons
GET    /api/stats                   # Total spend + category breakdown
```

### GET /api/stats

```json
{
  "total_spent": 4250.00,
  "currency": "EUR",
  "missions_completed": 8,
  "avg_deal_score": 73,
  "by_category": [
    { "category": "electronics", "spent": 2800, "count": 4 },
    { "category": "travel",      "spent": 1450, "count": 4 }
  ]
}
```

---

## Search

```
GET  /api/search?q=laptop&limit=20   # Semantic search (pgvector)
POST /api/search/reindex             # Rebuild all embeddings
```

---

## Collaboration

```
POST /api/missions/:id/invites        # Send invite
POST /api/invites/:token/join         # Accept invite (→ JoinPage)
GET  /api/missions/:id/collaborators  # List collaborators
DELETE /api/missions/:id/collaborators/:email  # Remove
```

---

## Wishlist

```
GET    /api/wishlist                    # List items
POST   /api/wishlist                    # Add item
PATCH  /api/wishlist/:id                # Update (target_price, priority)
DELETE /api/wishlist/:id                # Remove
GET    /api/wishlist/prioritized        # Ranked by urgency + trend + budget
POST   /api/wishlist/:id/price-alerts   # Set price alert
GET    /api/wishlist/share              # Get share token
```

---

## Settings & Admin

```
GET   /api/settings          # All settings
PATCH /api/settings          # Update (currency, locale, llm_provider)
DELETE /api/data             # Delete ALL data (X-Confirm: yes header required)
```

### PATCH /api/settings

```json
{
  "currency": "USD",
  "locale": "en-US",
  "llm_provider": "anthropic"
}
```

Allowed values:
- `currency`: ISO 4217 codes (EUR, USD, GBP, CHF, CAD…)
- `locale`: IETF language tags (fr-FR, en-US, en-GB…)
- `llm_provider`: `ollama`, `anthropic`, `routing`

---

## Error Responses

```json
{ "error": "mission not found" }
{ "error": "invalid request body" }
{ "error": "internal server error" }
```

HTTP status codes:
- `200` Success
- `201` Created
- `204` No content (delete)
- `400` Bad request (validation)
- `404` Not found
- `500` Internal server error
