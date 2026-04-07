package user

import (
	"errors"
	"testing"
)

const (
	testUserID    = "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	testFirstName = "user1"
	testLastName  = "lastname1"
	testEmail     = "user1@test.com"
)

func newFixtureUser() *User {
	return &User{
		ID:        testUserID,
		FirstName: testFirstName,
		LastName:  testLastName,
		Email:     testEmail,
		Password:  "stored-hash",
	}
}

func TestNew_success(t *testing.T) {
	t.Parallel()

	t.Run("preserves_id_and_plaintext_password", func(t *testing.T) {
		t.Parallel()
		plaintextPassword := "password123"
		userEntity, err := New(testUserID, testFirstName, testLastName, testEmail, plaintextPassword)
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if userEntity.ID != testUserID {
			t.Fatalf("ID: got %q want %q", userEntity.ID, testUserID)
		}
		if userEntity.FirstName != testFirstName || userEntity.LastName != testLastName {
			t.Fatalf("names: %+v", userEntity)
		}
		if userEntity.Email != testEmail {
			t.Fatalf("Email: got %q want %q", userEntity.Email, testEmail)
		}
		if userEntity.Password != plaintextPassword {
			t.Fatalf("Password: got %q want plaintext preserved", userEntity.Password)
		}
	})

	t.Run("normalizes_trimmed_names_email_and_accepts_minimum_password_length", func(t *testing.T) {
		t.Parallel()
		minimumLengthPassword := "12345678"
		userEntity, err := New(testUserID, "  user1\t", " lastname1 ", "  User1@TEST.COM  ", minimumLengthPassword)
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if userEntity.FirstName != testFirstName || userEntity.LastName != testLastName {
			t.Fatalf("trim: %+v", userEntity)
		}
		if userEntity.Email != testEmail {
			t.Fatalf("email normalize: got %q want %q", userEntity.Email, testEmail)
		}
	})

	t.Run("accepts_password_when_length_is_measured_in_runes_not_bytes", func(t *testing.T) {
		t.Parallel()
		eightRunePassword := "aaaaaaaé"
		if want, got := 8, len([]rune(eightRunePassword)); got != want {
			t.Fatalf("fixture rune count: got %d want %d", got, want)
		}
		userEntity, err := New(testUserID, testFirstName, testLastName, testEmail, eightRunePassword)
		if err != nil {
			t.Fatalf("8 runes should pass: %v", err)
		}
		if userEntity.Password != eightRunePassword {
			t.Fatalf("password: got %q", userEntity.Password)
		}
	})
}

func TestUser_Update_success(t *testing.T) {
	t.Parallel()

	t.Run("normalizes_first_name_and_email_and_leaves_other_fields_unchanged", func(t *testing.T) {
		t.Parallel()
		userEntity := newFixtureUser()
		userEntity.FirstName = "oldFirstName"
		userEntity.Email = "old@example.com"
		userEntity.Password = "unchanged-hash"
		firstName := "  newFirstName  "
		email := "NEW@EXAMPLE.COM"
		if err := userEntity.Update(UserUpdate{FirstName: &firstName, Email: &email}); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if userEntity.FirstName != "newFirstName" || userEntity.LastName != testLastName || userEntity.Email != "new@example.com" {
			t.Fatalf("unexpected user: %+v", userEntity)
		}
		if userEntity.Password != "unchanged-hash" {
			t.Fatal("password should be unchanged when not in patch")
		}
	})

	t.Run("trims_last_name_when_last_name_field_is_set", func(t *testing.T) {
		t.Parallel()
		userEntity := newFixtureUser()
		lastName := "  updatedLast  "
		if err := userEntity.Update(UserUpdate{LastName: &lastName}); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if userEntity.LastName != "updatedLast" {
			t.Fatalf("LastName: got %q", userEntity.LastName)
		}
	})

	t.Run("sets_plaintext_password_when_password_field_is_set", func(t *testing.T) {
		t.Parallel()
		userEntity := &User{ID: testUserID, FirstName: "a", LastName: "b", Email: "a@test.com", Password: "old"}
		newPassword := "12345678"
		if err := userEntity.Update(UserUpdate{Password: &newPassword}); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if userEntity.Password != newPassword {
			t.Fatalf("password: got %q want %q", userEntity.Password, newPassword)
		}
	})
}

