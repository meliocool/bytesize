package domain

import (
	"github.com/google/uuid"
	"time"
)

type File struct {
	ID        uuid.UUID
	Filename  string
	TotalSize int64
	CreatedAt time.Time
}
