package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"ilia-golang-challenge/ms-users/internal/application/user/usecase"
	domainuser "ilia-golang-challenge/ms-users/internal/domain/user"
	domainusermocks "ilia-golang-challenge/ms-users/internal/domain/user/mocks"
)

func newUserHandlerWithRepository(repository domainuser.Repository) *UserHandler {
	return NewUserHandler(
		usecase.NewCreateUserUseCase(repository, bcrypt.MinCost),
		usecase.NewListUsersUseCase(repository),
		usecase.NewGetUserUseCase(repository),
		usecase.NewUpdateUserUseCase(repository, bcrypt.MinCost),
		usecase.NewDeleteUserUseCase(repository),
	)
}

func postUserValidRegistrationJSON() string {
	return `{"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com","password":"secret123"}`
}

func TestUserHandler_PostUser_success(t *testing.T) {
	t.Parallel()

	t.Run("returns_201_with_location_header_and_json_body_without_password", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

		handler := newUserHandlerWithRepository(mockRepository)
		request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(postUserValidRegistrationJSON()))
		request.Header.Set("Content-Type", "application/json")
		responseRecorder := httptest.NewRecorder()
		handler.PostUser(responseRecorder, request)

		if responseRecorder.Code != http.StatusCreated {
			t.Fatalf("status: got %d want %d, body: %s", responseRecorder.Code, http.StatusCreated, responseRecorder.Body.String())
		}
		if contentType := responseRecorder.Header().Get("Content-Type"); contentType != "application/json" {
			t.Fatalf("Content-Type: got %q want application/json", contentType)
		}
		location := responseRecorder.Header().Get("Location")
		if !strings.HasPrefix(location, "/users/") || len(location) <= len("/users/") {
			t.Fatalf("Location: got %q want /users/{id}", location)
		}
		userIDFromLocation := strings.TrimPrefix(location, "/users/")
		var responseBody usersResponse
		if err := json.NewDecoder(responseRecorder.Body).Decode(&responseBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if responseBody.ID != userIDFromLocation {
			t.Fatalf("body id %q does not match Location id %q", responseBody.ID, userIDFromLocation)
		}
		if responseBody.FirstName != "Ada" || responseBody.LastName != "Lovelace" || responseBody.Email != "ada@example.com" {
			t.Fatalf("unexpected body: %+v", responseBody)
		}
	})

	t.Run("decodes_json_normalizes_fields_and_persists_bcrypt_hashed_password", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		var savedUser *domainuser.User
		mockRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Do(func(_ context.Context, user *domainuser.User) {
			savedUser = user
		}).Return(nil)

		handler := newUserHandlerWithRepository(mockRepository)
		requestBody := `{"first_name":"  Pat  ","last_name":" Kim ","email":"Pat@EXAMPLE.org","password":"abcdefgh"}`
		request := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(requestBody))
		request.Header.Set("Content-Type", "application/json")
		responseRecorder := httptest.NewRecorder()
		handler.PostUser(responseRecorder, request)

		if responseRecorder.Code != http.StatusCreated {
			t.Fatalf("status: got %d body: %s", responseRecorder.Code, responseRecorder.Body.String())
		}
		if savedUser == nil {
			t.Fatal("expected Save to be called")
		}
		if savedUser.FirstName != "Pat" || savedUser.LastName != "Kim" || savedUser.Email != "pat@example.org" {
			t.Fatalf("use case received wrong entity: %+v", savedUser)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(savedUser.Password), []byte("abcdefgh")); err != nil {
			t.Fatalf("persisted password should be bcrypt of request: %v", err)
		}
	})
}

