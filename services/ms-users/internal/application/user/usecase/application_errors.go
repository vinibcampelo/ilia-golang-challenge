package usecase

import (
	"errors"

	domainuser "ilia-golang-challenge/services/ms-users/internal/domain/user"
)

var ErrNotFound = domainuser.ErrNotFound
var ErrEmailAlreadyExists = domainuser.ErrEmailAlreadyExists

var ErrInvalidUserID = errors.New("invalid user id")
var ErrNoUpdateFields = errors.New("no fields to update")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrWalletHasNonZeroBalance = errors.New("wallet balance must be zero to delete account")
var ErrWalletServiceUnavailable = errors.New("wallet service unavailable")
