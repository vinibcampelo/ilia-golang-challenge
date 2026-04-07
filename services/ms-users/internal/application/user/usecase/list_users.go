package usecase

import (
	"context"
	"fmt"

	domainuser "ilia-golang-challenge/services/ms-users/internal/domain/user"
)

type ListUsersUseCase struct {
	repository domainuser.Repository
}

func NewListUsersUseCase(repository domainuser.Repository) *ListUsersUseCase {
	return &ListUsersUseCase{repository: repository}
}

func (useCase *ListUsersUseCase) Execute(ctx context.Context) ([]UserView, error) {
	users, err := useCase.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	userViews := make([]UserView, 0, len(users))
	for userIndex := range users {
		user := &users[userIndex]
		userViews = append(userViews, UserView{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		})
	}
	return userViews, nil
}
