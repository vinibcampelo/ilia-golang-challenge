package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"

	"ilia-golang-challenge/services/ms-users/internal/application/user/usecase"
	domainuser "ilia-golang-challenge/services/ms-users/internal/domain/user"
	domainusermocks "ilia-golang-challenge/services/ms-users/internal/domain/user/mocks"
	"ilia-golang-challenge/services/ms-users/internal/infrastructure/jwtissuer"
)

const routerTestJWTSecret = "router-test-jwt-secret-key-min-length-32!!"

func newTestRouter(tb *testing.T, repository domainuser.Repository) http.Handler {
	tb.Helper()
	tokenIssuer := &jwtissuer.HS256AccessTokenIssuer{
		Secret: []byte(routerTestJWTSecret),
		TTL:    time.Hour,
	}
	userHandler := NewUserHandler(
		usecase.NewCreateUserUseCase(repository, bcrypt.MinCost),
		usecase.NewListUsersUseCase(repository),
		usecase.NewGetUserUseCase(repository),
		usecase.NewUpdateUserUseCase(repository, bcrypt.MinCost),
		usecase.NewDeleteUserUseCase(repository),
	)
	authHandler := NewAuthHandler(usecase.NewAuthenticateUserUseCase(repository, tokenIssuer))
	return NewRouter(userHandler, authHandler, []byte(routerTestJWTSecret), nil)
}

func mintValidToken(tb *testing.T, subject string) string {
	tb.Helper()
	issuer := &jwtissuer.HS256AccessTokenIssuer{Secret: []byte(routerTestJWTSecret), TTL: time.Hour}
	token, err := issuer.IssueAccessToken(subject)
	if err != nil {
		tb.Fatalf("mint token: %v", err)
	}
	return token
}

func mintExpiredToken(tb *testing.T, subject string) string {
	tb.Helper()
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(routerTestJWTSecret))
	if err != nil {
		tb.Fatalf("sign expired token: %v", err)
	}
	return signed
}

func TestRouter_post_auth_success(tb *testing.T) {
	tb.Parallel()

	tb.Run("returns_access_token_and_user_on_valid_credentials", func(t *testing.T) {
		t.Parallel()
		userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
		plainPassword := "password123"
		hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.MinCost)
		if err != nil {
			t.Fatalf("bcrypt: %v", err)
		}
		storedUser := &domainuser.User{
			ID: userID, FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com",
			Password: string(hash),
		}
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockRepository.EXPECT().FindByEmail(gomock.Any(), "ada@example.com").Return(storedUser, nil)

		router := newTestRouter(t, mockRepository)
		body := `{"user":{"email":"ada@example.com","password":"password123"}}`
		request := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)

		if responseRecorder.Code != http.StatusOK {
			t.Fatalf("status: got %d body: %s", responseRecorder.Code, responseRecorder.Body.String())
		}
		var parsed struct {
			AccessToken string        `json:"access_token"`
			User        usersResponse `json:"user"`
		}
		if err := json.NewDecoder(responseRecorder.Body).Decode(&parsed); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if parsed.AccessToken == "" {
			t.Fatal("expected non-empty access_token")
		}
		if parsed.User.ID != userID || parsed.User.Email != "ada@example.com" {
			t.Fatalf("user: %+v", parsed.User)
		}
	})
}

