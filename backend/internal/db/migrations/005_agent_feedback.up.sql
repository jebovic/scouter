-- Phase 5: Agent Feedback Loop & Refinement
-- Adds agent_runs table and pin/reject columns on options and shopping_items.

CREATE TABLE agent_runs (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id      UUID        NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  agent_type      TEXT        NOT NULL,   -- 'research' | 'pricing'
  input_snapshot  JSONB       NOT NULL DEFAULT '{}',
  result_snapshot JSONB       NOT NULL DEFAULT '{}',
  diff            JSONB       NOT NULL DEFAULT '{}',
  feedback        TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX agent_runs_mission_id_idx ON agent_runs (mission_id);
CREATE INDEX agent_runs_mission_type_idx ON agent_runs (mission_id, agent_type, created_at DESC);

ALTER TABLE options
  ADD COLUMN pinned        BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN rejected      BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN reject_reason TEXT;

ALTER TABLE shopping_items
  ADD COLUMN pinned BOOLEAN NOT NULL DEFAULT false;
