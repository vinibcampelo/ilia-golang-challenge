package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainuser "ilia-golang-challenge/ms-users/internal/domain/user"
)

type GetUserUseCase struct {
	repository domainuser.Repository
}

func NewGetUserUseCase(repository domainuser.Repository) *GetUserUseCase {
	return &GetUserUseCase{repository: repository}
}

func (useCase *GetUserUseCase) Execute(ctx context.Context, userID string) (*UserView, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, fmt.Errorf("get user: %w", ErrInvalidUserID)
	}
	foundUser, err := useCase.repository.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &UserView{
		ID:        foundUser.ID,
		FirstName: foundUser.FirstName,
		LastName:  foundUser.LastName,
		Email:     foundUser.Email,
	}, nil
}
