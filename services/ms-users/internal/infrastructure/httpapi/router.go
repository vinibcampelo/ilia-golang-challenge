package httpapi

import (
	"net/http"

	"ilia-golang-challenge/services/ms-users/internal/infrastructure/swagger"
)

func NewRouter(
	userHandler *UserHandler,
	authHandler *AuthHandler,
	internalUserHandler *InternalUserHandler,
	jwtSecret []byte,
	internalJWTSecret []byte,
	openAPISpec []byte,
) http.Handler {
	mux := http.NewServeMux()
	swagger.Register(mux, openAPISpec)

	mux.Handle("POST /users", http.HandlerFunc(userHandler.PostUser))
	mux.Handle("POST /auth", http.HandlerFunc(authHandler.PostAuth))

	withJWT := func(next http.Handler) http.Handler {
		return JWTBearerMiddleware(jwtSecret, next)
	}
	selfOnly := RequirePathSubjectMatchesJWT("id")

	mux.Handle("GET /users", withJWT(http.HandlerFunc(userHandler.GetUsers)))
	mux.Handle("GET /users/{id}", withJWT(selfOnly(http.HandlerFunc(userHandler.GetUser))))
	mux.Handle("PATCH /users/{id}", withJWT(selfOnly(http.HandlerFunc(userHandler.PatchUser))))
	mux.Handle("DELETE /users/{id}", withJWT(selfOnly(http.HandlerFunc(userHandler.DeleteUser))))

	internalUsers := InternalJWTBearerMiddleware(
		internalJWTSecret,
		InternalAudienceUsers,
		InternalCallerTransactions,
		http.HandlerFunc(internalUserHandler.GetUserInternal),
	)
	mux.Handle("GET /internal/users/{id}", internalUsers)

	return mux
}
