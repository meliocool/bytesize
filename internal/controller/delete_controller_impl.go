package controller

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"meliocool/bytesize/internal/helper"
	deletefile "meliocool/bytesize/internal/service/delete"
)

type DeleteControllerImpl struct {
	Svc deletefile.DeleteService
}

func NewDeleteController(svc deletefile.DeleteService) DeleteController {
	return &DeleteControllerImpl{Svc: svc}
}

func (c *DeleteControllerImpl) Delete(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	idStr := params.ByName("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		helper.WriteErr(writer, helper.ErrBadRequest)
		return
	}

	res, derr := c.Svc.Delete(request.Context(), id)
	if derr != nil {
		switch {
		case errors.Is(derr, helper.ErrInvalidInput):
			helper.WriteErr(writer, helper.ErrInvalidInput)
		case errors.Is(derr, helper.ErrNotFound):
			helper.WriteErr(writer, helper.ErrNotFound)
		default:
			helper.WriteErr(writer, helper.ErrInternal)
		}
		return
	}

	helper.WriteToResponseBody(writer, res)
}
