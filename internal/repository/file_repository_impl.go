package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/model/domain"
)

type FileRepositoryImpl struct {
}

func NewFileRepositoryImpl() FileRepository {
	return &FileRepositoryImpl{}
}

func (f *FileRepositoryImpl) Create(ctx context.Context, tx pgx.Tx, file domain.File) (domain.File, error) {
	if file.TotalSize < 0 || file.Filename == "" {
		return domain.File{}, helper.ErrInvalidInput
	}

	SQL := "INSERT INTO files (filename, total_size) VALUES($1, $2) RETURNING id, filename, total_size, created_at"

	fileRow := domain.File{}

	if err := tx.QueryRow(ctx, SQL, file.Filename, file.TotalSize).Scan(&fileRow.ID, &fileRow.Filename, &fileRow.TotalSize, &fileRow.CreatedAt); err == nil {
		return fileRow, nil
	} else {
		return domain.File{}, err
	}
}

func (f *FileRepositoryImpl) FindByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (domain.File, error) {
	if id == uuid.Nil {
		return domain.File{}, helper.ErrInvalidInput
	}

	SQL := "SELECT id, filename, total_size, created_at FROM files WHERE id = $1"

	fileRow := domain.File{}

	if err := tx.QueryRow(ctx, SQL, id).Scan(&fileRow.ID, &fileRow.Filename, &fileRow.TotalSize, &fileRow.CreatedAt); err == nil {
		return fileRow, nil
	} else if errors.Is(err, pgx.ErrNoRows) {
		return domain.File{}, helper.ErrNotFound
	} else {
		return domain.File{}, err
	}
}

func (f *FileRepositoryImpl) List(ctx context.Context, tx pgx.Tx) ([]domain.File, error) {
	SQL := "SELECT id, filename, total_size, created_at FROM files ORDER BY created_at DESC"
	rows, err := tx.Query(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fileRows []domain.File
	for rows.Next() {
		file := domain.File{}
		err := rows.Scan(&file.ID, &file.Filename, &file.TotalSize, &file.CreatedAt)
		if err != nil {
			return nil, err
		}
		fileRows = append(fileRows, file)
	}
	if rowErr := rows.Err(); rowErr != nil {
		return nil, rowErr
	}
	return fileRows, nil
}
