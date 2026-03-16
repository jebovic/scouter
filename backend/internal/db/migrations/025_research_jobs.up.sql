CREATE TABLE research_jobs (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id    UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  status        TEXT NOT NULL DEFAULT 'pending',
  feedback      TEXT,
  error         TEXT,
  options_count INT,
  started_at    TIMESTAMPTZ,
  completed_at  TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON research_jobs(mission_id, created_at DESC);
