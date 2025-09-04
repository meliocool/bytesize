package controller

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type UploadController interface {
	Upload(writer http.ResponseWriter, request *http.Request, _ httprouter.Params)
}
