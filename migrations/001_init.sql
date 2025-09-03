-- ByteSize: INIT SCHEMA

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS chunks (
    hash TEXT PRIMARY KEY,
    size INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename TEXT,
    total_size BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS file_chunks (
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    idx INT NOT NULL,
    chunk_hash TEXT NOT NULL REFERENCES chunks(hash),
    size INT NOT NULL,
    PRIMARY KEY (file_id, idx)
);

CREATE INDEX IF NOT EXISTS idx_file_chunks_chunk_hash ON file_chunks(chunk_hash);