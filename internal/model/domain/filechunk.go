package domain

import "github.com/google/uuid"

type FileChunk struct {
	FileID    uuid.UUID
	Idx       int64
	ChunkHash string
	Size      int64
}
