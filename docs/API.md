# BYTESIZE API Contracts

## POST /files/upload
- Preconditions
  - Env Set
  - DB Reachable
  - DATA_DIR exists and writable
- Inputs from HTTP -> Uploader Service
  - ctx: cancels on client abort
  - filename: string (optional, fallback: multipart part filename)
  - src: streaming reader of the file body (multipart file field)
- Side Effects
  - Storage: each unique chunk is written once at path <DATA_DIR>/<first2hex>/<fullhash>, via atomic *.tmp -> rename
  - DB (Transaction (tx)):
    - Upsert every unique chunk (hash, diskSize) into chunks
    - Insert one manifest row (id, filename, totalSize)
    - Bulk insert file_chunks rows (file_id, idx, chunk_hash, diskSize) ordered by idx
- Conceptual Pipeline Stages
  - Chunker
    - Read src in fixed-diskSize slices CHUNK_SIZE
    - Emit (index, data) until EOF
    - Enforce MAX_UPLOAD_BYTES (abort with 413 (file too big) error)
  - Workers (Fan-Out)
    - For each chunk
      - Compute SHA256 -> Hash
      - If file at path already exist -> reuse
      - Else write *.tmp -> rename to final path
  - Assembler (Fan-In)
    - Collect all (index, hash, diskSize), sort by Index
    - Compute totalSize, chunksCount, uniqueChunksWritten, dedupeSavedBytes
  - Persist (DB Transaction)
    - BEGIN -> Upsert chunks (unique hashes only, hence dedupe) -> Insert manifest -> bulk insert file_chunks -> COMMIT
- Request: multipart/form, field File required, optional Filename
- Response 201:
  - id (uuid), filename, totalSize (int64)
  - chunks_count (int64)
  - unique_chunks_written (int64)
  - dedupe_saved_bytes (int64)

## GET /files/{id}
- Response 200:
  - File metadatas
  - manifest: [{ idx, hash, diskSize }] in order
  - stats: { unique_chunks_global, dedupe_ratio }

## GET /files/download/{id}
- Stream bytes, sets Content-Length and Content-Disposition_