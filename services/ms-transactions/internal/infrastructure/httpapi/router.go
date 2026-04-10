package httpapi

import (
	"net/http"

	"ilia-golang-challenge/services/ms-transactions/internal/infrastructure/swagger"
)

func NewRouter(
	handler *TransactionHandler,
	internalWalletHandler *InternalWalletHandler,
	jwtSecret []byte,
	internalJWTSecret []byte,
	openAPISpec []byte,
) http.Handler {
	mux := http.NewServeMux()
	swagger.Register(mux, openAPISpec)

	withJWT := func(next http.Handler) http.Handler {
		return JWTBearerMiddleware(jwtSecret, next)
	}

	mux.Handle("POST /transactions", withJWT(http.HandlerFunc(handler.PostTransaction)))
	mux.Handle("GET /transactions", withJWT(http.HandlerFunc(handler.GetTransactions)))
	mux.Handle("GET /balance", withJWT(http.HandlerFunc(handler.GetBalance)))

	internalBalance := InternalJWTBearerMiddleware(
		internalJWTSecret,
		InternalAudienceTransactions,
		InternalCallerUsers,
		http.HandlerFunc(internalWalletHandler.GetBalanceInternal),
	)
	mux.Handle("GET /internal/wallet/{userId}/balance", internalBalance)

	return mux
}
