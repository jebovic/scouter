ALTER TABLE missions ADD COLUMN weight_profile JSONB NOT NULL DEFAULT '{}';

CREATE TABLE decisions (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id UUID        NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  scores     JSONB       NOT NULL,
  summary    TEXT        NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
