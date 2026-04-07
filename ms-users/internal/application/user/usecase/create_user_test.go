package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	domainuser "ilia-golang-challenge/ms-users/internal/domain/user"
	domainusermocks "ilia-golang-challenge/ms-users/internal/domain/user/mocks"
)

func newMockRepositoryWithSaveCapture(t *testing.T) (*domainusermocks.MockRepository, *[]*domainuser.User) {
	t.Helper()

	controller := gomock.NewController(t)
	mockRepository := domainusermocks.NewMockRepository(controller)
	var savedUsers []*domainuser.User
	mockRepository.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, user *domainuser.User) error {
			savedUsers = append(savedUsers, user)
			return nil
		})

	return mockRepository, &savedUsers
}

func TestCreateUserUseCase_Execute_success(t *testing.T) {
	t.Run("normalizes_output_persists_once_and_hashes_password", func(t *testing.T) {
		t.Parallel()
		mockRepository, savedUsers := newMockRepositoryWithSaveCapture(t)
		createUserUseCase := NewCreateUserUseCase(mockRepository, bcrypt.MinCost)

		output, err := createUserUseCase.Execute(context.Background(), CreateUserInput{
			FirstName: "  user1  ",
			LastName:  " lastname1 ",
			Email:     "User1@TEST.COM",
			Password:  "password123",
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if _, err := uuid.Parse(output.ID); err != nil {
			t.Fatalf("output.ID is not a valid UUID: %q", output.ID)
		}
		if output.FirstName != "user1" || output.LastName != "lastname1" || output.Email != "user1@test.com" {
			t.Fatalf("unexpected output: %+v", output)
		}
		if len(*savedUsers) != 1 {
			t.Fatalf("Save calls: got %d want 1", len(*savedUsers))
		}
		savedUser := (*savedUsers)[0]
		if savedUser.ID != output.ID {
			t.Fatalf("saved ID: got %q want %q", savedUser.ID, output.ID)
		}
		if savedUser.FirstName != "user1" || savedUser.LastName != "lastname1" || savedUser.Email != "user1@test.com" {
			t.Fatalf("saved entity: %+v", savedUser)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(savedUser.Password), []byte("password123")); err != nil {
			t.Fatalf("saved password should be bcrypt of plaintext: %v", err)
		}
	})

	t.Run("accepts_password_at_minimum_length", func(t *testing.T) {
		t.Parallel()
		mockRepository, savedUsers := newMockRepositoryWithSaveCapture(t)
		createUserUseCase := NewCreateUserUseCase(mockRepository, bcrypt.MinCost)

		_, err := createUserUseCase.Execute(context.Background(), CreateUserInput{
			FirstName: "user1",
			LastName:  "lastname1",
			Email:     "user1@test.com",
			Password:  "12345678",
		})
		if err != nil {
			t.Fatalf("8-rune password should pass: %v", err)
		}
		if len(*savedUsers) != 1 {
			t.Fatalf("expected one Save, got %d", len(*savedUsers))
		}
	})
}

func TestCreateUserUseCase_Execute_validation_errors(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name          string
		input         CreateUserInput
		expectedError error
	}{{
		name: "empty first name",
		input: CreateUserInput{
			FirstName: "",
			LastName:  "lastname1",
			Email:     "user1@test.com",
			Password:  "12345678",
		},
		expectedError: domainuser.ErrInvalidFirstName,
	}, {
		name: "whitespace only first name",
		input: CreateUserInput{
			FirstName: "   \t",
			LastName:  "lastname1",
			Email:     "user1@test.com",
			Password:  "12345678",
		},
		expectedError: domainuser.ErrInvalidFirstName,
	}, {
		name: "empty last name",
		input: CreateUserInput{
			FirstName: "user1",
			LastName:  "",
			Email:     "user1@test.com",
			Password:  "12345678",
		},
		expectedError: domainuser.ErrInvalidLastName,
	}, {
		name: "empty email",
		input: CreateUserInput{
			FirstName: "user1",
			LastName:  "lastname1",
			Email:     "",
			Password:  "12345678",
		},
		expectedError: domainuser.ErrInvalidEmail,
	}, {
		name: "invalid email format",
		input: CreateUserInput{
			FirstName: "user1",
			LastName:  "lastname1",
			Email:     "not-an-email",
			Password:  "12345678",
		},
		expectedError: domainuser.ErrInvalidEmail,
	}, {
		name: "email with display name",
		input: CreateUserInput{
			FirstName: "user1",
			LastName:  "lastname1",
			Email:     "Other User <user1@test.com>",
			Password:  "12345678",
		},
		expectedError: domainuser.ErrInvalidEmail,
	}, {
		name: "empty password",
		input: CreateUserInput{
			FirstName: "user1",
			LastName:  "lastname1",
			Email:     "user1@test.com",
			Password:  "",
		},
		expectedError: domainuser.ErrInvalidPassword,
	}, {
		name: "password too short",
		input: CreateUserInput{
			FirstName: "user1",
			LastName:  "lastname1",
			Email:     "user1@test.com",
			Password:  "short7",
		},
		expectedError: domainuser.ErrPasswordTooShort,
	}}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			controller := gomock.NewController(t)
			mockRepository := domainusermocks.NewMockRepository(controller)
			createUserUseCase := NewCreateUserUseCase(mockRepository, bcrypt.MinCost)
			_, err := createUserUseCase.Execute(context.Background(), testCase.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, testCase.expectedError) {
				t.Fatalf("errors.Is: got %v, want %v", err, testCase.expectedError)
			}
		})
	}
}

func TestCreateUserUseCase_Execute_save_errors(t *testing.T) {
	t.Parallel()
	repositoryUnavailableError := errors.New("database unavailable")
	validInput := CreateUserInput{
		FirstName: "user1",
		LastName:  "lastname1",
		Email:     "user1@test.com",
		Password:  "12345678",
	}

	testCases := []struct {
		name               string
		saveReturnError    error
		expectedWrappedErr error
	}{
		{
			name:               "repository unavailable",
			saveReturnError:    repositoryUnavailableError,
			expectedWrappedErr: repositoryUnavailableError,
		},
		{
			name:               "duplicate email",
			saveReturnError:    fmt.Errorf("save user: %w", ErrEmailAlreadyExists),
			expectedWrappedErr: ErrEmailAlreadyExists,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			controller := gomock.NewController(t)
			mockRepository := domainusermocks.NewMockRepository(controller)
			mockRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(testCase.saveReturnError)
			createUserUseCase := NewCreateUserUseCase(mockRepository, bcrypt.MinCost)

			_, err := createUserUseCase.Execute(context.Background(), validInput)
			if !errors.Is(err, testCase.expectedWrappedErr) {
				t.Fatalf("errors.Is: got %v, want %v", err, testCase.expectedWrappedErr)
			}
		})
	}
}
