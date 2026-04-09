package httpapi

import (
	"net/http"

	"ilia-golang-challenge/services/ms-transactions/internal/infrastructure/swagger"
)

// NewRouter registers Swagger (when openAPISpec is non-empty) and JWT-protected transaction routes.
func NewRouter(handler *TransactionHandler, jwtSecret []byte, openAPISpec []byte) http.Handler {
	mux := http.NewServeMux()
	swagger.Register(mux, openAPISpec)

	withJWT := func(next http.Handler) http.Handler {
		return JWTBearerMiddleware(jwtSecret, next)
	}

	mux.Handle("POST /transactions", withJWT(http.HandlerFunc(handler.PostTransaction)))
	mux.Handle("GET /transactions", withJWT(http.HandlerFunc(handler.GetTransactions)))
	mux.Handle("GET /balance", withJWT(http.HandlerFunc(handler.GetBalance)))

	return mux
}
