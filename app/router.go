package app

import (
	"github.com/julienschmidt/httprouter"
	"meliocool/bytesize/internal/controller"
)

func NewRouter(uploadController controller.UploadController) *httprouter.Router {
	router := httprouter.New()

	router.POST("/upload", uploadController.Upload)

	return router
}
