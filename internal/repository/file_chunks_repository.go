package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"meliocool/bytesize/internal/model/domain"
)

type FileChunkRepository interface {
	AddChunks(ctx context.Context, tx pgx.Tx, fileID uuid.UUID, chunks []domain.FileChunk) error
	FindByFileID(ctx context.Context, tx pgx.Tx, fileID uuid.UUID) ([]domain.FileChunk, error)
}
