package usecase

import (
	"context"

	domaintransaction "ilia-golang-challenge/services/ms-transactions/internal/domain/transaction"
)

type ListTransactionsUseCase struct {
	repo domaintransaction.Repository
}

func NewListTransactionsUseCase(repo domaintransaction.Repository) *ListTransactionsUseCase {
	return &ListTransactionsUseCase{repo: repo}
}

func (uc *ListTransactionsUseCase) Execute(ctx context.Context, subjectUserID, typeQuery string) ([]TransactionView, error) {
	var filter *domaintransaction.Type
	if typeQuery != "" {
		t, err := domaintransaction.ParseType(typeQuery)
		if err != nil {
			return nil, ErrInvalidTypeQuery
		}
		filter = &t
	}

	rows, err := uc.repo.ListByUser(ctx, subjectUserID, filter)
	if err != nil {
		return nil, err
	}
	out := make([]TransactionView, 0, len(rows))
	for _, row := range rows {
		out = append(out, TransactionView{
			ID:     row.ID.String(),
			UserID: row.UserID.String(),
			Type:   row.Type.String(),
			Amount: row.Amount.MinorUnits(),
		})
	}
	return out, nil
}
