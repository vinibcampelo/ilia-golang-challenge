package user

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

const minPasswordRunes = 8

var (
	ErrInvalidFirstName   = errors.New("invalid first name")
	ErrInvalidLastName    = errors.New("invalid last name")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrPasswordTooShort   = errors.New("password does not meet minimum length")
	ErrEmailAlreadyExists = errors.New("email already registered")
)

type User struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	// Password is plaintext after NewUser; the create-user use case replaces it with a bcrypt hash before Repository.Save.
	Password string
}

func NewUser(id, rawFirstName, rawLastName, rawEmail, rawPassword string) (*User, error) {
	firstName := strings.TrimSpace(rawFirstName)
	if firstName == "" {
		return nil, ErrInvalidFirstName
	}
	lastName := strings.TrimSpace(rawLastName)
	if lastName == "" {
		return nil, ErrInvalidLastName
	}
	email := strings.TrimSpace(strings.ToLower(rawEmail))
	if email == "" {
		return nil, ErrInvalidEmail
	}
	if err := ensureBareEmailAddress(email); err != nil {
		return nil, ErrInvalidEmail
	}
	if rawPassword == "" {
		return nil, ErrInvalidPassword
	}
	if utf8.RuneCountInString(rawPassword) < minPasswordRunes {
		return nil, ErrPasswordTooShort
	}
	return &User{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Password:  rawPassword,
	}, nil
}

func ensureBareEmailAddress(email string) error {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return err
	}
	// Reject "Name <mailbox>" — registration accepts a single mailbox only.
	if addr.Name != "" {
		return errors.New("email must not include display name")
	}
	return nil
}
