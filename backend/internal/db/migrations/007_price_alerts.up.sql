-- Phase 7: Price Alerts & Deal Intelligence
-- Adds target_price to shopping_items and creates notifications table.

ALTER TABLE shopping_items
  ADD COLUMN target_price NUMERIC(12,2);

CREATE TABLE notifications (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id  UUID        NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  item_id     UUID        REFERENCES shopping_items(id) ON DELETE CASCADE,
  type        TEXT        NOT NULL,   -- 'price_drop' | 'target_hit' | 'trend_change'
  title       TEXT        NOT NULL,
  body        TEXT        NOT NULL,
  read        BOOLEAN     NOT NULL DEFAULT false,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX notifications_unread_idx ON notifications (read, created_at DESC) WHERE read = false;
CREATE INDEX notifications_mission_idx ON notifications (mission_id);
CREATE INDEX notifications_created_at_idx ON notifications (created_at DESC);

-- Dedup guard removed: TIMESTAMPTZ::date is timezone-dependent (not IMMUTABLE).
-- Application scheduler enforces one notification per item+type per day.
