package storage

import "io"

type ChunkStore interface {
	Put(hash string, reader io.Reader, size int64) error
	Get(hash string) (io.ReadCloser, int64, error)
	Exists(hash string) (bool, error)
}
