# Database Schema

PostgreSQL 16 + pgvector. All migrations managed by golang-migrate and auto-applied at server startup.

---

## Core Tables

### missions

```sql
CREATE TABLE missions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    description TEXT,
    budget      NUMERIC(12,2) NOT NULL,
    category    TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'active',  -- active | done | archived
    constraints JSONB DEFAULT '{}',
    lessons     TEXT,
    share_token TEXT UNIQUE,
    archived_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX missions_slug_idx ON missions(slug);
CREATE INDEX missions_status_idx ON missions(status);
```

**Notes:**
- `slug` is the URL-safe identifier (e.g. `work-laptop-2026`)
- `constraints` JSONB stores flexible constraint schemas (min RAM, max weight, etc.)
- `share_token` enables public read-only sharing
- `archived_at` soft-archives without deletion

---

### options

```sql
CREATE TABLE options (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id  UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    brand       TEXT,
    description TEXT,
    price       NUMERIC(12,2),
    currency    TEXT DEFAULT 'EUR',
    status      TEXT DEFAULT 'watch',  -- buy | watch | defer | rejected | recommended | flash-sale | preorder | crisis
    deal_score  INTEGER,               -- 0–100
    attributes  JSONB DEFAULT '{}',    -- flexible per-category attributes
    source_url  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX options_mission_id_idx ON options(mission_id);
```

---

### shopping_items

```sql
CREATE TABLE shopping_items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id   UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
    option_id    UUID REFERENCES options(id),
    name         TEXT NOT NULL,
    merchant     TEXT,
    current_price NUMERIC(12,2),
    target_price  NUMERIC(12,2),
    currency     TEXT DEFAULT 'EUR',
    status       TEXT DEFAULT 'watching',
    deal_score   INTEGER,
    trend        TEXT,                 -- up | flat | down
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE price_history (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shopping_item_id UUID NOT NULL REFERENCES shopping_items(id) ON DELETE CASCADE,
    price           NUMERIC(12,2) NOT NULL,
    merchant        TEXT,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

### notifications

```sql
CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id  UUID REFERENCES missions(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    body        TEXT NOT NULL,
    type        TEXT NOT NULL,    -- price_alert | research_done | deal_found
    read        BOOLEAN DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX notifications_read_idx ON notifications(read) WHERE NOT read;
```

---

### purchase_records

```sql
CREATE TABLE purchase_records (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id  UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
    option_id   UUID REFERENCES options(id),
    merchant    TEXT,
    price_paid  NUMERIC(12,2) NOT NULL,
    currency    TEXT DEFAULT 'EUR',
    rating      INTEGER CHECK (rating BETWEEN 1 AND 5),
    notes       TEXT,
    purchased_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

### settings

```sql
CREATE TABLE settings (
    key         TEXT PRIMARY KEY,
    value       JSONB NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Defaults inserted on migration:
INSERT INTO settings (key, value) VALUES
    ('currency', '"EUR"'),
    ('locale',   '"fr-FR"'),
    ('llm_provider', '"ollama"');
```

---

### collaborators / votes / comments

```sql
CREATE TABLE collaborators (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id  UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
    user_email  TEXT NOT NULL,
    role        TEXT DEFAULT 'viewer',  -- viewer | editor
    invited_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE option_votes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    option_id   UUID NOT NULL REFERENCES options(id) ON DELETE CASCADE,
    user_email  TEXT NOT NULL,
    vote        INTEGER NOT NULL CHECK (vote IN (-1, 1)),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (option_id, user_email)
);

CREATE TABLE option_comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    option_id   UUID NOT NULL REFERENCES options(id) ON DELETE CASCADE,
    user_email  TEXT NOT NULL,
    body        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

### embeddings (pgvector)

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE embeddings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    option_id   UUID UNIQUE NOT NULL REFERENCES options(id) ON DELETE CASCADE,
    vec         vector(1024),    -- Voyage AI v3 compatible, nullable until embedded
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- IVFFlat index for approximate nearest neighbor search
CREATE INDEX embeddings_ivfflat_idx
    ON embeddings USING ivfflat (vec vector_cosine_ops)
    WITH (lists = 10);
```

ANN query used by semantic search:

```sql
WITH ranked AS (
    SELECT o.*, e.vec <=> $1::vector AS distance
    FROM options o
    JOIN embeddings e ON e.option_id = o.id
    ORDER BY distance
    LIMIT 20
)
SELECT * FROM ranked WHERE distance < 0.5;
```

---

### wishlist_items

```sql
CREATE TABLE wishlist_items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL,
    url          TEXT,
    target_price NUMERIC(12,2),
    priority     INTEGER DEFAULT 0,
    notes        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE wishlist_price_alerts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wishlist_item_id UUID NOT NULL REFERENCES wishlist_items(id) ON DELETE CASCADE,
    threshold_price  NUMERIC(12,2) NOT NULL,
    triggered       BOOLEAN DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

## Migration System

**Tool**: [golang-migrate](https://github.com/golang-migrate/migrate)

**Format**: `NNN_description.up.sql` / `NNN_description.down.sql`

**Auto-run**: Applied at every server startup via:

```go
m.Up() // applies pending migrations; noop if already current
```

**Current migrations**: 22+ files covering all phases from 001 (initial) through latest.

### Running Manually

```bash
# Apply pending
make migrate-up

# Rollback one step
make migrate-down

# Check status
docker compose exec backend ./server migrate-status
```

---

## Pagination Pattern

All list endpoints use cursor-based pagination with probe-row:

```sql
-- Probe row: fetch limit+1 to detect if there's a next page
SELECT id, name, budget, status, created_at
FROM missions
WHERE ($1::uuid IS NULL OR id < $1)
ORDER BY created_at DESC, id DESC
LIMIT $2 + 1;
```

```go
func BuildPagedResponse[T any](items []T, limit int) PagedResponse[T] {
    hasNext := len(items) > limit
    if hasNext {
        items = items[:limit]
    }
    return PagedResponse[T]{Data: items, HasNext: hasNext}
}
```

---

## Key Indexes

| Table | Index | Purpose |
|-------|-------|---------|
| missions | `slug` (unique) | URL routing lookup |
| missions | `status` | Filter active/done/archived |
| options | `mission_id` | List options for mission |
| notifications | `read` (partial) | Unread count query |
| embeddings | `ivfflat` | ANN cosine similarity |
| collaborators | `mission_id` | List collaborators |
