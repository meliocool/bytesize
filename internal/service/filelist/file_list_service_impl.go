package filelist

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/repository"
	"time"
)

type FileDTO struct {
	ID        uuid.UUID
	Filename  string
	TotalSize int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileListServiceImpl struct {
	FileRepository repository.FileRepository
	DB             *pgxpool.Pool
}

func NewFileListService(fileRepo repository.FileRepository, db *pgxpool.Pool) FileListService {
	return &FileListServiceImpl{
		FileRepository: fileRepo,
		DB:             db,
	}
}

func (s *FileListServiceImpl) List(ctx context.Context) ([]FileDTO, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, helper.ErrInternal
	}
	files, qerr := s.FileRepository.List(ctx, tx)
	if qerr != nil {
		_ = tx.Rollback(ctx)
		return nil, helper.ErrInternal
	}
	if cerr := tx.Commit(ctx); cerr != nil {
		return nil, helper.ErrInternal
	}

	out := make([]FileDTO, 0, len(files))
	for _, f := range files {
		out = append(out, FileDTO{
			ID:        f.ID,
			Filename:  f.Filename,
			TotalSize: f.TotalSize,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
		})
	}
	return out, nil
}
