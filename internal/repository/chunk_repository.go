package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"meliocool/bytesize/internal/model/domain"
)

type ChunkRepository interface {
	Upsert(ctx context.Context, tx pgx.Tx, chunk domain.Chunk) (domain.Chunk, bool, error)
	FindByHash(ctx context.Context, tx pgx.Tx, hash string) (domain.Chunk, error)
}
