-- Phase 11: Semantic Search — IVFFlat index for cosine-distance ANN search.
-- The embedding column (vector(1024)) was added in migration 001 but left un-indexed.
-- lists=10 is appropriate for datasets up to ~10 000 rows; increase when needed.
CREATE INDEX IF NOT EXISTS options_embedding_ivfflat_idx
    ON options USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 10);
