package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domaintransaction "ilia-golang-challenge/services/ms-transactions/internal/domain/transaction"
)

type CreateTransactionUseCase struct {
	repo domaintransaction.Repository
}

func NewCreateTransactionUseCase(repo domaintransaction.Repository) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{repo: repo}
}

type CreateTransactionInput struct {
	SubjectUserID string
	BodyUserID    string
	Type          string
	Amount        int64
}

func (uc *CreateTransactionUseCase) Execute(ctx context.Context, in CreateTransactionInput) (*TransactionView, error) {
	userUUID, err := uuid.Parse(in.BodyUserID)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	if in.BodyUserID != in.SubjectUserID {
		return nil, ErrForbiddenUser
	}
	txType, err := domaintransaction.ParseType(in.Type)
	if err != nil {
		if errors.Is(err, domaintransaction.ErrInvalidType) {
			return nil, ErrInvalidTransactionType
		}
		return nil, err
	}
	minor, err := domaintransaction.NewMinorAmount(in.Amount)
	if err != nil {
		if errors.Is(err, domaintransaction.ErrInvalidMinorAmount) {
			return nil, ErrInvalidAmount
		}
		return nil, err
	}

	pending := domaintransaction.NewTransaction(userUUID, txType, minor)
	created, err := uc.repo.Create(ctx, pending)
	if err != nil {
		if errors.Is(err, domaintransaction.ErrInsufficientBalance) {
			return nil, ErrInsufficientBalance
		}
		return nil, err
	}

	return &TransactionView{
		ID:     created.ID.String(),
		UserID: created.UserID.String(),
		Type:   created.Type.String(),
		Amount: created.Amount.MinorUnits(),
	}, nil
}
