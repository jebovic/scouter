CREATE TABLE IF NOT EXISTS wish_list_items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL,
    url          TEXT,
    target_price NUMERIC(12,2),
    currency     TEXT NOT NULL DEFAULT 'EUR',
    notes        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wish_list_items_created_at ON wish_list_items(created_at DESC);
