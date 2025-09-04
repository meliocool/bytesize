# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

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
    - Enforces content-type and payload size validation.
    - Returns structured JSON responses with deduplication stats.
- Configurable constants for chunk size, worker pool, and batch size.
- Error helper utilities for consistent JSON responses.

### Changed
- Improved error handling and rollback logic in transactional operations.

### Notes
- Current focus: **Upload & deduplication pipeline** fully functional.
- Next steps: add download endpoints, metadata listing, logging/metrics, and graceful shutdown support.
