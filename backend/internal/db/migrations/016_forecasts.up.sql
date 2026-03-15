CREATE TABLE IF NOT EXISTS forecasts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id      UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
    estimated_total NUMERIC(12,2),
    confidence      TEXT NOT NULL DEFAULT 'medium',
    recommendations JSONB NOT NULL DEFAULT '[]',
    risk_factors    JSONB NOT NULL DEFAULT '[]',
    months_to_save  INT,
    optimal_buy_time TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_forecasts_mission_id ON forecasts(mission_id);
