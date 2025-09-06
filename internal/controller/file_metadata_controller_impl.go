package controller

import (
	"errors"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/service/filemeta"
	"net/http"
)

type FileMetaDataControllerImpl struct {
	FileMetaDataService filemeta.FileMetaDataService
}

func NewFileMetaDataController(FileMetaDataService filemeta.FileMetaDataService) FileMetaDataController {
	return &FileMetaDataControllerImpl{
		FileMetaDataService: FileMetaDataService,
	}
}

func (f *FileMetaDataControllerImpl) Get(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	ctx := request.Context()
	id := params.ByName("id")
	fileID, err := uuid.Parse(id)
	if err != nil {
		helper.WriteErr(writer, helper.ErrInvalidInput)
		return
	}

	DTO, getMetaErr := f.FileMetaDataService.GetMeta(ctx, fileID)
	if getMetaErr != nil {
		if errors.Is(getMetaErr, helper.ErrInvalidInput) {
			helper.WriteErr(writer, helper.ErrInvalidInput)
			return
		} else if errors.Is(getMetaErr, helper.ErrNotFound) {
			helper.WriteErr(writer, helper.ErrNotFound)
			return
		}
		helper.WriteErr(writer, helper.ErrInternal)
		return
	}

	// * Already sets Content-Type to application/json
	helper.WriteToResponseBody(writer, DTO)
}
