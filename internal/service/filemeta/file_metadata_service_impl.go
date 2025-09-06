package filemeta

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/repository"
	"time"
)

type FileMetaDataServiceImpl struct {
	FileRepository      repository.FileRepository
	FileChunkRepository repository.FileChunkRepository
	DB                  *pgxpool.Pool
	Logger              *slog.Logger
}

func NewFileMetaDataService(fileRepository repository.FileRepository, fileChunkRepository repository.FileChunkRepository, db *pgxpool.Pool, logger *slog.Logger) FileMetaDataService {
	return &FileMetaDataServiceImpl{
		FileRepository:      fileRepository,
		FileChunkRepository: fileChunkRepository,
		DB:                  db,
		Logger:              logger,
	}
}

func (f *FileMetaDataServiceImpl) GetMeta(ctx context.Context, fileID uuid.UUID) (MetaDataDTO, error) {
	start := time.Now()
	f.Logger.Info("meta_start", slog.String("file_id", fileID.String()))

	if fileID == uuid.Nil {
		return MetaDataDTO{}, helper.ErrInvalidInput
	}

	tx, dbErr := f.DB.Begin(ctx)
	if dbErr != nil {
		return MetaDataDTO{}, helper.ErrInternal
	}
	fileRow, err := f.FileRepository.FindByID(ctx, tx, fileID)
	if err != nil {
		if errors.Is(err, helper.ErrNotFound) {
			_ = tx.Rollback(ctx)
			f.Logger.Error("meta_err", slog.String("stage", "find_file"), slog.String("file_id", fileID.String()), slog.Any("err", err))
			return MetaDataDTO{}, helper.ErrNotFound
		}
		_ = tx.Rollback(ctx)
		return MetaDataDTO{}, helper.ErrInternal
	}
	manifest, err := f.FileChunkRepository.FindByFileID(ctx, tx, fileID)
	if err != nil {
		_ = tx.Rollback(ctx)
		f.Logger.Error("meta_err", slog.String("stage", "find_manifest"), slog.String("file_id", fileID.String()), slog.Any("err", err))
		return MetaDataDTO{}, helper.ErrInternal
	}

	chunksCount := int64(len(manifest))
	if commErr := tx.Commit(ctx); commErr != nil {
		f.Logger.Error("meta_err", slog.String("stage", "commit"), slog.String("file_id", fileID.String()), slog.Any("err", commErr))
		return MetaDataDTO{}, helper.ErrInternal
	}

	MetaData := MetaDataDTO{
		ID:          fileRow.ID,
		Filename:    fileRow.Filename,
		TotalSize:   fileRow.TotalSize,
		ChunksCount: chunksCount,
		CreatedAt:   fileRow.CreatedAt,
		UpdatedAt:   fileRow.UpdatedAt,
	}

	f.Logger.Info(
		"meta_ok",
		slog.String("file_id", fileID.String()),
		slog.Int64("chunks_count", chunksCount),
		slog.Int64("total_size", fileRow.TotalSize),
		slog.Duration("took", time.Since(start)),
	)

	return MetaData, nil
}
