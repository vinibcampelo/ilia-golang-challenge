package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	domaintransaction "ilia-golang-challenge/services/ms-transactions/internal/domain/transaction"
	transactionmocks "ilia-golang-challenge/services/ms-transactions/internal/domain/transaction/mocks"
)

func TestListTransactionsUseCase_Execute_success(t *testing.T) {
	t.Parallel()

	t.Run("maps_transactions_to_views", func(t *testing.T) {
		t.Parallel()
		subject := testTransactionSubjectUserID
		uid := uuid.MustParse(subject)
		credit := mustMinor(t, 10)
		debit := mustMinor(t, 5)
		stored := []domaintransaction.Transaction{
			{
				ID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				UserID: uid,
				Type:   domaintransaction.TypeCredit,
				Amount: credit,
			},
			{
				ID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				UserID: uid,
				Type:   domaintransaction.TypeDebit,
				Amount: debit,
			},
		}

		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		mockRepository.EXPECT().ListByUser(gomock.Any(), subject, gomock.Nil()).Return(stored, nil)

		useCase := NewListTransactionsUseCase(mockRepository)
		views, err := useCase.Execute(context.Background(), subject, "")
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if len(views) != 2 {
			t.Fatalf("len: got %d want 2", len(views))
		}
		if views[0].Type != "CREDIT" || views[0].Amount != 10 {
			t.Fatalf("first view: %+v", views[0])
		}
		if views[1].Type != "DEBIT" || views[1].Amount != 5 {
			t.Fatalf("second view: %+v", views[1])
		}
	})

	t.Run("empty_repository_returns_empty_non_nil_slice", func(t *testing.T) {
		t.Parallel()
		subject := testTransactionSubjectUserID
		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		mockRepository.EXPECT().ListByUser(gomock.Any(), subject, gomock.Nil()).
			Return([]domaintransaction.Transaction{}, nil)

		useCase := NewListTransactionsUseCase(mockRepository)
		views, err := useCase.Execute(context.Background(), subject, "")
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if views == nil || len(views) != 0 {
			t.Fatalf("want empty non-nil slice, got %#v", views)
		}
	})
}

func TestListTransactionsUseCase_Execute_errors(t *testing.T) {
	t.Parallel()

	t.Run("invalid_type_query_returns_ErrInvalidTypeQuery", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		useCase := NewListTransactionsUseCase(mockRepository)

		_, err := useCase.Execute(context.Background(), testTransactionSubjectUserID, "FOO")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidTypeQuery) {
			t.Fatalf("errors.Is: got %v, want %v", err, ErrInvalidTypeQuery)
		}
	})

	t.Run("repository_error_is_propagated", func(t *testing.T) {
		t.Parallel()
		repositoryError := errors.New("db down")
		subject := testTransactionSubjectUserID
		credit := domaintransaction.TypeCredit

		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		mockRepository.EXPECT().ListByUser(gomock.Any(), subject, &credit).Return(nil, repositoryError)

		useCase := NewListTransactionsUseCase(mockRepository)
		_, err := useCase.Execute(context.Background(), subject, "CREDIT")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, repositoryError) {
			t.Fatalf("errors.Is: got %v, want %v", err, repositoryError)
		}
	})
}
