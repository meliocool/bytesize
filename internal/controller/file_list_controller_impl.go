package controller

import (
	"github.com/julienschmidt/httprouter"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/service/filelist"
	"net/http"
)

type FileListControllerImpl struct {
	Service filelist.FileListService
}

func NewFileListController(svc filelist.FileListService) FileListController {
	return &FileListControllerImpl{Service: svc}
}

func (c *FileListControllerImpl) List(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	ctx := request.Context()
	items, err := c.Service.List(ctx)
	if err != nil {
		helper.WriteErr(writer, helper.ErrInternal)
		return
	}
	helper.WriteToResponseBody(writer, items)
}
