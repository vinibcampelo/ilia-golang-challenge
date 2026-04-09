package transaction

import (
	"errors"
	"testing"
)

func TestNewMinorAmount_validation_errors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value int64
	}{
		{name: "zero", value: 0},
		{name: "negative", value: -1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewMinorAmount(testCase.value)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, ErrInvalidMinorAmount) {
				t.Fatalf("errors.Is: got %v, want %v", err, ErrInvalidMinorAmount)
			}
		})
	}
}

func TestNewMinorAmount_MinorUnits(t *testing.T) {
	t.Parallel()
	m, err := NewMinorAmount(1050)
	if err != nil {
		t.Fatalf("NewMinorAmount: %v", err)
	}
	if m.MinorUnits() != 1050 {
		t.Fatalf("MinorUnits: got %d want 1050", m.MinorUnits())
	}
}

func TestMinorAmount_CanCoverDebit(t *testing.T) {
	t.Parallel()
	m, err := NewMinorAmount(1050)
	if err != nil {
		t.Fatalf("NewMinorAmount: %v", err)
	}

	t.Run("sufficient_balance", func(t *testing.T) {
		t.Parallel()
		if !m.CanCoverDebit(1050) {
			t.Fatal("balance 1050 should cover debit 1050")
		}
		if !m.CanCoverDebit(2000) {
			t.Fatal("balance 2000 should cover debit 1050")
		}
	})

	t.Run("insufficient_balance", func(t *testing.T) {
		t.Parallel()
		if m.CanCoverDebit(1049) {
			t.Fatal("balance 1049 should not cover debit 1050")
		}
	})
}

func TestRehydrateMinorAmount(t *testing.T) {
	t.Parallel()
	m := RehydrateMinorAmount(99)
	if m.MinorUnits() != 99 {
		t.Fatalf("MinorUnits: got %d want 99", m.MinorUnits())
	}
}
