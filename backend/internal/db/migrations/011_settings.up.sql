CREATE TABLE settings (
  key        TEXT        PRIMARY KEY,
  value      JSONB       NOT NULL DEFAULT '{}',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default settings
INSERT INTO settings (key, value) VALUES
  ('currency',     '"EUR"'),
  ('locale',       '"fr-FR"'),
  ('llm_provider', '"ollama"');
