package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FSChunkStore struct {
	BaseDir string
}

func NewFSChunkStore(baseDir string) *FSChunkStore {
	return &FSChunkStore{BaseDir: baseDir}
}

func (s *FSChunkStore) Put(hash string, reader io.Reader, size int64) error {
	if len(hash) != 64 {
		return fmt.Errorf("invalid hash length: %d", len(hash))
	}
	finalPath := s.pathFromHash(hash)

	if ok, _ := s.Exists(hash); ok {
		_, _ = io.Copy(io.Discard, reader)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(finalPath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(finalPath), hash+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tempName := tempFile.Name()
	defer func() { _ = os.Remove(tempName) }()

	var written int64
	if size >= 0 {
		written, err = io.Copy(tempFile, io.LimitReader(reader, size))
		if err == nil {
			_, _ = io.Copy(io.Discard, reader)
		}
	} else {
		written, err = io.Copy(tempFile, reader)
	}
	if err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("copy to temp: %w", err)
	}
	if size >= 0 && written != size {
		_ = tempFile.Close()
		return fmt.Errorf("short write: expected %d, got %d", size, written)
	}

	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("sync temp: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	if _, err := os.Stat(finalPath); err == nil {
		return nil
	}

	if err := os.Rename(tempName, finalPath); err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}

func (s *FSChunkStore) Get(hash string) (io.ReadCloser, int64, error) {
	if len(hash) != 64 {
		return nil, 0, fmt.Errorf("invalid hash length: %d", len(hash))
	}
	finalPath := s.pathFromHash(hash)

	f, err := os.Open(finalPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, 0, fmt.Errorf("chunk not found")
		}
		return nil, 0, fmt.Errorf("open: %w", err)
	}

	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, 0, fmt.Errorf("stat: %w", err)
	}

	return f, stat.Size(), nil
}

func (s *FSChunkStore) Exists(hash string) (bool, error) {
	if len(hash) != 64 {
		return false, fmt.Errorf("invalid hash length: %d", len(hash))
	}
	_, err := os.Stat(s.pathFromHash(hash))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (s *FSChunkStore) pathFromHash(hash string) string {
	return filepath.Join(s.BaseDir, hash[0:2], hash[2:4], hash)
}

func (s *FSChunkStore) Delete(hash string) error {
	path := s.pathFromHash(hash)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	_ = os.Remove(filepath.Dir(path))
	_ = os.Remove(filepath.Dir(filepath.Dir(path)))
	return nil
}
