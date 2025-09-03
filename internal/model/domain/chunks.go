package domain

import "time"

type Chunk struct {
	Hash      string
	Size      int64
	CreatedAt time.Time
}
