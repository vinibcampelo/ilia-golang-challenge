package userdirectory

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"ilia-golang-challenge/services/ms-transactions/internal/application/transaction/usecase"
	"ilia-golang-challenge/services/ms-transactions/internal/infrastructure/httpapi"
)

const internalTokenTTL = 60 * time.Second

type HTTPUserActiveGate struct {
	baseURL    string
	secret     []byte
	httpClient *http.Client
}

func NewHTTPUserActiveGate(baseURL string, jwtInternalSecret []byte, clientTimeout time.Duration) *HTTPUserActiveGate {
	return &HTTPUserActiveGate{
		baseURL: strings.TrimRight(baseURL, "/"),
		secret:  jwtInternalSecret,
		httpClient: &http.Client{
			Timeout: clientTimeout,
		},
	}
}

func (g *HTTPUserActiveGate) EnsureActive(ctx context.Context, userID string) error {
	token, err := mintInternalToken(g.secret, httpapi.InternalCallerTransactions, httpapi.InternalAudienceUsers)
	if err != nil {
		return fmt.Errorf("users client: %w", usecase.ErrUsersServiceUnavailable)
	}
	url := fmt.Sprintf("%s/internal/users/%s", g.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("users client: %w", usecase.ErrUsersServiceUnavailable)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("users client: %w", usecase.ErrUsersServiceUnavailable)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return usecase.ErrUserNotActive
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("users client: %w", usecase.ErrUsersServiceUnavailable)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("users client: %w", usecase.ErrUsersServiceUnavailable)
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	return nil
}

func mintInternalToken(secret []byte, subject, audience string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(now.Add(internalTokenTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
		Audience:  jwt.ClaimStrings{audience},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(secret)
}

var _ usecase.UserActiveGate = (*HTTPUserActiveGate)(nil)
