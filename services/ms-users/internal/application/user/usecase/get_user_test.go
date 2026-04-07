package usecase

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	domainuser "ilia-golang-challenge/services/ms-users/internal/domain/user"
	domainusermocks "ilia-golang-challenge/services/ms-users/internal/domain/user/mocks"
)

const (
	testUserID    = "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	testFirstName = "user1"
	testLastName  = "lastname1"
	testEmail     = "user1@test.com"
)

func TestGetUserUseCase_Execute_success(t *testing.T) {
	t.Parallel()

	t.Run("returns_user_view_without_password", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockRepository.EXPECT().FindByID(gomock.Any(), testUserID).Return(&domainuser.User{
			ID:        testUserID,
			FirstName: testFirstName,
			LastName:  testLastName,
			Email:     testEmail,
			Password:  "bcrypt-hash",
		}, nil)
		useCase := NewGetUserUseCase(mockRepository)

		userView, err := useCase.Execute(context.Background(), testUserID)
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if userView.ID != testUserID || userView.FirstName != testFirstName || userView.LastName != testLastName || userView.Email != testEmail {
			t.Fatalf("unexpected view: %+v", userView)
		}
	})
}

func TestGetUserUseCase_Execute_errors(t *testing.T) {
	t.Parallel()

	t.Run("invalid_user_id_returns_ErrInvalidUserID", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		useCase := NewGetUserUseCase(mockRepository)

		_, err := useCase.Execute(context.Background(), "not-a-uuid")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrInvalidUserID) {
			t.Fatalf("errors.Is: got %v want %v", err, ErrInvalidUserID)
		}
	})

	t.Run("not_found_is_propagated", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockRepository.EXPECT().FindByID(gomock.Any(), testUserID).Return(nil, ErrNotFound)
		useCase := NewGetUserUseCase(mockRepository)

		_, err := useCase.Execute(context.Background(), testUserID)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("errors.Is: got %v want %v", err, ErrNotFound)
		}
	})

	t.Run("repository_error_is_propagated", func(t *testing.T) {
		t.Parallel()
		repositoryError := errors.New("connection reset")
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockRepository.EXPECT().FindByID(gomock.Any(), testUserID).Return(nil, repositoryError)
		useCase := NewGetUserUseCase(mockRepository)

		_, err := useCase.Execute(context.Background(), testUserID)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, repositoryError) {
			t.Fatalf("errors.Is: got %v want %v", err, repositoryError)
		}
	})
}
