ALTER TABLE missions ADD COLUMN envelope_id UUID REFERENCES envelopes(id) ON DELETE SET NULL;
CREATE INDEX missions_envelope_id_idx ON missions(envelope_id);
