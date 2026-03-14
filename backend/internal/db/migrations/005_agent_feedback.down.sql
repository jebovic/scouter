-- Rollback Phase 5: Agent Feedback Loop & Refinement

ALTER TABLE shopping_items
  DROP COLUMN IF EXISTS pinned;

ALTER TABLE options
  DROP COLUMN IF EXISTS reject_reason,
  DROP COLUMN IF EXISTS rejected,
  DROP COLUMN IF EXISTS pinned;

DROP TABLE IF EXISTS agent_runs;
