CREATE TABLE IF NOT EXISTS envelopes (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name           TEXT NOT NULL,
  monthly_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
  currency       TEXT NOT NULL DEFAULT 'EUR',
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