func TestRouter_post_auth_errors(tb *testing.T) {
	tb.Parallel()

	testCases := []struct {
		name           string
		requestBody    string
		setupMock      func(repository *domainusermocks.MockRepository)
		wantHTTPStatus int
		wantBodyPrefix string
	}{
		{
			name:           "malformed_json_returns_400",
			requestBody:    `{`,
			setupMock:      func(repository *domainusermocks.MockRepository) {},
			wantHTTPStatus: http.StatusBadRequest,
			wantBodyPrefix: "invalid JSON body",
		},
		{
			name:        "wrong_password_returns_401",
			requestBody: `{"user":{"email":"ada@example.com","password":"nope"}}`,
			setupMock: func(repository *domainusermocks.MockRepository) {
				hash, err := bcrypt.GenerateFromPassword([]byte("real"), bcrypt.MinCost)
				if err != nil {
					tb.Fatalf("bcrypt: %v", err)
				}
				repository.EXPECT().FindByEmail(gomock.Any(), "ada@example.com").Return(&domainuser.User{
					ID: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", Email: "ada@example.com",
					FirstName: "A", LastName: "B", Password: string(hash),
				}, nil)
			},
			wantHTTPStatus: http.StatusUnauthorized,
			wantBodyPrefix: usecase.ErrInvalidCredentials.Error(),
		},
	}

	for _, testCase := range testCases {
		tb.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			controller := gomock.NewController(t)
			mockRepository := domainusermocks.NewMockRepository(controller)
			testCase.setupMock(mockRepository)
			router := newTestRouter(t, mockRepository)
			request := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(testCase.requestBody))
			request.Header.Set("Content-Type", "application/json")
			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, request)
			if responseRecorder.Code != testCase.wantHTTPStatus {
				t.Fatalf("status: got %d want %d body: %s", responseRecorder.Code, testCase.wantHTTPStatus, responseRecorder.Body.String())
			}
			if testCase.wantBodyPrefix != "" {
				body := strings.TrimSpace(responseRecorder.Body.String())
				if !strings.HasPrefix(body, testCase.wantBodyPrefix) {
					t.Fatalf("body: got %q want prefix %q", body, testCase.wantBodyPrefix)
				}
			}
		})
	}
}

func TestRouter_protected_routes_require_valid_bearer(tb *testing.T) {
	tb.Parallel()

	userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	testCases := []struct {
		name            string
		method          string
		path            string
		authorization   string
		wantHTTPStatus  int
		setupRepository func(repository *domainusermocks.MockRepository)
	}{
		{
			name:           "GET_users_without_authorization_header_returns_401",
			method:         http.MethodGet,
			path:           "/users",
			authorization:  "",
			wantHTTPStatus: http.StatusUnauthorized,
			setupRepository: func(repository *domainusermocks.MockRepository) {
			},
		},
		{
			name:           "GET_users_with_malformed_bearer_prefix_returns_401",
			method:         http.MethodGet,
			path:           "/users",
			authorization:  "Token abc",
			wantHTTPStatus: http.StatusUnauthorized,
			setupRepository: func(repository *domainusermocks.MockRepository) {
			},
		},
		{
			name:           "GET_users_with_invalid_jwt_returns_401",
			method:         http.MethodGet,
			path:           "/users",
			authorization:  "Bearer not-a-valid-jwt",
			wantHTTPStatus: http.StatusUnauthorized,
			setupRepository: func(repository *domainusermocks.MockRepository) {
			},
		},
		{
			name:           "GET_users_with_expired_jwt_returns_401",
			method:         http.MethodGet,
			path:           "/users",
			authorization:  "Bearer " + mintExpiredToken(tb, userID),
			wantHTTPStatus: http.StatusUnauthorized,
			setupRepository: func(repository *domainusermocks.MockRepository) {
			},
		},
		{
			name:           "GET_user_by_id_without_bearer_returns_401",
			method:         http.MethodGet,
			path:           "/users/" + userID,
			authorization:  "",
			wantHTTPStatus: http.StatusUnauthorized,
			setupRepository: func(repository *domainusermocks.MockRepository) {
			},
		},
	}

	for _, testCase := range testCases {
		tb.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			controller := gomock.NewController(t)
			mockRepository := domainusermocks.NewMockRepository(controller)
			testCase.setupRepository(mockRepository)
			router := newTestRouter(t, mockRepository)
			request := httptest.NewRequest(testCase.method, testCase.path, nil)
			if testCase.authorization != "" {
				request.Header.Set("Authorization", testCase.authorization)
			}
			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, request)
			if responseRecorder.Code != testCase.wantHTTPStatus {
				t.Fatalf("status: got %d want %d body: %s", responseRecorder.Code, testCase.wantHTTPStatus, responseRecorder.Body.String())
			}
			body := strings.TrimSpace(responseRecorder.Body.String())
			if !strings.Contains(body, unauthorizedDescription) {
				t.Fatalf("body should mention unauthorized description; got %q", body)
			}
		})
	}
}

