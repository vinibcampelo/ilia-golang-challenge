package usecase

import "errors"

// ErrInvalidUserID is returned when a path id is not a valid UUID.
var ErrInvalidUserID = errors.New("invalid user id")

// ErrNoUpdateFields is returned when a PATCH body omits every updatable field.
var ErrNoUpdateFields = errors.New("no fields to update")

// ErrNotFound is returned when a user record does not exist for this application (repository contract).
var ErrNotFound = errors.New("user not found")

// ErrEmailAlreadyExists is returned when save or update would violate unique email.
var ErrEmailAlreadyExists = errors.New("email already registered")

// ErrInvalidCredentials is returned when login email/password do not match an active user.
var ErrInvalidCredentials = errors.New("invalid credentials")
