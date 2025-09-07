package delete

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/repository"
	"meliocool/bytesize/internal/storage"
)

type Result struct {
	FileID              uuid.UUID `json:"file_id"`
	OrphanChunksDeleted int64     `json:"orphan_chunks_deleted"`
	OrphanBytesDeleted  int64     `json:"orphan_bytes_deleted"`
}

type DeleteServiceImpl struct {
	FileRepo      repository.FileRepository
	FileChunkRepo repository.FileChunkRepository
	ChunkStore    storage.ChunkStore
	DB            *pgxpool.Pool
}

func NewDeleteService(fileRepo repository.FileRepository, fileChunkRepo repository.FileChunkRepository, chunkStore storage.ChunkStore, db *pgxpool.Pool) DeleteService {
	return &DeleteServiceImpl{FileRepo: fileRepo, FileChunkRepo: fileChunkRepo, ChunkStore: chunkStore, DB: db}
}

func (s *DeleteServiceImpl) Delete(ctx context.Context, id uuid.UUID) (Result, error) {
	if id == uuid.Nil {
		return Result{}, helper.ErrInvalidInput
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return Result{}, helper.ErrInternal
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = s.FileRepo.FindByID(ctx, tx, id)
	if err != nil {
		if err == helper.ErrNotFound {
			return Result{}, helper.ErrNotFound
		}
		return Result{}, helper.ErrInternal
	}
	manifest, err := s.FileChunkRepo.FindByFileID(ctx, tx, id)
	if err != nil {
		return Result{}, helper.ErrInternal
	}

	seen := make(map[string]int64)
	for _, fc := range manifest {
		if _, ok := seen[fc.ChunkHash]; !ok {
			seen[fc.ChunkHash] = fc.Size
		}
	}

	if err := s.FileRepo.Delete(ctx, tx, id); err != nil {
		if err == helper.ErrNotFound {
			return Result{}, helper.ErrNotFound
		}
		return Result{}, helper.ErrInternal
	}

	var orphanHashes []string
	var orphanBytes int64
	for h, sz := range seen {
		cmd, derr := tx.Exec(ctx, `
            DELETE FROM chunks c
            WHERE c.hash = $1
              AND NOT EXISTS (SELECT 1 FROM file_chunks fc WHERE fc.chunk_hash = $1)
        `, h)
		if derr != nil {
			return Result{}, helper.ErrInternal
		}
		if cmd.RowsAffected() > 0 {
			orphanHashes = append(orphanHashes, h)
			orphanBytes += sz
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Result{}, helper.ErrInternal
	}

	var deleted int64
	for _, h := range orphanHashes {
		if derr := s.ChunkStore.Delete(h); derr == nil {
			deleted++
		}
	}

	return Result{
		FileID:              id,
		OrphanChunksDeleted: deleted,
		OrphanBytesDeleted:  orphanBytes,
	}, nil
}
