package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainuser "ilia-golang-challenge/services/ms-users/internal/domain/user"
)

type DeleteUserUseCase struct {
	repository   domainuser.Repository
	walletGate   WalletZeroBalanceChecker
}

func NewDeleteUserUseCase(repository domainuser.Repository, walletGate WalletZeroBalanceChecker) *DeleteUserUseCase {
	return &DeleteUserUseCase{repository: repository, walletGate: walletGate}
}

func (useCase *DeleteUserUseCase) Execute(ctx context.Context, userID string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("delete user: %w", ErrInvalidUserID)
	}
	if err := useCase.walletGate.AssertZeroBalance(ctx, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if err := useCase.repository.DeleteByID(ctx, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
