package usecase

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	transactionmocks "ilia-golang-challenge/services/ms-transactions/internal/domain/transaction/mocks"
)

func TestGetBalanceUseCase_Execute_success(t *testing.T) {
	t.Parallel()

	t.Run("returns_balance_view_in_minor_units", func(t *testing.T) {
		t.Parallel()
		subject := testTransactionSubjectUserID
		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		mockRepository.EXPECT().GetBalance(gomock.Any(), subject).Return(int64(42), nil)

		useCase := NewGetBalanceUseCase(mockRepository)
		view, err := useCase.Execute(context.Background(), subject)
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if view.Amount != 42 {
			t.Fatalf("Amount: got %d want 42", view.Amount)
		}
	})
}

func TestGetBalanceUseCase_Execute_errors(t *testing.T) {
	t.Parallel()

	t.Run("repository_error_is_propagated", func(t *testing.T) {
		t.Parallel()
		repositoryError := errors.New("connection reset")
		subject := testTransactionSubjectUserID

		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		mockRepository.EXPECT().GetBalance(gomock.Any(), subject).Return(int64(0), repositoryError)

		useCase := NewGetBalanceUseCase(mockRepository)
		_, err := useCase.Execute(context.Background(), subject)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, repositoryError) {
			t.Fatalf("errors.Is: got %v, want %v", err, repositoryError)
		}
	})
}
