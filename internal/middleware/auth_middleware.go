package middleware

import (
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/model/web"
	"net/http"
	"os"
)

type AuthMiddleware struct {
	Handler http.Handler
}

func NewAuthMiddleware(handler http.Handler) *AuthMiddleware {
	return &AuthMiddleware{Handler: handler}
}

func (middleware *AuthMiddleware) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	expected := os.Getenv("MIDDLEWARE_KEY")
	apiKey := request.Header.Get("X-API-Key")
	if apiKey == "" {
		apiKey = request.URL.Query().Get("api_key")
	}

	if apiKey == expected && expected != "" {
		middleware.Handler.ServeHTTP(writer, request)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusUnauthorized)

	webResponse := web.WebResponse{
		Code:   http.StatusUnauthorized,
		Status: "UNAUTHORIZED",
	}
	helper.WriteToResponseBody(writer, webResponse)
}
