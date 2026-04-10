package usecase

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	domainusermocks "ilia-golang-challenge/services/ms-users/internal/domain/user/mocks"
	usecasemocks "ilia-golang-challenge/services/ms-users/internal/application/user/usecase/mocks"
)

func TestDeleteUserUseCase_Execute_success(t *testing.T) {
	t.Parallel()

	t.Run("deletes_user", func(t *testing.T) {
		t.Parallel()
		userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockWallet := usecasemocks.NewMockWalletZeroBalanceChecker(controller)
		mockWallet.EXPECT().AssertZeroBalance(gomock.Any(), userID).Return(nil)
		mockRepository.EXPECT().DeleteByID(gomock.Any(), userID).Return(nil)
		useCase := NewDeleteUserUseCase(mockRepository, mockWallet)

		err := useCase.Execute(context.Background(), userID)
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
	})
}

func TestDeleteUserUseCase_Execute_errors(t *testing.T) {
	t.Parallel()

	t.Run("invalid_user_id_returns_ErrInvalidUserID", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockWallet := usecasemocks.NewMockWalletZeroBalanceChecker(controller)
		useCase := NewDeleteUserUseCase(mockRepository, mockWallet)

		err := useCase.Execute(context.Background(), "bad-id")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrInvalidUserID) {
			t.Fatalf("errors.Is: got %v want %v", err, ErrInvalidUserID)
		}
	})

	t.Run("repository_returns_not_found_is_propagated", func(t *testing.T) {
		t.Parallel()
		userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockWallet := usecasemocks.NewMockWalletZeroBalanceChecker(controller)
		mockWallet.EXPECT().AssertZeroBalance(gomock.Any(), userID).Return(nil)
		mockRepository.EXPECT().DeleteByID(gomock.Any(), userID).Return(ErrNotFound)
		useCase := NewDeleteUserUseCase(mockRepository, mockWallet)

		err := useCase.Execute(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("errors.Is: got %v want %v", err, ErrNotFound)
		}
	})

	t.Run("wallet_non_zero_balance_returns_ErrWalletHasNonZeroBalance", func(t *testing.T) {
		t.Parallel()
		userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockWallet := usecasemocks.NewMockWalletZeroBalanceChecker(controller)
		mockWallet.EXPECT().AssertZeroBalance(gomock.Any(), userID).Return(ErrWalletHasNonZeroBalance)
		useCase := NewDeleteUserUseCase(mockRepository, mockWallet)

		err := useCase.Execute(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrWalletHasNonZeroBalance) {
			t.Fatalf("errors.Is: got %v want %v", err, ErrWalletHasNonZeroBalance)
		}
	})

	t.Run("wallet_unavailable_returns_ErrWalletServiceUnavailable", func(t *testing.T) {
		t.Parallel()
		userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockWallet := usecasemocks.NewMockWalletZeroBalanceChecker(controller)
		mockWallet.EXPECT().AssertZeroBalance(gomock.Any(), userID).Return(ErrWalletServiceUnavailable)
		useCase := NewDeleteUserUseCase(mockRepository, mockWallet)

		err := useCase.Execute(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrWalletServiceUnavailable) {
			t.Fatalf("errors.Is: got %v want %v", err, ErrWalletServiceUnavailable)
		}
	})
}
