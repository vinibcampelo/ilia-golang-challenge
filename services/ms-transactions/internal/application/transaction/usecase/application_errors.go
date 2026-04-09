package usecase

import "errors"

var (
	ErrInvalidUserID          = errors.New("invalid user id")
	ErrForbiddenUser          = errors.New("forbidden")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrInvalidTransactionType = errors.New("invalid transaction type")
	ErrInvalidAmount          = errors.New("amount must be positive")
	ErrInvalidTypeQuery       = errors.New("invalid transaction type filter")
)
