package walletclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"ilia-golang-challenge/services/ms-users/internal/application/user/usecase"
	"ilia-golang-challenge/services/ms-users/internal/infrastructure/httpapi"
)

const internalTokenTTL = 60 * time.Second

type HTTPWalletZeroBalanceChecker struct {
	baseURL    string
	secret     []byte
	httpClient *http.Client
}

func NewHTTPWalletZeroBalanceChecker(baseURL string, jwtInternalSecret []byte, clientTimeout time.Duration) *HTTPWalletZeroBalanceChecker {
	return &HTTPWalletZeroBalanceChecker{
		baseURL: strings.TrimRight(baseURL, "/"),
		secret:  jwtInternalSecret,
		httpClient: &http.Client{
			Timeout: clientTimeout,
		},
	}
}

func (c *HTTPWalletZeroBalanceChecker) AssertZeroBalance(ctx context.Context, userID string) error {
	token, err := mintInternalToken(c.secret, httpapi.InternalCallerUsers, httpapi.InternalAudienceTransactions)
	if err != nil {
		return fmt.Errorf("wallet client: %w", usecase.ErrWalletServiceUnavailable)
	}
	url := fmt.Sprintf("%s/internal/wallet/%s/balance", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("wallet client: %w", usecase.ErrWalletServiceUnavailable)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("wallet client: %w", usecase.ErrWalletServiceUnavailable)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("wallet client: %w", usecase.ErrWalletServiceUnavailable)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wallet client: %w", usecase.ErrWalletServiceUnavailable)
	}

	var body struct {
		Balance int64 `json:"balance"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return fmt.Errorf("wallet client: %w", usecase.ErrWalletServiceUnavailable)
	}
	if body.Balance != 0 {
		return usecase.ErrWalletHasNonZeroBalance
	}
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
	s, err := tok.SignedString(secret)
	if err != nil {
		return "", err
	}
	return s, nil
}

var _ usecase.WalletZeroBalanceChecker = (*HTTPWalletZeroBalanceChecker)(nil)
