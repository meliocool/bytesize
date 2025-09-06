package controller

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type FileMetaDataController interface {
	Get(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
}
