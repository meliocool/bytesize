-- ByteSize: INIT SCHEMA

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS chunks (
    hash TEXT PRIMARY KEY,
    diskSize INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS manifest (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename TEXT,
    totalSize BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS file_chunks (
    file_id UUID NOT NULL REFERENCES manifest(id) ON DELETE CASCADE,
    idx INT NOT NULL,
    chunk_hash TEXT NOT NULL REFERENCES chunks(hash),
    diskSize INT NOT NULL,
    PRIMARY KEY (file_id, idx)
);

CREATE INDEX IF NOT EXISTS idx_file_chunks_chunk_hash ON file_chunks(chunk_hash);