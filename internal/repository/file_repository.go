package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"meliocool/bytesize/internal/model/domain"
)

type FileRepository interface {
	Create(ctx context.Context, tx pgx.Tx, file domain.File) (domain.File, error)
	FindByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (domain.File, error)
	List(ctx context.Context, tx pgx.Tx) ([]domain.File, error)
	UpdateTotals(ctx context.Context, tx pgx.Tx, id uuid.UUID, totalSize int64) error
	Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
}
