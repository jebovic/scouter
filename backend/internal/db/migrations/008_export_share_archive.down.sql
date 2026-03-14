-- Phase 10: Rollback Export, Share & Archive
DROP INDEX IF EXISTS missions_archived_at_idx;
DROP INDEX IF EXISTS missions_share_token_idx;

ALTER TABLE missions
  DROP COLUMN IF EXISTS archived_at,
  DROP COLUMN IF EXISTS share_token;
