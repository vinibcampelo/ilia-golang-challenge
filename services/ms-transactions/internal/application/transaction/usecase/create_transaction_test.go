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

const (
	testTransactionSubjectUserID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	testTransactionOtherUserID   = "7c9e6679-7425-40de-944b-e07fc1f90ae7"
)

func mustMinor(t *testing.T, v int64) domaintransaction.MinorAmount {
	t.Helper()
	m, err := domaintransaction.NewMinorAmount(v)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestCreateTransactionUseCase_Execute_success(t *testing.T) {
	t.Run("credit_persists_via_repository_and_returns_view", func(t *testing.T) {
		t.Parallel()
		subject := testTransactionSubjectUserID
		uid := uuid.MustParse(subject)
		pending := domaintransaction.NewTransaction(uid, domaintransaction.TypeCredit, mustMinor(t, 50))

		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		mockRepository.EXPECT().Create(gomock.Any(), pending).
			Return(&domaintransaction.Transaction{
				ID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				UserID: uid,
				Type:   domaintransaction.TypeCredit,
				Amount: mustMinor(t, 50),
			}, nil)

		useCase := NewCreateTransactionUseCase(mockRepository)
		view, err := useCase.Execute(context.Background(), CreateTransactionInput{
			SubjectUserID: subject,
			BodyUserID:    subject,
			Type:          "CREDIT",
			Amount:        50,
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if view.Amount != 50 || view.Type != "CREDIT" || view.UserID != subject {
			t.Fatalf("unexpected view: %+v", view)
		}
	})
}

func TestCreateTransactionUseCase_Execute_validation_errors(t *testing.T) {
	t.Parallel()
	subject := testTransactionSubjectUserID

	testCases := []struct {
		name          string
		input         CreateTransactionInput
		expectedError error
	}{
		{
			name: "invalid body user id",
			input: CreateTransactionInput{
				SubjectUserID: subject,
				BodyUserID:    "not-a-uuid",
				Type:          "CREDIT",
				Amount:        10,
			},
			expectedError: ErrInvalidUserID,
		},
		{
			name: "forbidden user id mismatch",
			input: CreateTransactionInput{
				SubjectUserID: subject,
				BodyUserID:    testTransactionOtherUserID,
				Type:          "CREDIT",
				Amount:        10,
			},
			expectedError: ErrForbiddenUser,
		},
		{
			name: "invalid transaction type",
			input: CreateTransactionInput{
				SubjectUserID: subject,
				BodyUserID:    subject,
				Type:          "WIRED",
				Amount:        10,
			},
			expectedError: ErrInvalidTransactionType,
		},
		{
			name: "non positive amount",
			input: CreateTransactionInput{
				SubjectUserID: subject,
				BodyUserID:    subject,
				Type:          "CREDIT",
				Amount:        0,
			},
			expectedError: ErrInvalidAmount,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			controller := gomock.NewController(t)
			mockRepository := transactionmocks.NewMockRepository(controller)
			useCase := NewCreateTransactionUseCase(mockRepository)

			_, err := useCase.Execute(context.Background(), testCase.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, testCase.expectedError) {
				t.Fatalf("errors.Is: got %v, want %v", err, testCase.expectedError)
			}
		})
	}
}

func TestCreateTransactionUseCase_Execute_errors(t *testing.T) {
	t.Parallel()
	subject := testTransactionSubjectUserID
	uid := uuid.MustParse(subject)

	t.Run("insufficient_balance_returns_ErrInsufficientBalance", func(t *testing.T) {
		t.Parallel()
		pending := domaintransaction.NewTransaction(uid, domaintransaction.TypeDebit, mustMinor(t, 100))

		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		mockRepository.EXPECT().Create(gomock.Any(), pending).
			Return(nil, domaintransaction.ErrInsufficientBalance)

		useCase := NewCreateTransactionUseCase(mockRepository)
		_, err := useCase.Execute(context.Background(), CreateTransactionInput{
			SubjectUserID: subject,
			BodyUserID:    subject,
			Type:          "DEBIT",
			Amount:        100,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrInsufficientBalance) {
			t.Fatalf("errors.Is: got %v, want %v", err, ErrInsufficientBalance)
		}
	})

	t.Run("repository_error_is_propagated", func(t *testing.T) {
		t.Parallel()
		repositoryError := errors.New("database unavailable")
		pending := domaintransaction.NewTransaction(uid, domaintransaction.TypeCredit, mustMinor(t, 1))

		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		mockRepository.EXPECT().Create(gomock.Any(), pending).Return(nil, repositoryError)

		useCase := NewCreateTransactionUseCase(mockRepository)
		_, err := useCase.Execute(context.Background(), CreateTransactionInput{
			SubjectUserID: subject,
			BodyUserID:    subject,
			Type:          "CREDIT",
			Amount:        1,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, repositoryError) {
			t.Fatalf("errors.Is: got %v, want %v", err, repositoryError)
		}
	})
}
