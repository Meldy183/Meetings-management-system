CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_participants_name_trgm
    ON participants
    USING GIN (lower(last_name || ' ' || first_name || ' ' || middle_name) gin_trgm_ops);
