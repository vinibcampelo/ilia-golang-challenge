package httpapi

import (
	"net/http"

	"ilia-golang-challenge/ms-users/internal/infrastructure/swagger"
)

// NewRouter registers Swagger (when openAPISpec is non-empty), then user routes.
func NewRouter(userHandler *UserHandler, openAPISpec []byte) http.Handler {
	mux := http.NewServeMux()
	swagger.Register(mux, openAPISpec)
	mux.HandleFunc("/users", userHandler.PostUser)
	return mux
}
