package helper

import (
	"encoding/json"
	"errors"
	"meliocool/bytesize/internal/model/web"
	"net/http"
)

var ErrBadRequest = errors.New("invalid request")
var ErrTooLarge = errors.New("payload too large")
var ErrNotFound = errors.New("resource not found")
var ErrInvalidInput = errors.New("invalid input")
var ErrInternal = errors.New("internal server error")

func WriteErr(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	if errors.Is(err, ErrBadRequest) {
		w.WriteHeader(http.StatusBadRequest)
		encoder := json.NewEncoder(w)
		webResponse := web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "Bad Request!",
			Data:   err.Error(),
		}
		encoder.Encode(webResponse)
	} else if errors.Is(err, ErrTooLarge) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		encoder := json.NewEncoder(w)
		webResponse := web.WebResponse{
			Code:   http.StatusRequestEntityTooLarge,
			Status: "Payload Too Large!",
			Data:   err.Error(),
		}
		encoder.Encode(webResponse)
	} else if errors.Is(err, ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		encoder := json.NewEncoder(w)
		webResponse := web.WebResponse{
			Code:   http.StatusNotFound,
			Status: "Resource Not Found!",
			Data:   err.Error(),
		}
		encoder.Encode(webResponse)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		encoder := json.NewEncoder(w)
		webResponse := web.WebResponse{
			Code:       http.StatusInternalServerError,
			Status:     "Internal Server Error!",
			Data:       "internal server error",
			LimitBytes: "2 GB",
		}
		encoder.Encode(webResponse)
	}
}
