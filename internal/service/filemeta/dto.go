package filemeta

import (
	"github.com/google/uuid"
	"time"
)

type MetaDataDTO struct {
	ID          uuid.UUID
	Filename    string
	TotalSize   int64
	ChunksCount int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
