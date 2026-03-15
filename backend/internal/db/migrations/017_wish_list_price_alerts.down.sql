ALTER TABLE wish_list_items
  DROP COLUMN IF EXISTS last_checked_at,
  DROP COLUMN IF EXISTS last_checked_price,
  DROP COLUMN IF EXISTS alert_enabled;
