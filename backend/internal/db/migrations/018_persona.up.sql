CREATE TABLE IF NOT EXISTS persona (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    archetype   TEXT NOT NULL,
    traits      JSONB NOT NULL DEFAULT '[]',
    tips        JSONB NOT NULL DEFAULT '[]',
    summary     TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
