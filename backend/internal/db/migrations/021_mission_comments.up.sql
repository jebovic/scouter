CREATE TABLE mission_comments (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id  UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  option_id   UUID REFERENCES options(id) ON DELETE SET NULL,
  author      TEXT NOT NULL DEFAULT 'You',
  body        TEXT NOT NULL CHECK(char_length(body) <= 1000),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mission_comments_mission ON mission_comments(mission_id, created_at DESC);
