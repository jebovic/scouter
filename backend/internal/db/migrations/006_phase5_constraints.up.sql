-- Phase 5 fix: pin and reject are mutually exclusive on options
ALTER TABLE options
  ADD CONSTRAINT pin_reject_mutex CHECK (NOT (pinned AND rejected));

-- Phase 5 fix: enforce valid agent_type values
ALTER TABLE agent_runs
  ADD CONSTRAINT agent_type_valid CHECK (agent_type IN ('research', 'pricing'));
