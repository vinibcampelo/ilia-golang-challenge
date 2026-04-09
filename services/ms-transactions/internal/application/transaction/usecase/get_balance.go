package usecase

import (
	"context"

	domaintransaction "ilia-golang-challenge/services/ms-transactions/internal/domain/transaction"
)

type GetBalanceUseCase struct {
	repo domaintransaction.Repository
}

func NewGetBalanceUseCase(repo domaintransaction.Repository) *GetBalanceUseCase {
	return &GetBalanceUseCase{repo: repo}
}

func (uc *GetBalanceUseCase) Execute(ctx context.Context, subjectUserID string) (*BalanceView, error) {
	amount, err := uc.repo.GetBalance(ctx, subjectUserID)
	if err != nil {
		return nil, err
	}
	return &BalanceView{Amount: amount}, nil
}
