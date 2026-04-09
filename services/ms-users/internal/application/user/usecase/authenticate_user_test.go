package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	domainuser "ilia-golang-challenge/services/ms-users/internal/domain/user"
	domainusermocks "ilia-golang-challenge/services/ms-users/internal/domain/user/mocks"
)

type stubTokenIssuer struct {
	token string
	err   error
}

func (issuer *stubTokenIssuer) IssueAccessToken(subjectUserID string) (string, error) {
	if issuer.err != nil {
		return "", issuer.err
	}
	if issuer.token != "" {
		return issuer.token, nil
	}
	return "signed-for-" + subjectUserID, nil
}

func TestAuthenticateUserUseCase_Execute_success(t *testing.T) {
	t.Parallel()

	t.Run("valid_email_password_returns_token_and_user_view", func(t *testing.T) {
		t.Parallel()
		userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
		plainPassword := "password123"
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.MinCost)
		if err != nil {
			t.Fatalf("bcrypt: %v", err)
		}
		storedUser := &domainuser.User{
			ID:        userID,
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "ada@example.com",
			Password:  string(passwordHash),
		}

		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockRepository.EXPECT().FindByEmail(gomock.Any(), "ada@example.com").Return(storedUser, nil)

		useCase := NewAuthenticateUserUseCase(mockRepository, &stubTokenIssuer{token: "jwt-token"})
		result, err := useCase.Execute(context.Background(), AuthenticateUserInput{
			Email:    "Ada@Example.com",
			Password: plainPassword,
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if result.AccessToken != "jwt-token" {
			t.Fatalf("token: got %q", result.AccessToken)
		}
		if result.User.ID != userID || result.User.Email != "ada@example.com" || result.User.FirstName != "Ada" {
			t.Fatalf("user view: %+v", result.User)
		}
	})
}

func TestAuthenticateUserUseCase_Execute_errors(t *testing.T) {
	t.Parallel()

	issuerError := errors.New("signer failed")
	wrongPassHash, err := bcrypt.GenerateFromPassword([]byte("right-pass"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	tokenCaseHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	testCases := []struct {
		name                    string
		input                   AuthenticateUserInput
		setupMock               func(tb *testing.T, repository *domainusermocks.MockRepository)
		tokenIssuer             AccessTokenIssuer
		wantErrorIs             error
		wantErrorContainsSigner bool
	}{
		{
			name:  "invalid_email_shape_maps_to_invalid_credentials",
			input: AuthenticateUserInput{Email: "not-an-email", Password: "12345678"},
			setupMock: func(tb *testing.T, repository *domainusermocks.MockRepository) {
				_ = tb
			},
			tokenIssuer: &stubTokenIssuer{},
			wantErrorIs: ErrInvalidCredentials,
		},
		{
			name:  "unknown_email_maps_to_invalid_credentials",
			input: AuthenticateUserInput{Email: "missing@example.com", Password: "12345678"},
			setupMock: func(tb *testing.T, repository *domainusermocks.MockRepository) {
				_ = tb
				repository.EXPECT().FindByEmail(gomock.Any(), "missing@example.com").Return(nil, ErrNotFound)
			},
			tokenIssuer: &stubTokenIssuer{},
			wantErrorIs: ErrInvalidCredentials,
		},
		{
			name:  "wrong_password_maps_to_invalid_credentials",
			input: AuthenticateUserInput{Email: "ada@example.com", Password: "wrong-pass"},
			setupMock: func(tb *testing.T, repository *domainusermocks.MockRepository) {
				_ = tb
				repository.EXPECT().FindByEmail(gomock.Any(), "ada@example.com").Return(&domainuser.User{
					ID: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", Email: "ada@example.com",
					FirstName: "A", LastName: "B", Password: string(wrongPassHash),
				}, nil)
			},
			tokenIssuer: &stubTokenIssuer{},
			wantErrorIs: ErrInvalidCredentials,
		},
		{
			name:  "repository_error_is_propagated_wrapped",
			input: AuthenticateUserInput{Email: "ada@example.com", Password: "x"},
			setupMock: func(tb *testing.T, repository *domainusermocks.MockRepository) {
				_ = tb
				repository.EXPECT().FindByEmail(gomock.Any(), "ada@example.com").Return(nil, fmt.Errorf("db: %w", errors.New("down")))
			},
			tokenIssuer: &stubTokenIssuer{},
			wantErrorIs: nil,
		},
		{
			name:  "token_issuer_error_is_propagated",
			input: AuthenticateUserInput{Email: "ada@example.com", Password: "password123"},
			setupMock: func(tb *testing.T, repository *domainusermocks.MockRepository) {
				_ = tb
				repository.EXPECT().FindByEmail(gomock.Any(), "ada@example.com").Return(&domainuser.User{
					ID: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", Email: "ada@example.com",
					FirstName: "A", LastName: "B", Password: string(tokenCaseHash),
				}, nil)
			},
			tokenIssuer:             &stubTokenIssuer{err: issuerError},
			wantErrorContainsSigner: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			controller := gomock.NewController(t)
			mockRepository := domainusermocks.NewMockRepository(controller)
			testCase.setupMock(t, mockRepository)
			useCase := NewAuthenticateUserUseCase(mockRepository, testCase.tokenIssuer)
			_, err := useCase.Execute(context.Background(), testCase.input)
			if testCase.wantErrorContainsSigner {
				if err == nil || !errors.Is(err, issuerError) {
					t.Fatalf("expected wrapped signer error, got %v", err)
				}
				return
			}
			if testCase.wantErrorIs != nil {
				if err == nil {
					t.Fatal("expected error")
				}
				if !errors.Is(err, testCase.wantErrorIs) {
					t.Fatalf("errors.Is: got %v want %v", err, testCase.wantErrorIs)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("unexpected invalid credentials: %v", err)
			}
		})
	}
}
