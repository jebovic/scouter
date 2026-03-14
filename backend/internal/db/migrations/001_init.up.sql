-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- missions
CREATE TABLE missions (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  slug            TEXT        NOT NULL UNIQUE,
  name            TEXT        NOT NULL,
  icon            TEXT        NOT NULL DEFAULT 'target',
  category        TEXT        NOT NULL,   -- travel | renovation | electronics | computing | custom
  budget          NUMERIC(12,2) NOT NULL,
  currency        TEXT        NOT NULL DEFAULT 'EUR',
  locale          TEXT        NOT NULL DEFAULT 'fr-FR',
  phase           TEXT        NOT NULL DEFAULT 'researching',
  constraints     JSONB       NOT NULL DEFAULT '[]',
  cost_categories JSONB       NOT NULL DEFAULT '[]',
  timeline        JSONB       NOT NULL DEFAULT '[]',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX missions_slug_idx ON missions (slug);

-- options (flexible attributes via jsonb)
CREATE TABLE options (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id  UUID        NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  name        TEXT        NOT NULL,
  category    TEXT        NOT NULL,
  badge       TEXT        NOT NULL DEFAULT 'watch', -- recommended | alternative | rejected | watch
  attributes  JSONB       NOT NULL DEFAULT '[]',
  price_range JSONB,                                -- {min, max, best}
  notes       TEXT,
  warnings    JSONB       NOT NULL DEFAULT '[]',
  url         TEXT,
  embedding   vector(1024),                         -- pgvector: nullable until embedding provider chosen
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX options_mission_id_idx ON options (mission_id);

-- shopping items
CREATE TABLE shopping_items (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id        UUID        NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  name              TEXT        NOT NULL,
  merchant          TEXT        NOT NULL,
  cost_category     TEXT        NOT NULL,
  price             NUMERIC(12,2) NOT NULL,
  original_estimate NUMERIC(12,2),
  status            TEXT        NOT NULL DEFAULT 'watch', -- buy | flash-sale | preorder | defer | watch | crisis
  note              TEXT,
  url               TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX shopping_items_mission_id_idx ON shopping_items (mission_id);

-- price history (per shopping item)
CREATE TABLE price_history (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  item_id     UUID        NOT NULL REFERENCES shopping_items(id) ON DELETE CASCADE,
  price       NUMERIC(12,2) NOT NULL,
  recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  note        TEXT
);

CREATE INDEX price_history_item_id_idx ON price_history (item_id);
