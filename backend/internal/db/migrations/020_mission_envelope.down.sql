DROP INDEX IF EXISTS missions_envelope_id_idx;
ALTER TABLE missions DROP COLUMN IF EXISTS envelope_id;
