package controller

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type DeleteController interface {
	Delete(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
}
