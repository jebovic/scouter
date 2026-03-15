CREATE TABLE IF NOT EXISTS comparison_weights (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id  UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
    criteria    TEXT NOT NULL,
    weight      INT NOT NULL DEFAULT 50 CHECK (weight >= 0 AND weight <= 100),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (mission_id, criteria)
);
CREATE INDEX IF NOT EXISTS idx_comparison_weights_mission_id ON comparison_weights(mission_id);
