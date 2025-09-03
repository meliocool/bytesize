package web

import "github.com/google/uuid"

type UploadResponse struct {
	FileID              uuid.UUID
	TotalSize           int64
	ChunksCount         int64
	UniqueChunksWritten int64
	DedupeSavedBytes    int64
}
