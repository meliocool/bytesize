package controller

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/model/web"
	"meliocool/bytesize/internal/service/upload"
	"net/http"
	"strings"
)

type UploadControllerImpl struct {
	UploadService upload.UploadService
}

func NewUploadController(uploadService upload.UploadService) UploadController {
	return &UploadControllerImpl{
		UploadService: uploadService,
	}
}

func (u *UploadControllerImpl) Upload(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	request.Body = http.MaxBytesReader(writer, request.Body, helper.MaxBytes)

	if !strings.HasPrefix(request.Header.Get("Content-Type"), "multipart/form-data") {
		helper.WriteErr(writer, helper.ErrUnsupportedMediaType)
		return
	}
	err := request.ParseMultipartForm(helper.MaxMemoryBytes)
	if err != nil {
		if _, ok := err.(*http.MaxBytesError); ok {
			helper.WriteErr(writer, helper.ErrTooLarge)
			return
		}
		helper.WriteErr(writer, helper.ErrBadRequest)
		return
	}
	defer request.MultipartForm.RemoveAll()

	fileReader, fileHeader, reqErr := request.FormFile("file")
	if reqErr != nil {
		helper.WriteErr(writer, helper.ErrBadRequest)
		return
	}
	defer fileReader.Close()

	formFieldName := request.FormValue("filename")
	if formFieldName == "" {
		formFieldName = fileHeader.Filename
		if formFieldName == "" {
			helper.WriteErr(writer, helper.ErrBadRequest)
			return
		}
	}

	uploadReq := web.UploadRequest{
		Ctx:      request.Context(),
		FileName: formFieldName,
		Reader:   fileReader,
	}

	resp, uploadErr := u.UploadService.Upload(request.Context(), uploadReq)
	if uploadErr != nil {
		if errors.Is(uploadErr, helper.ErrInvalidInput) {
			helper.WriteErr(writer, helper.ErrBadRequest)
			return
		} else if errors.Is(uploadErr, helper.ErrNotFound) {
			helper.WriteErr(writer, helper.ErrNotFound)
			return
		} else {
			helper.WriteErr(writer, helper.ErrInternal)
			return
		}
	}

	webResponse := web.WebResponse{
		Code:   http.StatusCreated,
		Status: "Success!",
		Data:   resp,
	}

	helper.WriteToResponseBody(writer, webResponse)
}
