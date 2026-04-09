package transaction

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

const testTransactionDomainUserID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

func mustMinor(t *testing.T, v int64) MinorAmount {
	t.Helper()
	m, err := NewMinorAmount(v)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestTransaction_RejectsDebitIfInsufficient(t *testing.T) {
	t.Parallel()
	uid := uuid.MustParse(testTransactionDomainUserID)
	debit := NewTransaction(uid, TypeDebit, mustMinor(t, 50))

	t.Run("allows_when_balance_covers", func(t *testing.T) {
		t.Parallel()
		if err := debit.RejectsDebitIfInsufficient(100); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if err := debit.RejectsDebitIfInsufficient(50); err != nil {
			t.Fatalf("exact balance should allow: %v", err)
		}
	})

	t.Run("rejects_when_balance_short", func(t *testing.T) {
		t.Parallel()
		err := debit.RejectsDebitIfInsufficient(49)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrInsufficientBalance) {
			t.Fatalf("errors.Is: got %v, want %v", err, ErrInsufficientBalance)
		}
	})

	t.Run("credit_never_rejects", func(t *testing.T) {
		t.Parallel()
		credit := NewTransaction(uid, TypeCredit, mustMinor(t, 50))
		if err := credit.RejectsDebitIfInsufficient(0); err != nil {
			t.Fatalf("credit should not check balance: %v", err)
		}
	})
}
