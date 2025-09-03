package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/model/domain"
)

type FileChunkRepositoryImpl struct {
}

func NewFileChunksRepository() FileChunkRepository {
	return &FileChunkRepositoryImpl{}
}

func (f *FileChunkRepositoryImpl) AddChunks(ctx context.Context, tx pgx.Tx, fileID uuid.UUID, chunks []domain.FileChunk) error {
	if fileID == uuid.Nil || len(chunks) == 0 {
		return helper.ErrInvalidInput
	}
	regex := helper.HashRegex()
	for i := 0; i < len(chunks); i++ {
		if chunks[i].Idx < 0 || chunks[i].Size <= 0 || !regex.MatchString(chunks[i].ChunkHash) {
			return helper.ErrInvalidInput
		}
		if i == 0 {
			if chunks[i].Idx != 0 {
				return helper.ErrInvalidInput
			}
		} else {
			if chunks[i].Idx != chunks[i-1].Idx+1 {
				return helper.ErrInvalidInput
			}
		}
	}
	SQL := "INSERT INTO file_chunks(file_id, idx, chunk_hash, size) VALUES($1, $2, $3, $4)"
	batch := &pgx.Batch{}

	for _, chunk := range chunks {
		batch.Queue(SQL, fileID, chunk.Idx, chunk.ChunkHash, chunk.Size)
	}

	br := tx.SendBatch(ctx, batch)

	for i := 0; i < len(chunks); i++ {
		exec, err := br.Exec()
		if err != nil {
			_ = br.Close()
			return err
		}
		if exec.RowsAffected() != 1 {
			_ = br.Close()
			return helper.ErrInternal
		}
	}
	err := br.Close()
	if err != nil {
		return err
	}
	return nil
}

func (f *FileChunkRepositoryImpl) FindByFileID(ctx context.Context, tx pgx.Tx, fileID uuid.UUID) ([]domain.FileChunk, error) {
	if fileID == uuid.Nil {
		return nil, helper.ErrInvalidInput
	}

	SQL := "SELECT file_id, idx, chunk_hash, size FROM file_chunks WHERE file_id = $1 ORDER BY idx ASC"
	rows, err := tx.Query(ctx, SQL, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fileChunkRows []domain.FileChunk
	for rows.Next() {
		fileChunk := domain.FileChunk{}
		err := rows.Scan(&fileChunk.FileID, &fileChunk.Idx, &fileChunk.ChunkHash, &fileChunk.Size)
		if err != nil {
			return nil, err
		}
		fileChunkRows = append(fileChunkRows, fileChunk)
	}
	if rowErr := rows.Err(); rowErr != nil {
		return nil, rowErr
	}
	return fileChunkRows, nil
}