func TestNew_validation_errors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		firstName     string
		lastName      string
		email         string
		password      string
		expectedError error
	}{
		{
			name: "empty_first_name", firstName: "", lastName: testLastName, email: testEmail, password: "12345678",
			expectedError: ErrInvalidFirstName,
		},
		{
			name: "whitespace_only_first_name", firstName: " \t\n", lastName: testLastName, email: testEmail, password: "12345678",
			expectedError: ErrInvalidFirstName,
		},
		{
			name: "empty_last_name", firstName: testFirstName, lastName: "", email: testEmail, password: "12345678",
			expectedError: ErrInvalidLastName,
		},
		{
			name: "whitespace_only_last_name", firstName: testFirstName, lastName: "   ", email: testEmail, password: "12345678",
			expectedError: ErrInvalidLastName,
		},
		{
			name: "empty_email", firstName: testFirstName, lastName: testLastName, email: "", password: "12345678",
			expectedError: ErrInvalidEmail,
		},
		{
			name: "email_only_spaces", firstName: testFirstName, lastName: testLastName, email: "  \t ", password: "12345678",
			expectedError: ErrInvalidEmail,
		},
		{
			name: "invalid_email", firstName: testFirstName, lastName: testLastName, email: "not-email", password: "12345678",
			expectedError: ErrInvalidEmail,
		},
		{
			name: "email_missing_domain", firstName: testFirstName, lastName: testLastName, email: "local@", password: "12345678",
			expectedError: ErrInvalidEmail,
		},
		{
			name: "display_name_form_not_allowed", firstName: testFirstName, lastName: testLastName, email: "Other User <user1@test.com>", password: "12345678",
			expectedError: ErrInvalidEmail,
		},
		{
			name: "empty_password", firstName: testFirstName, lastName: testLastName, email: testEmail, password: "",
			expectedError: ErrInvalidPassword,
		},
		{
			name: "password_seven_ascii_runes", firstName: testFirstName, lastName: testLastName, email: testEmail, password: "1234567",
			expectedError: ErrPasswordTooShort,
		},
		{
			name: "password_seven_runes_with_unicode", firstName: testFirstName, lastName: testLastName, email: testEmail, password: "aaaaaaé",
			expectedError: ErrPasswordTooShort,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			userEntity, err := New(testUserID, testCase.firstName, testCase.lastName, testCase.email, testCase.password)
			if err == nil {
				t.Fatalf("expected error, got user %+v", userEntity)
			}
			if !errors.Is(err, testCase.expectedError) {
				t.Fatalf("errors.Is: got %v want %v", err, testCase.expectedError)
			}
		})
	}
}

func TestUser_Update_validation_errors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		apply         func(*User) error
		expectedError error
	}{
		{
			name: "empty_first_name_returns_ErrInvalidFirstName",
			apply: func(userEntity *User) error {
				empty := ""
				return userEntity.Update(UserUpdate{FirstName: &empty})
			},
			expectedError: ErrInvalidFirstName,
		},
		{
			name: "whitespace_only_first_name_returns_ErrInvalidFirstName",
			apply: func(userEntity *User) error {
				onlySpace := " \t\n"
				return userEntity.Update(UserUpdate{FirstName: &onlySpace})
			},
			expectedError: ErrInvalidFirstName,
		},
		{
			name: "empty_last_name_returns_ErrInvalidLastName",
			apply: func(userEntity *User) error {
				empty := ""
				return userEntity.Update(UserUpdate{LastName: &empty})
			},
			expectedError: ErrInvalidLastName,
		},
		{
			name: "invalid_email_returns_ErrInvalidEmail",
			apply: func(userEntity *User) error {
				invalidEmail := "not-an-email"
				return userEntity.Update(UserUpdate{Email: &invalidEmail})
			},
			expectedError: ErrInvalidEmail,
		},
		{
			name: "empty_password_returns_ErrInvalidPassword",
			apply: func(userEntity *User) error {
				empty := ""
				return userEntity.Update(UserUpdate{Password: &empty})
			},
			expectedError: ErrInvalidPassword,
		},
		{
			name: "short_password_returns_ErrPasswordTooShort",
			apply: func(userEntity *User) error {
				short := "short"
				return userEntity.Update(UserUpdate{Password: &short})
			},
			expectedError: ErrPasswordTooShort,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			userEntity := newFixtureUser()
			err := testCase.apply(userEntity)
			if err == nil {
				t.Fatal("expected error")
			}
			if !errors.Is(err, testCase.expectedError) {
				t.Fatalf("errors.Is: got %v want %v", err, testCase.expectedError)
			}
		})
	}
}
