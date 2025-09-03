package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/model/domain"
)

type ChunkRepositoryImpl struct {
}

func NewChunkRepository() ChunkRepository {
	return &ChunkRepositoryImpl{}
}

func (c *ChunkRepositoryImpl) Upsert(ctx context.Context, tx pgx.Tx, chunk domain.Chunk) (domain.Chunk, bool, error) {
	regex := helper.HashRegex()
	if chunk.Size <= 0 || !regex.MatchString(chunk.Hash) {
		return domain.Chunk{}, false, helper.ErrInvalidInput
	}

	var wasNew bool = false

	SQL := "INSERT INTO chunks(hash, size) VALUES($1, $2) ON CONFLICT(hash) DO NOTHING RETURNING hash, size, created_at"

	chunkRow := domain.Chunk{}

	if err := tx.QueryRow(ctx, SQL, chunk.Hash, chunk.Size).Scan(&chunkRow.Hash, &chunkRow.Size, &chunkRow.CreatedAt); err == nil {
		wasNew = true
		return chunkRow, wasNew, nil
	} else if errors.Is(err, pgx.ErrNoRows) {
		SQL = "SELECT hash, size, created_at FROM chunks WHERE hash = $1"
		if err := tx.QueryRow(ctx, SQL, chunk.Hash).Scan(&chunkRow.Hash, &chunkRow.Size, &chunkRow.CreatedAt); err == nil {
			return chunkRow, wasNew, nil
		} else {
			return domain.Chunk{}, false, err
		}
	} else {
		return domain.Chunk{}, false, err
	}
}

func (c *ChunkRepositoryImpl) FindByHash(ctx context.Context, tx pgx.Tx, hash string) (domain.Chunk, error) {
	regex := helper.HashRegex()
	if !regex.MatchString(hash) {
		return domain.Chunk{}, helper.ErrInvalidInput
	}

	SQL := "SELECT hash, size, created_at FROM chunks WHERE hash = $1"

	chunkRow := domain.Chunk{}

	if err := tx.QueryRow(ctx, SQL, hash).Scan(&chunkRow.Hash, &chunkRow.Size, &chunkRow.CreatedAt); err == nil {
		return chunkRow, nil
	} else if errors.Is(err, pgx.ErrNoRows) {
		return domain.Chunk{}, helper.ErrNotFound
	} else {
		return domain.Chunk{}, err
	}
}
