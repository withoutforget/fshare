CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE files
(
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    s3_key      TEXT        NOT NULL UNIQUE,
    uploaded_at TIMESTAMPTZ NOT NULL,
    uploaded_by TEXT        NOT NULL
);

CREATE INDEX idx_files_uploaded_by ON files (uploaded_by);
CREATE INDEX idx_files_uploaded_at ON files (uploaded_at);