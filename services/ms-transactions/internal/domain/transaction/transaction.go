package transaction

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrIdempotencyConflict = errors.New("idempotency key reused with a different request body")
)

type Transaction struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      Type
	Amount    MinorAmount
	CreatedAt time.Time
}

func NewTransaction(userID uuid.UUID, txType Type, amount MinorAmount) Transaction {
	return Transaction{
		UserID: userID,
		Type:   txType,
		Amount: amount,
	}
}

func (t Transaction) RejectsDebitIfInsufficient(balanceMinorUnits int64) error {
	if !t.Type.IsDebit() {
		return nil
	}
	if !t.Amount.CanCoverDebit(balanceMinorUnits) {
		return ErrInsufficientBalance
	}
	return nil
}