func TestUserHandler_PostUser_errors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                    string
		requestBody             io.Reader
		repositorySaveError     error
		expectRepositorySave    bool
		wantHTTPStatus          int
		wantResponseBodyPrefix  string
		wantResponseBodyContain string
	}{
		{
			name:                   "invalid_json_returns_400",
			requestBody:            strings.NewReader(`{not-json`),
			expectRepositorySave:   false,
			wantHTTPStatus:         http.StatusBadRequest,
			wantResponseBodyPrefix: "invalid JSON body",
		},
		{
			name:                    "empty_first_name_returns_400_with_validation_message",
			requestBody:             strings.NewReader(`{"first_name":"","last_name":"L","email":"a@b.co","password":"12345678"}`),
			expectRepositorySave:    false,
			wantHTTPStatus:          http.StatusBadRequest,
			wantResponseBodyContain: domainuser.ErrInvalidFirstName.Error(),
		},
		{
			name:                    "empty_last_name_returns_400_with_validation_message",
			requestBody:             strings.NewReader(`{"first_name":"A","last_name":"","email":"a@b.co","password":"12345678"}`),
			expectRepositorySave:    false,
			wantHTTPStatus:          http.StatusBadRequest,
			wantResponseBodyContain: domainuser.ErrInvalidLastName.Error(),
		},
		{
			name:                    "invalid_email_returns_400_with_validation_message",
			requestBody:             strings.NewReader(`{"first_name":"A","last_name":"B","email":"not-an-email","password":"12345678"}`),
			expectRepositorySave:    false,
			wantHTTPStatus:          http.StatusBadRequest,
			wantResponseBodyContain: domainuser.ErrInvalidEmail.Error(),
		},
		{
			name:                    "empty_password_returns_400_with_validation_message",
			requestBody:             strings.NewReader(`{"first_name":"A","last_name":"B","email":"a@b.co","password":""}`),
			expectRepositorySave:    false,
			wantHTTPStatus:          http.StatusBadRequest,
			wantResponseBodyContain: domainuser.ErrInvalidPassword.Error(),
		},
		{
			name:                    "password_too_short_returns_400_with_validation_message",
			requestBody:             strings.NewReader(`{"first_name":"A","last_name":"B","email":"a@b.co","password":"short"}`),
			expectRepositorySave:    false,
			wantHTTPStatus:          http.StatusBadRequest,
			wantResponseBodyContain: domainuser.ErrPasswordTooShort.Error(),
		},
		{
			name:                   "duplicate_email_from_repository_returns_409",
			requestBody:            strings.NewReader(postUserValidRegistrationJSON()),
			repositorySaveError:    fmt.Errorf("save user: %w", usecase.ErrEmailAlreadyExists),
			expectRepositorySave:   true,
			wantHTTPStatus:         http.StatusConflict,
			wantResponseBodyPrefix: usecase.ErrEmailAlreadyExists.Error(),
		},
		{
			name:                   "repository_failure_returns_500",
			requestBody:            strings.NewReader(postUserValidRegistrationJSON()),
			repositorySaveError:    errors.New("persist failed"),
			expectRepositorySave:   true,
			wantHTTPStatus:         http.StatusInternalServerError,
			wantResponseBodyPrefix: http.StatusText(http.StatusInternalServerError),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			controller := gomock.NewController(t)
			mockRepository := domainusermocks.NewMockRepository(controller)
			if testCase.expectRepositorySave {
				mockRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(testCase.repositorySaveError)
			}

			handler := newUserHandlerWithRepository(mockRepository)
			request := httptest.NewRequest(http.MethodPost, "/users", testCase.requestBody)
			if testCase.requestBody != nil {
				request.Header.Set("Content-Type", "application/json")
			}
			responseRecorder := httptest.NewRecorder()
			handler.PostUser(responseRecorder, request)

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

func TestIsUserRegistrationValidationError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		inputError        error
		expectedRecognize bool
	}{
		{
			name:              "direct_ErrInvalidFirstName",
			inputError:        domainuser.ErrInvalidFirstName,
			expectedRecognize: true,
		},
		{
			name:              "wrapped_with_errors_join",
			inputError:        errors.Join(errors.New("outer"), domainuser.ErrInvalidEmail),
			expectedRecognize: true,
		},
		{
			name:              "wrapped_with_fmt_like_create_use_case",
			inputError:        fmt.Errorf("create user: %w", domainuser.ErrPasswordTooShort),
			expectedRecognize: true,
		},
		{
			name:              "ErrEmailAlreadyExists_is_not_validation",
			inputError:        usecase.ErrEmailAlreadyExists,
			expectedRecognize: false,
		},
		{
			name:              "generic_error",
			inputError:        errors.New("database down"),
			expectedRecognize: false,
		},
		{
			name:              "nil_error",
			inputError:        nil,
			expectedRecognize: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := isUserRegistrationValidationError(testCase.inputError); got != testCase.expectedRecognize {
				t.Fatalf("isUserRegistrationValidationError(%v) = %v, want %v", testCase.inputError, got, testCase.expectedRecognize)
			}
		})
	}
}
