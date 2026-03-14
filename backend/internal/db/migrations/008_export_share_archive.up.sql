-- Phase 10: Export, Share & Archive
ALTER TABLE missions
  ADD COLUMN share_token TEXT UNIQUE,
  ADD COLUMN archived_at TIMESTAMPTZ;

CREATE INDEX missions_share_token_idx ON missions (share_token) WHERE share_token IS NOT NULL;
CREATE INDEX missions_archived_at_idx ON missions (archived_at) WHERE archived_at IS NOT NULL;
