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

	"golang.org/x/crypto/bcrypt"

	"ilia-golang-challenge/ms-users/internal/application/user/usecase"
	domainuser "ilia-golang-challenge/ms-users/internal/domain/user"
)

type stubUserRepository struct {
	saveErr error
}

func (s *stubUserRepository) Save(ctx context.Context, entity *domainuser.User) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	return nil
}

func validUserJSON() string {
	return `{"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com","password":"secret123"}`
}

func TestUserHandler_PostUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		body           io.Reader
		saveErr        error
		wantCode       int
		wantBodyPrefix string
		wantBodySubstr string
		checkCreated   func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:           "method not allowed returns 405",
			method:         http.MethodGet,
			body:           strings.NewReader(validUserJSON()),
			wantCode:       http.StatusMethodNotAllowed,
			wantBodyPrefix: http.StatusText(http.StatusMethodNotAllowed),
		},
		{
			name:           "invalid JSON returns 400",
			method:         http.MethodPost,
			body:           strings.NewReader(`{not-json`),
			wantCode:       http.StatusBadRequest,
			wantBodyPrefix: "invalid JSON body",
		},
		{
			name:           "empty first name returns 400 with validation context",
			method:         http.MethodPost,
			body:           strings.NewReader(`{"first_name":"","last_name":"L","email":"a@b.co","password":"12345678"}`),
			wantCode:       http.StatusBadRequest,
			wantBodySubstr: domainuser.ErrInvalidFirstName.Error(),
		},
		{
			name:           "empty last name returns 400",
			method:         http.MethodPost,
			body:           strings.NewReader(`{"first_name":"A","last_name":"","email":"a@b.co","password":"12345678"}`),
			wantCode:       http.StatusBadRequest,
			wantBodySubstr: domainuser.ErrInvalidLastName.Error(),
		},
		{
			name:           "invalid email returns 400",
			method:         http.MethodPost,
			body:           strings.NewReader(`{"first_name":"A","last_name":"B","email":"not-an-email","password":"12345678"}`),
			wantCode:       http.StatusBadRequest,
			wantBodySubstr: domainuser.ErrInvalidEmail.Error(),
		},
		{
			name:           "empty password returns 400",
			method:         http.MethodPost,
			body:           strings.NewReader(`{"first_name":"A","last_name":"B","email":"a@b.co","password":""}`),
			wantCode:       http.StatusBadRequest,
			wantBodySubstr: domainuser.ErrInvalidPassword.Error(),
		},
		{
			name:           "password too short returns 400",
			method:         http.MethodPost,
			body:           strings.NewReader(`{"first_name":"A","last_name":"B","email":"a@b.co","password":"short"}`),
			wantCode:       http.StatusBadRequest,
			wantBodySubstr: domainuser.ErrPasswordTooShort.Error(),
		},
		{
			name:           "duplicate email returns 409",
			method:         http.MethodPost,
			body:           strings.NewReader(validUserJSON()),
			saveErr:        fmt.Errorf("save user: %w", domainuser.ErrEmailAlreadyExists),
			wantCode:       http.StatusConflict,
			wantBodyPrefix: domainuser.ErrEmailAlreadyExists.Error(),
		},
		{
			name:           "repository error returns 500",
			method:         http.MethodPost,
			body:           strings.NewReader(validUserJSON()),
			saveErr:        errors.New("persist failed"),
			wantCode:       http.StatusInternalServerError,
			wantBodyPrefix: http.StatusText(http.StatusInternalServerError),
		},
		{
			name:     "success returns 201 location and JSON body without password",
			method:   http.MethodPost,
			body:     strings.NewReader(validUserJSON()),
			wantCode: http.StatusCreated,
			checkCreated: func(t *testing.T, rr *httptest.ResponseRecorder) {
				t.Helper()
				if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
					t.Fatalf("Content-Type: got %q want application/json", ct)
				}
				loc := rr.Header().Get("Location")
				if !strings.HasPrefix(loc, "/users/") || len(loc) <= len("/users/") {
					t.Fatalf("Location: got %q want /users/{id}", loc)
				}
				id := strings.TrimPrefix(loc, "/users/")
				var got usersResponse
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got.ID != id {
					t.Fatalf("body id %q does not match Location id %q", got.ID, id)
				}
				if got.FirstName != "Ada" || got.LastName != "Lovelace" || got.Email != "ada@example.com" {
					t.Fatalf("unexpected body: %+v", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &stubUserRepository{saveErr: tt.saveErr}
			h := NewUserHandler(usecase.NewCreateUserUseCase(repo, bcrypt.MinCost))

			req := httptest.NewRequest(tt.method, "/users", tt.body)
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			rr := httptest.NewRecorder()
			h.PostUser(rr, req)

			if rr.Code != tt.wantCode {
				t.Fatalf("status: got %d want %d, body: %s", rr.Code, tt.wantCode, rr.Body.String())
			}
			body := strings.TrimSpace(rr.Body.String())
			if tt.wantBodyPrefix != "" && !strings.HasPrefix(body, tt.wantBodyPrefix) {
				t.Fatalf("body prefix: got %q want prefix %q", body, tt.wantBodyPrefix)
			}
			if tt.wantBodySubstr != "" && !strings.Contains(body, tt.wantBodySubstr) {
				t.Fatalf("body: got %q want substring %q", body, tt.wantBodySubstr)
			}
			if tt.checkCreated != nil {
				tt.checkCreated(t, rr)
			}
		})
	}
}

func TestIsUserRegistrationValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "direct invalid first name",
			err:  domainuser.ErrInvalidFirstName,
			want: true,
		},
		{
			name: "wrapped validation error",
			err:  errors.Join(errors.New("outer"), domainuser.ErrInvalidEmail),
			want: true,
		},
		{
			name: "wrapped with fmt like use case",
			err:  fmt.Errorf("create user: %w", domainuser.ErrPasswordTooShort),
			want: true,
		},
		{
			name: "email already exists is not validation",
			err:  domainuser.ErrEmailAlreadyExists,
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("database down"),
			want: false,
		},
		{
			name: "nil",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isUserRegistrationValidationError(tt.err); got != tt.want {
				t.Fatalf("isUserRegistrationValidationError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestUserHandler_PostUser_decodeMapsAllFieldsToUseCase(t *testing.T) {
	t.Parallel()

	repo := &stubCaptureRepository{}
	h := NewUserHandler(usecase.NewCreateUserUseCase(repo, bcrypt.MinCost))

	body := `{"first_name":"  Pat  ","last_name":" Kim ","email":"Pat@EXAMPLE.org","password":"abcdefgh"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.PostUser(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: got %d body: %s", rr.Code, rr.Body.String())
	}
	if repo.saved == nil {
		t.Fatal("expected Save to be called")
	}
	u := repo.saved
	if u.FirstName != "Pat" || u.LastName != "Kim" || u.Email != "pat@example.org" {
		t.Fatalf("use case received wrong entity: %+v", u)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("abcdefgh")); err != nil {
		t.Fatalf("persisted password should be bcrypt of request: %v", err)
	}
}

type stubCaptureRepository struct {
	saved *domainuser.User
}

func (s *stubCaptureRepository) Save(ctx context.Context, entity *domainuser.User) error {
	s.saved = entity
	return nil
}
