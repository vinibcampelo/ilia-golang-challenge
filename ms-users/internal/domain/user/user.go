package user

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

const minPasswordLengthRunes = 8

var (
	ErrInvalidFirstName = errors.New("invalid first name")
	ErrInvalidLastName  = errors.New("invalid last name")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrPasswordTooShort = errors.New("password does not meet minimum length")
)

type User struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Password  string
}

type UserUpdate struct {
	FirstName *string
	LastName  *string
	Email     *string
	Password  *string
}

func (update UserUpdate) IsEmpty() bool {
	return update.FirstName == nil && update.LastName == nil && update.Email == nil && update.Password == nil
}

func New(userID, rawFirstName, rawLastName, rawEmail, rawPassword string) (*User, error) {
	firstName, err := TrimmedNonEmptyName(rawFirstName, ErrInvalidFirstName)
	if err != nil {
		return nil, err
	}
	lastName, err := TrimmedNonEmptyName(rawLastName, ErrInvalidLastName)
	if err != nil {
		return nil, err
	}
	email, err := NormalizeEmail(rawEmail)
	if err != nil {
		return nil, err
	}
	if err := ValidatePasswordLength(rawPassword); err != nil {
		return nil, err
	}
	return &User{
		ID:        userID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Password:  rawPassword,
	}, nil
}

func (user *User) Update(changes UserUpdate) error {
	if changes.FirstName != nil {
		firstName, err := TrimmedNonEmptyName(*changes.FirstName, ErrInvalidFirstName)
		if err != nil {
			return err
		}
		user.FirstName = firstName
	}
	if changes.LastName != nil {
		lastName, err := TrimmedNonEmptyName(*changes.LastName, ErrInvalidLastName)
		if err != nil {
			return err
		}
		user.LastName = lastName
	}
	if changes.Email != nil {
		email, err := NormalizeEmail(*changes.Email)
		if err != nil {
			return err
		}
		user.Email = email
	}
	if changes.Password != nil {
		if err := ValidatePasswordLength(*changes.Password); err != nil {
			return err
		}
		user.Password = *changes.Password
	}
	return nil
}

func ensureBareEmailAddress(mailbox string) error {
	parsedAddress, err := mail.ParseAddress(mailbox)
	if err != nil {
		return err
	}
	if parsedAddress.Name != "" {
		return errors.New("email must not include display name")
	}
	return nil
}

func NormalizeEmail(raw string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" {
		return "", ErrInvalidEmail
	}
	if err := ensureBareEmailAddress(normalized); err != nil {
		return "", ErrInvalidEmail
	}
	return normalized, nil
}

func TrimmedNonEmptyName(raw string, emptyErr error) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", emptyErr
	}
	return trimmed, nil
}

func ValidatePasswordLength(raw string) error {
	if raw == "" {
		return ErrInvalidPassword
	}
	if utf8.RuneCountInString(raw) < minPasswordLengthRunes {
		return ErrPasswordTooShort
	}
	return nil
}
