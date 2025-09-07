# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

---

## [0.4.0] - 2025-09-07
### Added
- **Files list endpoint**: `GET /files` returns all files (`id, filename, total_size, created_at, updated_at`) ordered by `created_at DESC`.
- **Delete endpoint**: `DELETE /files/del/:id`
  - Deletes the file row.
  - Cascades removal of its manifest (`file_chunks`) via FK `ON DELETE CASCADE`.
  - **GC of orphan chunks**: removes unreferenced rows from `chunks` and deletes corresponding chunk files from `FSChunkStore`.
- **Frontend (Next.js + Tailwind + shadcn/ui + Axios)**:
  - One-page dashboard with **table/grid toggle**, upload modal, per-item **Download** (and Delete action wired to new API).
  - Axios adds `X-API-Key` and `?api_key=` on every request; `/api/*` rewrite proxies to the Go backend.
- **CLI dev runner**:
  - `bytesize-cli run` prints friendly banner and URLs, launches backend + frontend, and cleans up on Ctrl+C.

### Changed
- Storage layer: added `ChunkStore.Delete(hash)` to remove chunk files (and optionally tidy empty directories).
- Router/middleware updated to register **list** and **delete** routes (and allow `DELETE`, plus `OPTIONS` if CORS path is used).

---

## [0.3.0] - 2025-09-06
### Added
- **Metadata service & controller**: `GET /files/metadata/:id` returns `id, filename, total_size, chunks_count, created_at, updated_at`.
- **Structured logging (slog)** across Upload, Download, and Metadata:
  - `*_start`, `*_ok` (with `took`), and `*_err` with minimal `stage` context.
- **Prometheus metrics** and `/metrics` endpoint:
  - `bytesize_requests_total{endpoint}`
  - `bytesize_errors_total{endpoint}`
  - `bytesize_request_duration_seconds{endpoint}`
  - `bytesize_bytes_uploaded_total`
  - `bytesize_bytes_streamed_total`
- **Auth on metrics** (if middleware is global): `/metrics` also accepts `X-API-Key` or `?api_key=`.

### Changed
- Services now accept a shared `*slog.Logger` and emit minimal structured logs.
- Instrumented services with counters/histograms for basic observability.

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
