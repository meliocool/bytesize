package filemeta

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/repository"
)

type FileMetaDataServiceImpl struct {
	FileRepository      repository.FileRepository
	FileChunkRepository repository.FileChunkRepository
	DB                  *pgxpool.Pool
}

func NewFileMetaDataService(fileRepository repository.FileRepository, fileChunkRepository repository.FileChunkRepository, db *pgxpool.Pool) FileMetaDataService {
	return &FileMetaDataServiceImpl{
		FileRepository:      fileRepository,
		FileChunkRepository: fileChunkRepository,
		DB:                  db,
	}
}

func (f *FileMetaDataServiceImpl) GetMeta(ctx context.Context, fileID uuid.UUID) (MetaDataDTO, error) {
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
			return MetaDataDTO{}, helper.ErrNotFound
		}
		_ = tx.Rollback(ctx)
		return MetaDataDTO{}, helper.ErrInternal
	}
	manifest, err := f.FileChunkRepository.FindByFileID(ctx, tx, fileID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return MetaDataDTO{}, helper.ErrInternal
	}

	chunksCount := int64(len(manifest))
	if commErr := tx.Commit(ctx); commErr != nil {
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
	return MetaData, nil
}
