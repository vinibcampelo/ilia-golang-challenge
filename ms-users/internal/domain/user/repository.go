package user

import "context"

type Repository interface {
	Save(ctx context.Context, entity *User) error
}
