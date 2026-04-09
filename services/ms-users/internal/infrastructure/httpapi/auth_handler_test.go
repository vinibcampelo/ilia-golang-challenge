package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"ilia-golang-challenge/services/ms-users/internal/application/user/usecase"
	domainuser "ilia-golang-challenge/services/ms-users/internal/domain/user"
	domainusermocks "ilia-golang-challenge/services/ms-users/internal/domain/user/mocks"
	"ilia-golang-challenge/services/ms-users/internal/infrastructure/jwtissuer"
)

const (
	authHandlerTestUserID    = "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	authHandlerTestJWTSecret = "auth-handler-test-jwt-secret-min-length!!"
)

func newAuthHandlerWithRepository(tb *testing.T, repository domainuser.Repository) *AuthHandler {
	tb.Helper()
	tokenIssuer := &jwtissuer.HS256AccessTokenIssuer{
		Secret: []byte(authHandlerTestJWTSecret),
		TTL:    time.Hour,
	}
	authenticateUserUseCase := usecase.NewAuthenticateUserUseCase(repository, tokenIssuer)
	return NewAuthHandler(authenticateUserUseCase)
}

func validAuthLoginJSON() string {
	return `{"user":{"email":"ada@example.com","password":"password123"}}`
}

func TestAuthHandler_PostAuth_success(tb *testing.T) {
	tb.Parallel()

	tb.Run("returns_200_with_access_token_and_user_profile_json", func(t *testing.T) {
		t.Parallel()
		plainPassword := "password123"
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.MinCost)
		if err != nil {
			t.Fatalf("bcrypt: %v", err)
		}
		storedUser := &domainuser.User{
			ID:        authHandlerTestUserID,
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "ada@example.com",
			Password:  string(passwordHash),
		}

		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockRepository.EXPECT().FindByEmail(gomock.Any(), "ada@example.com").Return(storedUser, nil)

		handler := newAuthHandlerWithRepository(t, mockRepository)
		request := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(validAuthLoginJSON()))
		request.Header.Set("Content-Type", "application/json")
		responseRecorder := httptest.NewRecorder()
		handler.PostAuth(responseRecorder, request)

		if responseRecorder.Code != http.StatusOK {
			t.Fatalf("status: got %d want %d, body: %s", responseRecorder.Code, http.StatusOK, responseRecorder.Body.String())
		}
		if contentType := responseRecorder.Header().Get("Content-Type"); contentType != "application/json" {
			t.Fatalf("Content-Type: got %q want application/json", contentType)
		}

		var parsed struct {
			AccessToken string        `json:"access_token"`
			User        usersResponse `json:"user"`
		}
		if err := json.NewDecoder(responseRecorder.Body).Decode(&parsed); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if strings.TrimSpace(parsed.AccessToken) == "" {
			t.Fatal("success contract requires non-empty access_token")
		}
		if parsed.User.ID != authHandlerTestUserID {
			t.Fatalf("user.id: got %q want %q", parsed.User.ID, authHandlerTestUserID)
		}
		if parsed.User.FirstName != "Ada" || parsed.User.LastName != "Lovelace" || parsed.User.Email != "ada@example.com" {
			t.Fatalf("user profile: %+v", parsed.User)
		}
	})
}

func TestAuthHandler_PostAuth_errors(tb *testing.T) {
	tb.Parallel()

	wrongPasswordHash, err := bcrypt.GenerateFromPassword([]byte("correct-only"), bcrypt.MinCost)
	if err != nil {
		tb.Fatalf("bcrypt: %v", err)
	}
	repositoryUnavailableError := errors.New("database unavailable")

	testCases := []struct {
		name                    string
		requestBody             io.Reader
		setupMock               func(mockRepository *domainusermocks.MockRepository)
		wantHTTPStatus          int
		wantResponseBodyPrefix  string
		wantResponseBodyContain string
	}{
		{
			name:                   "malformed_json_returns_400",
			requestBody:            strings.NewReader(`{not-json`),
			setupMock:              func(mockRepository *domainusermocks.MockRepository) {},
			wantHTTPStatus:         http.StatusBadRequest,
			wantResponseBodyPrefix: "invalid JSON body",
		},
		{
			name:                   "empty_body_returns_400",
			requestBody:            bytes.NewReader(nil),
			setupMock:              func(mockRepository *domainusermocks.MockRepository) {},
			wantHTTPStatus:         http.StatusBadRequest,
			wantResponseBodyPrefix: "invalid JSON body",
		},
		{
			name:        "invalid_credentials_returns_401",
			requestBody: strings.NewReader(validAuthLoginJSON()),
			setupMock: func(mockRepository *domainusermocks.MockRepository) {
				mockRepository.EXPECT().FindByEmail(gomock.Any(), "ada@example.com").Return(&domainuser.User{
					ID: authHandlerTestUserID, FirstName: "A", LastName: "B", Email: "ada@example.com",
					Password: string(wrongPasswordHash),
				}, nil)
			},
			wantHTTPStatus:         http.StatusUnauthorized,
			wantResponseBodyPrefix: usecase.ErrInvalidCredentials.Error(),
		},
		{
			name:        "use_case_non_credential_error_returns_500",
			requestBody: strings.NewReader(validAuthLoginJSON()),
			setupMock: func(mockRepository *domainusermocks.MockRepository) {
				mockRepository.EXPECT().FindByEmail(gomock.Any(), "ada@example.com").Return(nil, repositoryUnavailableError)
			},
			wantHTTPStatus:         http.StatusInternalServerError,
			wantResponseBodyPrefix: http.StatusText(http.StatusInternalServerError),
		},
	}

	for _, testCase := range testCases {
		tb.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			controller := gomock.NewController(t)
			mockRepository := domainusermocks.NewMockRepository(controller)
			testCase.setupMock(mockRepository)

			handler := newAuthHandlerWithRepository(t, mockRepository)
			request := httptest.NewRequest(http.MethodPost, "/auth", testCase.requestBody)
			if testCase.requestBody != nil {
				request.Header.Set("Content-Type", "application/json")
			}
			responseRecorder := httptest.NewRecorder()
			handler.PostAuth(responseRecorder, request)

			if responseRecorder.Code != testCase.wantHTTPStatus {
				t.Fatalf("status: got %d want %d, body: %s", responseRecorder.Code, testCase.wantHTTPStatus, responseRecorder.Body.String())
			}
			responseBody := strings.TrimSpace(responseRecorder.Body.String())
			if testCase.wantResponseBodyPrefix != "" && !strings.HasPrefix(responseBody, testCase.wantResponseBodyPrefix) {
				t.Fatalf("body prefix: got %q want prefix %q", responseBody, testCase.wantResponseBodyPrefix)
			}
			if testCase.wantResponseBodyContain != "" && !strings.Contains(responseBody, testCase.wantResponseBodyContain) {
				t.Fatalf("body: got %q want substring %q", responseBody, testCase.wantResponseBodyContain)
			}
		})
	}
}
