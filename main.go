package main

import (
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"meliocool/bytesize/app"
	"meliocool/bytesize/internal/controller"
	"meliocool/bytesize/internal/middleware"
	"meliocool/bytesize/internal/repository"
	"meliocool/bytesize/internal/service/download"
	"meliocool/bytesize/internal/service/upload"
	"meliocool/bytesize/internal/storage"
	"net/http"
	"os"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("failed to load .env: " + err.Error())
	}
	db := app.NewDB()

	validate := validator.New()

	chunkRepository := repository.NewChunkRepository()
	fileRepository := repository.NewFileRepository()
	fileChunksRepository := repository.NewFileChunksRepository()
	chunkStorage := storage.NewFSChunkStore(os.Getenv("BASE_DIR"))

	uploadService := upload.NewUploadService(chunkRepository, fileRepository, fileChunksRepository, chunkStorage, db, validate)
	uploadController := controller.NewUploadController(uploadService)

	downloadService := download.NewDownloadService(fileRepository, fileChunksRepository, chunkStorage, db)
	downloadController := controller.NewDownloadController(downloadService, fileRepository, db)

	router := httprouter.New()

	router.POST("/files/upload", uploadController.Upload)
	router.GET("/files/download/:id", downloadController.Download)

	server := http.Server{
		Addr:    "localhost:8080",
		Handler: middleware.NewAuthMiddleware(router),
	}

	err := server.ListenAndServe()
	if err != nil {
		panic("Server Stopped Abruptly!")
	}
}
