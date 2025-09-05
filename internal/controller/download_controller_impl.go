package controller

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/repository"
	"meliocool/bytesize/internal/service/download"
	"net/http"
	"strconv"
)

type DownloadControllerImpl struct {
	DownloadService download.DownloadService
	FileRepository  repository.FileRepository
	DB              *pgxpool.Pool
}

func NewDownloadController(downloadService download.DownloadService, fileRepository repository.FileRepository, db *pgxpool.Pool) DownloadController {
	return &DownloadControllerImpl{
		DownloadService: downloadService,
		FileRepository:  fileRepository,
		DB:              db,
	}
}

func (d *DownloadControllerImpl) Download(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	ctx := request.Context()

	id := params.ByName("id")
	fileID, err := uuid.Parse(id)
	if err != nil {
		helper.WriteErr(writer, helper.ErrBadRequest)
		return
	}

	tx, dbErr := d.DB.Begin(ctx)
	if dbErr != nil {
		helper.WriteErr(writer, helper.ErrInternal)
		return
	}

	fileRow, fileRowErr := d.FileRepository.FindByID(ctx, tx, fileID)
	if errors.Is(fileRowErr, helper.ErrNotFound) {
		_ = tx.Rollback(ctx)
		helper.WriteErr(writer, helper.ErrNotFound)
		return
	}
	if fileRowErr != nil {
		_ = tx.Rollback(ctx)
		helper.WriteErr(writer, helper.ErrInternal)
		return
	}

	totalSize := fileRow.TotalSize
	fileName := fileRow.Filename
	commErr := tx.Commit(ctx)
	if commErr != nil {
		helper.WriteErr(writer, helper.ErrInternal)
		return
	}

	writer.Header().Set("Content-Type", "application/octet-stream")
	writer.Header().Set("Content-Length", strconv.FormatInt(totalSize, 10))
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	downloadErr := d.DownloadService.Stream(ctx, fileID, writer)
	if downloadErr != nil {
		return
	}
}
