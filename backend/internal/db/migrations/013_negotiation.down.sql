-- Phase 17: AI Negotiation Coach (rollback)
ALTER TABLE purchase_records
    DROP COLUMN IF EXISTS negotiation_suggested_price,
    DROP COLUMN IF EXISTS negotiation_actual_price,
    DROP COLUMN IF EXISTS negotiation_notes;
