package user

import (
	"errors"
	"testing"
)

const testUserID = "test-user-1"

func TestNewUser_success_preservesIDAndPassword(t *testing.T) {
	t.Parallel()
	u, err := NewUser(testUserID, "user1", "lastname1", "user1@test.com", "password123")
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	if u.ID != testUserID {
		t.Fatalf("ID: got %q want %q", u.ID, testUserID)
	}
	if u.FirstName != "user1" || u.LastName != "lastname1" {
		t.Fatalf("names: %+v", u)
	}
	if u.Email != "user1@test.com" {
		t.Fatalf("Email: got %q", u.Email)
	}
	if u.Password != "password123" {
		t.Fatalf("Password not preserved")
	}
}

func TestNewUser_success_normalizesInput(t *testing.T) {
	t.Parallel()
	u, err := NewUser(testUserID, "  user1\t", " lastname1 ", "  User1@TEST.COM  ", "12345678")
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	if u.FirstName != "user1" || u.LastName != "lastname1" {
		t.Fatalf("trim failed: %+v", u)
	}
	if u.Email != "user1@test.com" {
		t.Fatalf("email normalize: got %q want user1@test.com", u.Email)
	}
}

func TestNewUser_passwordLengthByRunes_notBytes(t *testing.T) {
	t.Parallel()
	password := "aaaaaaaé"
	if want, got := 8, len([]rune(password)); got != want {
		t.Fatalf("test string rune count: got %d want %d", got, want)
	}
	u, err := NewUser(testUserID, "user1", "lastname1", "user1@test.com", password)
	if err != nil {
		t.Fatalf("8 runes should pass: %v", err)
	}
	if u.Password != password {
		t.Fatalf("password mismatch")
	}
}

func TestNewUser_passwordExactlyMinRunes(t *testing.T) {
	t.Parallel()
	_, err := NewUser(testUserID, "user1", "lastname1", "user1@test.com", "12345678")
	if err != nil {
		t.Fatalf("8 ASCII runes: %v", err)
	}
}

func TestNewUser_validationErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		first, last string
		email       string
		password    string
		wantErr     error
	}{
		{
			name: "empty first name", first: "", last: "lastname1", email: "user1@test.com", password: "12345678",
			wantErr: ErrInvalidFirstName,
		},
		{
			name: "whitespace first name", first: " \t\n", last: "lastname1", email: "user1@test.com", password: "12345678",
			wantErr: ErrInvalidFirstName,
		},
		{
			name: "empty last name", first: "user1", last: "", email: "user1@test.com", password: "12345678",
			wantErr: ErrInvalidLastName,
		},
		{
			name: "whitespace last name", first: "user1", last: "   ", email: "user1@test.com", password: "12345678",
			wantErr: ErrInvalidLastName,
		},
		{
			name: "empty email", first: "user1", last: "lastname1", email: "", password: "12345678",
			wantErr: ErrInvalidEmail,
		},
		{
			name: "email only spaces", first: "user1", last: "lastname1", email: "  \t ", password: "12345678",
			wantErr: ErrInvalidEmail,
		},
		{
			name: "invalid email", first: "user1", last: "lastname1", email: "not-email", password: "12345678",
			wantErr: ErrInvalidEmail,
		},
		{
			name: "email missing domain", first: "user1", last: "lastname1", email: "local@", password: "12345678",
			wantErr: ErrInvalidEmail,
		},
		{
			name: "display name form", first: "user1", last: "lastname1", email: "Other User <user1@test.com>", password: "12345678",
			wantErr: ErrInvalidEmail,
		},
		{
			name: "empty password", first: "user1", last: "lastname1", email: "user1@test.com", password: "",
			wantErr: ErrInvalidPassword,
		},
		{
			name: "password seven runes", first: "user1", last: "lastname1", email: "user1@test.com", password: "1234567",
			wantErr: ErrPasswordTooShort,
		},
		{
			name: "password seven runes unicode", first: "user1", last: "lastname1", email: "user1@test.com", password: "aaaaaaé",
			wantErr: ErrPasswordTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u, err := NewUser(testUserID, tt.first, tt.last, tt.email, tt.password)
			if err == nil {
				t.Fatalf("expected error, got user %+v", u)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}
