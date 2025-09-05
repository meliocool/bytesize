package download

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/repository"
	"meliocool/bytesize/internal/storage"
)

type DownloadServiceImpl struct {
	FileRepository      repository.FileRepository
	FileChunkRepository repository.FileChunkRepository
	ChunkStore          storage.ChunkStore
	DB                  *pgxpool.Pool
}

func NewDownloadService(fileRepository repository.FileRepository, fileChunkRepository repository.FileChunkRepository, chunkStore storage.ChunkStore, db *pgxpool.Pool) DownloadService {
	return &DownloadServiceImpl{
		FileRepository:      fileRepository,
		FileChunkRepository: fileChunkRepository,
		ChunkStore:          chunkStore,
		DB:                  db,
	}
}

func (d *DownloadServiceImpl) Stream(ctx context.Context, fileID uuid.UUID, w io.Writer) error {
	if fileID == uuid.Nil {
		return helper.ErrInvalidInput
	}

	tx, err := d.DB.Begin(ctx)
	if err != nil {
		return err
	}

	fileRow, fileRowErr := d.FileRepository.FindByID(ctx, tx, fileID)
	if fileRowErr != nil {
		_ = tx.Rollback(ctx)
		return helper.ErrNotFound
	}

	totalSize := fileRow.TotalSize

	manifest, filesErr := d.FileChunkRepository.FindByFileID(ctx, tx, fileID)
	if filesErr != nil {
		_ = tx.Rollback(ctx)
		return helper.ErrInternal
	}

	if len(manifest) == 0 && totalSize == 0 {
		commErr := tx.Commit(ctx)
		if commErr != nil {
			return helper.ErrInternal
		}
		return nil
	}

	if len(manifest) == 0 && totalSize > 0 {
		_ = tx.Rollback(ctx)
		return helper.ErrInternal
	}

	var expectedBytes int64 = 0
	for _, fc := range manifest {
		expectedBytes += fc.Size
	}
	if expectedBytes != totalSize {
		_ = tx.Rollback(ctx)
		return helper.ErrInternal
	}
	commErr := tx.Commit(ctx)
	if commErr != nil {
		return helper.ErrInternal
	}

	byteSize := helper.StreamByteSize
	if byteSize <= 0 {
		byteSize = 64 * 1024
	}

	buffer := make([]byte, byteSize)
	var writtenTotal int64 = 0

	for _, fc := range manifest {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		rc, _, getErr := d.ChunkStore.Get(fc.ChunkHash)
		if getErr != nil {
			return helper.ErrInternal
		}
		n, copyErr := io.CopyBuffer(w, rc, buffer)
		closeErr := rc.Close()
		if closeErr != nil {
			return helper.ErrInternal
		}
		if copyErr != nil {
			return helper.ErrInternal
		}
		if fc.Size != n {
			return helper.ErrInternal
		}
		writtenTotal += n
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	if writtenTotal != totalSize {
		return helper.ErrInternal
	}
	return nil
}
