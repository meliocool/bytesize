package download

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"log/slog"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/metrics"
	"meliocool/bytesize/internal/repository"
	"meliocool/bytesize/internal/storage"
	"time"
)

type DownloadServiceImpl struct {
	FileRepository      repository.FileRepository
	FileChunkRepository repository.FileChunkRepository
	ChunkStore          storage.ChunkStore
	DB                  *pgxpool.Pool
	Logger              *slog.Logger
}

func NewDownloadService(fileRepository repository.FileRepository, fileChunkRepository repository.FileChunkRepository, chunkStore storage.ChunkStore, db *pgxpool.Pool, logger *slog.Logger) DownloadService {
	return &DownloadServiceImpl{
		FileRepository:      fileRepository,
		FileChunkRepository: fileChunkRepository,
		ChunkStore:          chunkStore,
		DB:                  db,
		Logger:              logger,
	}
}

func (d *DownloadServiceImpl) Stream(ctx context.Context, fileID uuid.UUID, w io.Writer) error {
	start := time.Now()
	metrics.RequestsTotal.WithLabelValues("download").Inc()
	defer metrics.RequestDuration.WithLabelValues("download").Observe(time.Since(start).Seconds())
	d.Logger.Info("download_start", slog.String("file_id", fileID.String()))

	if fileID == uuid.Nil {
		d.Logger.Error("download_err", slog.String("stage", "validate"), slog.String("file_id", fileID.String()), slog.String("reason", "nil_uuid"))
		metrics.ErrorsTotal.WithLabelValues("download").Inc()
		return helper.ErrInvalidInput
	}

	tx, err := d.DB.Begin(ctx)
	if err != nil {
		d.Logger.Error("download_err", slog.String("stage", "db_begin"), slog.String("file_id", fileID.String()), slog.Any("err", err))
		metrics.ErrorsTotal.WithLabelValues("download").Inc()
		return err
	}

	fileRow, fileRowErr := d.FileRepository.FindByID(ctx, tx, fileID)
	if fileRowErr != nil {
		_ = tx.Rollback(ctx)
		metrics.ErrorsTotal.WithLabelValues("download").Inc()
		return helper.ErrNotFound
	}

	totalSize := fileRow.TotalSize

	manifest, filesErr := d.FileChunkRepository.FindByFileID(ctx, tx, fileID)
	if filesErr != nil {
		_ = tx.Rollback(ctx)
		d.Logger.Error("download_err", slog.String("stage", "find_manifest"), slog.String("file_id", fileID.String()), slog.Any("err", filesErr))
		metrics.ErrorsTotal.WithLabelValues("download").Inc()
		return helper.ErrInternal
	}

	if len(manifest) == 0 && totalSize == 0 {
		commErr := tx.Commit(ctx)
		if commErr != nil {
			d.Logger.Error("download_err", slog.String("stage", "commit"), slog.String("file_id", fileID.String()), slog.Any("err", commErr))
			metrics.ErrorsTotal.WithLabelValues("download").Inc()
			return helper.ErrInternal
		}
		return nil
	}

	if len(manifest) == 0 && totalSize > 0 {
		_ = tx.Rollback(ctx)
		metrics.ErrorsTotal.WithLabelValues("download").Inc()
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
		metrics.ErrorsTotal.WithLabelValues("download").Inc()
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
			d.Logger.Error("download_err", slog.String("stage", "chunk_get"), slog.String("file_id", fileID.String()), slog.String("hash", fc.ChunkHash), slog.Any("err", getErr))
			metrics.ErrorsTotal.WithLabelValues("download").Inc()
			return helper.ErrInternal
		}
		n, copyErr := io.CopyBuffer(w, rc, buffer)
		closeErr := rc.Close()
		if closeErr != nil {
			metrics.ErrorsTotal.WithLabelValues("download").Inc()
			return helper.ErrInternal
		}
		if copyErr != nil {
			d.Logger.Error("download_err", slog.String("stage", "copy"), slog.String("file_id", fileID.String()), slog.String("hash", fc.ChunkHash), slog.Any("err", copyErr))
			metrics.ErrorsTotal.WithLabelValues("download").Inc()
			return helper.ErrInternal
		}
		if fc.Size != n {
			metrics.ErrorsTotal.WithLabelValues("download").Inc()
			return helper.ErrInternal
		}
		writtenTotal += n
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	if writtenTotal != totalSize {
		metrics.ErrorsTotal.WithLabelValues("download").Inc()
		return helper.ErrInternal
	}
	d.Logger.Info(
		"download_ok",
		slog.String("file_id", fileID.String()),
		slog.Int64("total_size", totalSize),
		slog.Duration("took", time.Since(start)),
	)

	metrics.BytesStreamedTotal.Add(float64(writtenTotal))

	return nil
}
