package user

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=repository.go -destination=mocks/repository_mock.go -package=mocks

import "context"

type Repository interface {
	Save(ctx context.Context, entity *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	List(ctx context.Context) ([]User, error)
	Update(ctx context.Context, entity *User) error
	DeleteByID(ctx context.Context, id string) error
}
