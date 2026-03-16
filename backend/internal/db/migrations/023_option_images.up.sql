CREATE TABLE option_images (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    option_id    UUID NOT NULL REFERENCES options(id) ON DELETE CASCADE,
    minio_key    TEXT NOT NULL UNIQUE,
    content_type TEXT NOT NULL,
    width        INT,
    height       INT,
    source_url   TEXT,
    sort_order   INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON option_images(option_id);
