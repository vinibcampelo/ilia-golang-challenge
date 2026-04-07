package httpapi

import (
	"net/http"

	"ilia-golang-challenge/ms-users/internal/infrastructure/swagger"
)

func NewRouter(userHandler *UserHandler, openAPISpec []byte) http.Handler {
	mux := http.NewServeMux()
	swagger.Register(mux, openAPISpec)

	mux.HandleFunc("GET /users", userHandler.GetUsers)
	mux.HandleFunc("POST /users", userHandler.PostUser)
	mux.HandleFunc("GET /users/{id}", userHandler.GetUser)
	mux.HandleFunc("PATCH /users/{id}", userHandler.PatchUser)
	mux.HandleFunc("DELETE /users/{id}", userHandler.DeleteUser)

	return mux
}
