package jwtissuer

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// HS256AccessTokenIssuer issues JWT access tokens with HS256; subject claim is the user id.
type HS256AccessTokenIssuer struct {
	Secret []byte
	TTL    time.Duration
}

func (issuer *HS256AccessTokenIssuer) IssueAccessToken(subjectUserID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   subjectUserID,
		ExpiresAt: jwt.NewNumericDate(now.Add(issuer.TTL)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(issuer.Secret)
}
