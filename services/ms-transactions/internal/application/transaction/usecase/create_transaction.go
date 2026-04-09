package usecase

import (
	"context"
	"errors"
	"fmt"

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
	SubjectUserID  string
	BodyUserID     string
	Type           string
	Amount         int64
	IdempotencyKey string
}

func TransactionRequestFingerprint(bodyUserID string, txType domaintransaction.Type, amountMinor int64) string {
	return fmt.Sprintf("%s|%s|%d", bodyUserID, txType.String(), amountMinor)
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
	repoIn := domaintransaction.RepositoryCreateInput{Transaction: pending}
	if in.IdempotencyKey != "" {
		repoIn.IdempotencyKey = in.IdempotencyKey
		repoIn.RequestFingerprint = TransactionRequestFingerprint(in.BodyUserID, txType, in.Amount)
	}

	created, err := uc.repo.Create(ctx, repoIn)
	if err != nil {
		if errors.Is(err, domaintransaction.ErrInsufficientBalance) {
			return nil, ErrInsufficientBalance
		}
		if errors.Is(err, domaintransaction.ErrIdempotencyConflict) {
			return nil, ErrIdempotencyConflict
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
