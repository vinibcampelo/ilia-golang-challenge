package transaction

import (
	"context"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=repository.go -destination=mocks/repository_mock.go -package=mocks

type Repository interface {
	Create(ctx context.Context, tr Transaction) (*Transaction, error)
	ListByUser(ctx context.Context, userID string, filterType *Type) ([]Transaction, error)
	GetBalance(ctx context.Context, userID string) (int64, error)
}
