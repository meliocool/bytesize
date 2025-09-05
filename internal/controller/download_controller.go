package controller

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type DownloadController interface {
	Download(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
}
