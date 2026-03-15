-- Phase 17: AI Negotiation Coach
-- Adds negotiation tracking columns to purchase_records.
ALTER TABLE purchase_records
    ADD COLUMN IF NOT EXISTS negotiation_suggested_price NUMERIC(12,2),
    ADD COLUMN IF NOT EXISTS negotiation_actual_price    NUMERIC(12,2),
    ADD COLUMN IF NOT EXISTS negotiation_notes           TEXT;
