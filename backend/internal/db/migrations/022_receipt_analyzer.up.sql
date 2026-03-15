CREATE TABLE receipt_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_slug TEXT NOT NULL REFERENCES missions(slug) ON DELETE CASCADE,
    raw_text TEXT NOT NULL,
    items JSONB NOT NULL DEFAULT '[]',
    total NUMERIC(12,2),
    currency TEXT NOT NULL DEFAULT 'EUR',
    analyzed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_receipt_analyses_mission ON receipt_analyses(mission_slug);
