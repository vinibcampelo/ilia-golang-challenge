package user

import "errors"

var ErrNotFound = errors.New("user not found")
var ErrEmailAlreadyExists = errors.New("email already registered")
