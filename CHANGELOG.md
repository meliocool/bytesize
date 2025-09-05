# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

---

## [0.2.0] - 2025-09-05
### Added
- **Download service**: streams reconstructed file bytes chunk-by-chunk from `FSChunkStore` using the file’s manifest (ordered by `idx`).
- **Download controller & route**: `GET /files/download/:id`
  - Sets `Content-Type: application/octet-stream`
  - Sets `Content-Length` from `files.total_size`
  - Sets `Content-Disposition` with original filename for proper downloads
- **Auth middleware**: accepts API key via `X-API-Key` **or** `?api_key=` query param. Reads expected key from `MIDDLEWARE_KEY` env.
- **Config guard**: safe fallback buffer size for streaming if `helper.StreamByteSize` is misconfigured.

### Changed
- **Transaction flow (download)**: DB reads happen inside a tx; tx is **committed before streaming** to avoid holding DB resources during long I/O.
- **Error mapping**: clearer separation between `ErrNotFound` (missing file) vs `ErrInternal` (DB/FS failures).
- **Sanity validation**: verify manifest byte total equals `files.total_size` before streaming.

### Notes
- Core features **Upload + Download** are now functional.
- Next up (planned):
  - **Metadata endpoints**: `GET /files/:id` (single), `GET /files` (list), and optional `GET /files/:id/chunks`.
  - **Observability**: slog logging + Prometheus counters/histograms (duration, bytes, chunks, dedupe savings, errors by stage).
  - **Graceful shutdown** for HTTP + in-flight uploads.
  - **Containerization**: Dockerfile & (optional) Kubernetes manifests.
  - **Tiny frontend**: simple page with an upload button, table of files, and per-row download button.

---

## [0.1.0] - 2025-09-04
### Added
- Initial project setup for **ByteSize**.
- Domain models: `Chunk`, `File`, and `FileChunk`.
- PostgreSQL repositories with validation and error handling:
  - `ChunkRepository` (with deduplicating upsert).
  - `FileRepository` (create, list, find).
  - `FileChunkRepository` (batch insert and manifest queries).
- `FSChunkStore` for chunk persistence on disk.
- **Upload pipeline** with concurrent goroutines:
  - Chunking, hashing, storing, and deduplication.
  - Manifest batching with DB commits.
  - File totals update after successful upload.
- REST API controller for file uploads:
  - Supports multipart file uploads.
  - Enforces content-type and payload diskSize validation.
  - Returns structured JSON responses with deduplication stats.
- Configurable constants for chunk diskSize, worker pool, and batch diskSize.
- Error helper utilities for consistent JSON responses.

### Changed
- Improved error handling and rollback logic in transactional operations.

### Notes
- Current focus: **Upload & deduplication pipeline** fully functional.
- Next steps: add download endpoints, metadata listing, logging/metrics, and graceful shutdown support.
