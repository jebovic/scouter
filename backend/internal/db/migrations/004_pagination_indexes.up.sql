-- Indexes to support cursor-based pagination (created_at DESC) on all list endpoints.
CREATE INDEX missions_created_at_idx         ON missions       (created_at DESC);
CREATE INDEX options_created_at_idx          ON options        (created_at DESC);
CREATE INDEX shopping_items_created_at_idx   ON shopping_items (created_at DESC);