func TestRouter_get_users_success_with_bearer(tb *testing.T) {
	tb.Parallel()
	tb.Run("returns_json_array_when_repository_lists_users", func(t *testing.T) {
		t.Parallel()
		userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockRepository.EXPECT().List(gomock.Any()).Return([]domainuser.User{
			{ID: userID, FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com"},
		}, nil)

		router := newTestRouter(t, mockRepository)
		request := httptest.NewRequest(http.MethodGet, "/users", nil)
		request.Header.Set("Authorization", "Bearer "+mintValidToken(t, userID))
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)

		if responseRecorder.Code != http.StatusOK {
			t.Fatalf("status: got %d body: %s", responseRecorder.Code, responseRecorder.Body.String())
		}
		var users []usersResponse
		if err := json.NewDecoder(responseRecorder.Body).Decode(&users); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(users) != 1 || users[0].ID != userID {
			t.Fatalf("users: %+v", users)
		}
	})
}

func TestRouter_user_scoped_routes_require_matching_subject(tb *testing.T) {
	tb.Parallel()

	selfID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	otherID := "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12"

	tb.Run("GET_user_returns_403_when_token_subject_differs_from_path_id", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		router := newTestRouter(t, mockRepository)
		request := httptest.NewRequest(http.MethodGet, "/users/"+selfID, nil)
		request.Header.Set("Authorization", "Bearer "+mintValidToken(t, otherID))
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusForbidden {
			t.Fatalf("status: got %d body: %s", responseRecorder.Code, responseRecorder.Body.String())
		}
	})

	tb.Run("GET_user_returns_200_when_token_subject_matches_path_id", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockRepository.EXPECT().FindByID(gomock.Any(), selfID).Return(&domainuser.User{
			ID: selfID, FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com", Password: "x",
		}, nil)
		router := newTestRouter(t, mockRepository)
		request := httptest.NewRequest(http.MethodGet, "/users/"+selfID, nil)
		request.Header.Set("Authorization", "Bearer "+mintValidToken(t, selfID))
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusOK {
			t.Fatalf("status: got %d body: %s", responseRecorder.Code, responseRecorder.Body.String())
		}
	})

	tb.Run("PATCH_user_returns_403_when_token_subject_differs_from_path_id", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		router := newTestRouter(t, mockRepository)
		body := `{"first_name":"A"}`
		request := httptest.NewRequest(http.MethodPatch, "/users/"+selfID, bytes.NewBufferString(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+mintValidToken(t, otherID))
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusForbidden {
			t.Fatalf("status: got %d body: %s", responseRecorder.Code, responseRecorder.Body.String())
		}
	})

	tb.Run("DELETE_user_returns_204_without_body_when_subject_matches", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockRepository.EXPECT().DeleteByID(gomock.Any(), selfID).Return(nil)
		router := newTestRouter(t, mockRepository)
		request := httptest.NewRequest(http.MethodDelete, "/users/"+selfID, nil)
		request.Header.Set("Authorization", "Bearer "+mintValidToken(t, selfID))
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusNoContent {
			t.Fatalf("status: got %d body: %s", responseRecorder.Code, responseRecorder.Body.String())
		}
		if responseRecorder.Body.Len() != 0 {
			t.Fatalf("expected empty body for 204, got %q", responseRecorder.Body.String())
		}
	})
}
