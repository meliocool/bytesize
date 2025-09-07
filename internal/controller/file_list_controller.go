package controller

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type FileListController interface {
	List(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
}
