package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type subjectContextKeyType struct{}

var subjectContextKey = subjectContextKeyType{}

func WithSubject(ctx context.Context, subject string) context.Context {
	return context.WithValue(ctx, subjectContextKey, subject)
}

func SubjectFromContext(ctx context.Context) (string, bool) {
	subject, ok := ctx.Value(subjectContextKey).(string)
	return subject, ok && subject != ""
}

const (
	unauthorizedDescription = "Access token is missing or invalid"
	forbiddenDescription    = "forbidden"
)

func writeUnauthorized(w http.ResponseWriter) {
	http.Error(w, unauthorizedDescription, http.StatusUnauthorized)
}

func writeForbidden(w http.ResponseWriter) {
	http.Error(w, forbiddenDescription, http.StatusForbidden)
}

func bearerTokenFromAuthorizationHeader(rawHeader string) (string, bool) {
	rawAuthorization := strings.TrimSpace(rawHeader)
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(rawAuthorization, bearerPrefix) {
		return "", false
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(rawAuthorization, bearerPrefix))
	if tokenString == "" {
		return "", false
	}
	return tokenString, true
}

func JWTBearerMiddleware(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(secret) == 0 {
			writeUnauthorized(w)
			return
		}
		tokenString, ok := bearerTokenFromAuthorizationHeader(r.Header.Get("Authorization"))
		if !ok {
			writeUnauthorized(w)
			return
		}
		parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		claims := &jwt.RegisteredClaims{}
		token, err := parser.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return secret, nil
		})
		if err != nil || !token.Valid {
			writeUnauthorized(w)
			return
		}
		if claims.Subject == "" {
			writeUnauthorized(w)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithSubject(r.Context(), claims.Subject)))
	})
}
