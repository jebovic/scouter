-- Add lessons field to missions
ALTER TABLE missions ADD COLUMN lessons TEXT;

-- Create purchase_records table
CREATE TABLE purchase_records (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  option_id UUID REFERENCES options(id) ON DELETE SET NULL,
  purchased_at DATE NOT NULL,
  final_price NUMERIC(12,2) NOT NULL,
  merchant TEXT NOT NULL,
  satisfaction SMALLINT CHECK (satisfaction BETWEEN 1 AND 5),
  review TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX purchase_records_mission_id_idx ON purchase_records(mission_id);
