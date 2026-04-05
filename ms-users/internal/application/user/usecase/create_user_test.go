package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	domainuser "ilia-golang-challenge/ms-users/internal/domain/user"
)

type stubUserRepository struct {
	saveErr error
	saved   []*domainuser.User
}

func (s *stubUserRepository) Save(ctx context.Context, entity *domainuser.User) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saved = append(s.saved, entity)
	return nil
}

func TestCreateUserUseCase_Execute_success_normalizesAndPersists(t *testing.T) {
	t.Parallel()
	repo := &stubUserRepository{}
	createUserUseCase := NewCreateUserUseCase(repo, bcrypt.MinCost)

	out, err := createUserUseCase.Execute(context.Background(), CreateUserInput{
		FirstName: "  user1  ",
		LastName:  " lastname1 ",
		Email:     "User1@TEST.COM",
		Password:  "password123",
	})

	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := uuid.Parse(out.ID); err != nil {
		t.Fatalf("output.ID is not a valid UUID: %q", out.ID)
	}
	if out.FirstName != "user1" || out.LastName != "lastname1" || out.Email != "user1@test.com" {
		t.Fatalf("unexpected output: %+v", out)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("Save calls: got %d want 1", len(repo.saved))
	}
	got := repo.saved[0]
	if got.ID != out.ID {
		t.Fatalf("saved ID: got %q want %q", got.ID, out.ID)
	}
	if got.FirstName != "user1" || got.LastName != "lastname1" || got.Email != "user1@test.com" {
		t.Fatalf("saved entity: %+v", got)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(got.Password), []byte("password123")); err != nil {
		t.Fatalf("saved password should be bcrypt of plaintext: %v", err)
	}
}

func TestCreateUserUseCase_Execute_passwordMinLengthBoundary(t *testing.T) {
	t.Parallel()
	repo := &stubUserRepository{}
	createUserUseCase := NewCreateUserUseCase(repo, bcrypt.MinCost)

	_, err := createUserUseCase.Execute(context.Background(), CreateUserInput{
		FirstName: "user1",
		LastName:  "lastname1",
		Email:     "user1@test.com",
		Password:  "12345678",
	})
	if err != nil {
		t.Fatalf("8-rune password should pass: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected one Save, got %d", len(repo.saved))
	}
}

func TestCreateUserUseCase_Execute_domainErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   CreateUserInput
		wantErr error
	}{
		{
			name: "empty first name",
			input: CreateUserInput{
				FirstName: "",
				LastName:  "lastname1",
				Email:     "user1@test.com",
				Password:  "12345678",
			},
			wantErr: domainuser.ErrInvalidFirstName,
		},
		{
			name: "whitespace only first name",
			input: CreateUserInput{
				FirstName: "   \t",
				LastName:  "lastname1",
				Email:     "user1@test.com",
				Password:  "12345678",
			},
			wantErr: domainuser.ErrInvalidFirstName,
		},
		{
			name: "empty last name",
			input: CreateUserInput{
				FirstName: "user1",
				LastName:  "",
				Email:     "user1@test.com",
				Password:  "12345678",
			},
			wantErr: domainuser.ErrInvalidLastName,
		},
		{
			name: "empty email",
			input: CreateUserInput{
				FirstName: "user1",
				LastName:  "lastname1",
				Email:     "",
				Password:  "12345678",
			},
			wantErr: domainuser.ErrInvalidEmail,
		},
		{
			name: "invalid email format",
			input: CreateUserInput{
				FirstName: "user1",
				LastName:  "lastname1",
				Email:     "not-an-email",
				Password:  "12345678",
			},
			wantErr: domainuser.ErrInvalidEmail,
		},
		{
			name: "email with display name",
			input: CreateUserInput{
				FirstName: "user1",
				LastName:  "lastname1",
				Email:     "Other User <user1@test.com>",
				Password:  "12345678",
			},
			wantErr: domainuser.ErrInvalidEmail,
		},
		{
			name: "empty password",
			input: CreateUserInput{
				FirstName: "user1",
				LastName:  "lastname1",
				Email:     "user1@test.com",
				Password:  "",
			},
			wantErr: domainuser.ErrInvalidPassword,
		},
		{
			name: "password too short",
			input: CreateUserInput{
				FirstName: "user1",
				LastName:  "lastname1",
				Email:     "user1@test.com",
				Password:  "short7",
			},
			wantErr: domainuser.ErrPasswordTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &stubUserRepository{}
			createUserUseCase := NewCreateUserUseCase(repo, bcrypt.MinCost)
			_, err := createUserUseCase.Execute(context.Background(), tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("errors.Is: got %v, want %v", err, tt.wantErr)
			}
			if len(repo.saved) != 0 {
				t.Fatalf("Save should not be called on domain failure, saved: %d", len(repo.saved))
			}
		})
	}
}

func TestCreateUserUseCase_Execute_repositorySaveError(t *testing.T) {
	t.Parallel()
	saveErr := errors.New("database unavailable")
	repo := &stubUserRepository{saveErr: saveErr}
	createUserUseCase := NewCreateUserUseCase(repo, bcrypt.MinCost)

	_, err := createUserUseCase.Execute(context.Background(), CreateUserInput{
		FirstName: "user1",
		LastName:  "lastname1",
		Email:     "user1@test.com",
		Password:  "12345678",
	})
	if !errors.Is(err, saveErr) {
		t.Fatalf("got %v, want %v", err, saveErr)
	}
	if len(repo.saved) != 0 {
		t.Fatalf("Save should not append on error")
	}
}

func TestCreateUserUseCase_Execute_duplicateEmail(t *testing.T) {
	t.Parallel()
	repo := &stubUserRepository{saveErr: fmt.Errorf("save user: %w", domainuser.ErrEmailAlreadyExists)}
	createUserUseCase := NewCreateUserUseCase(repo, bcrypt.MinCost)

	_, err := createUserUseCase.Execute(context.Background(), CreateUserInput{
		FirstName: "user1",
		LastName:  "lastname1",
		Email:     "user1@test.com",
		Password:  "12345678",
	})
	if !errors.Is(err, domainuser.ErrEmailAlreadyExists) {
		t.Fatalf("errors.Is: got %v, want %v", err, domainuser.ErrEmailAlreadyExists)
	}
}
