package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainuser "ilia-golang-challenge/ms-users/internal/domain/user"
)

type DeleteUserUseCase struct {
	repository domainuser.Repository
}

func NewDeleteUserUseCase(repository domainuser.Repository) *DeleteUserUseCase {
	return &DeleteUserUseCase{repository: repository}
}

func (useCase *DeleteUserUseCase) Execute(ctx context.Context, userID string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("delete user: %w", ErrInvalidUserID)
	}
	if err := useCase.repository.DeleteByID(ctx, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
