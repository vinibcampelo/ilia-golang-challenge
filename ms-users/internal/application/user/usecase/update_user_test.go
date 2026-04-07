package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	domainuser "ilia-golang-challenge/ms-users/internal/domain/user"
	domainusermocks "ilia-golang-challenge/ms-users/internal/domain/user/mocks"
)

func newMockRepository(t *testing.T) *domainusermocks.MockRepository {
	t.Helper()

	controller := gomock.NewController(t)
	return domainusermocks.NewMockRepository(controller)
}

func TestUpdateUserUseCase_Execute_success(t *testing.T) {
	t.Parallel()

	t.Run("finds_user_applies_patch_persists_and_returns_view_without_password", func(t *testing.T) {
		t.Parallel()
		existingUser := &domainuser.User{
			ID:        "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
			FirstName: "user1",
			LastName:  "lastname1",
			Email:     "user1@test.com",
			Password:  "hash",
		}
		existingUser.FirstName = "oldFirstName"
		existingUser.Email = "old@example.com"
		existingUser.Password = "unchanged-hash"

		mockRepository := newMockRepository(t)
		mockRepository.EXPECT().FindByID(gomock.Any(), existingUser.ID).Return(existingUser, nil)

		var updatedUser *domainuser.User
		mockRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Do(func(_ context.Context, entity *domainuser.User) {
			copied := *entity
			updatedUser = &copied
		}).Return(nil)

		useCase := NewUpdateUserUseCase(mockRepository, bcrypt.MinCost)
		firstName := "  newFirstName  "
		email := "NEW@EXAMPLE.COM"
		userView, err := useCase.Execute(context.Background(), existingUser.ID, UpdateUserInput{
			FirstName: &firstName,
			Email:     &email,
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if userView.FirstName != "newFirstName" || userView.LastName != existingUser.LastName || userView.Email != "new@example.com" {
			t.Fatalf("unexpected view: %+v", userView)
		}
		if updatedUser == nil {
			t.Fatal("expected Update to be called")
		}
		if updatedUser.FirstName != "newFirstName" || updatedUser.Email != "new@example.com" {
			t.Fatalf("persisted: %+v", updatedUser)
		}
		if updatedUser.Password != "unchanged-hash" {
			t.Fatal("password should be unchanged when not in patch")
		}
	})

	t.Run("rehashes_password_when_password_field_is_set", func(t *testing.T) {
		t.Parallel()
		existingUser := &domainuser.User{
			ID:        "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
			FirstName: "user1",
			LastName:  "lastname1",
			Email:     "user1@test.com",
			Password:  "hash",
		}
		existingUser.Password = "old-hash"

		mockRepository := newMockRepository(t)
		mockRepository.EXPECT().FindByID(gomock.Any(), existingUser.ID).Return(existingUser, nil)

		var updatedUser *domainuser.User
		mockRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Do(func(_ context.Context, entity *domainuser.User) {
			copied := *entity
			updatedUser = &copied
		}).Return(nil)

		useCase := NewUpdateUserUseCase(mockRepository, bcrypt.MinCost)
		newPassword := "newpass12"
		_, err := useCase.Execute(context.Background(), existingUser.ID, UpdateUserInput{Password: &newPassword})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if updatedUser == nil {
			t.Fatal("expected Update")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(updatedUser.Password), []byte("newpass12")); err != nil {
			t.Fatalf("password not bcrypt of new value: %v", err)
		}
	})
}

func TestUpdateUserUseCase_Execute_errors(t *testing.T) {
	t.Parallel()

	t.Run("invalid_user_id_returns_ErrInvalidUserID", func(t *testing.T) {
		t.Parallel()
		mockRepository := newMockRepository(t)
		useCase := NewUpdateUserUseCase(mockRepository, bcrypt.MinCost)
		firstName := "user1"
		_, err := useCase.Execute(context.Background(), "nope", UpdateUserInput{FirstName: &firstName})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrInvalidUserID) {
			t.Fatalf("errors.Is: got %v want %v", err, ErrInvalidUserID)
		}
	})

	t.Run("no_update_fields_returns_ErrNoUpdateFields", func(t *testing.T) {
		t.Parallel()
		mockRepository := newMockRepository(t)
		useCase := NewUpdateUserUseCase(mockRepository, bcrypt.MinCost)
		_, err := useCase.Execute(context.Background(), "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", UpdateUserInput{})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrNoUpdateFields) {
			t.Fatalf("errors.Is: got %v want %v", err, ErrNoUpdateFields)
		}
	})

	t.Run("not_found_from_repository_is_propagated", func(t *testing.T) {
		t.Parallel()
		userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
		mockRepository := newMockRepository(t)
		mockRepository.EXPECT().FindByID(gomock.Any(), userID).Return(nil, ErrNotFound)
		useCase := NewUpdateUserUseCase(mockRepository, bcrypt.MinCost)
		firstName := "newFirstName"
		_, err := useCase.Execute(context.Background(), userID, UpdateUserInput{FirstName: &firstName})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("errors.Is: got %v want %v", err, ErrNotFound)
		}
	})

	t.Run("duplicate_email_from_repository_is_propagated", func(t *testing.T) {
		t.Parallel()
		mockRepository := newMockRepository(t)
		existingUser := &domainuser.User{
			ID:        "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
			FirstName: "user1",
			LastName:  "lastname1",
			Email:     "user1@test.com",
			Password:  "hash",
		}
		existingUser.Email = "old@example.com"
		mockRepository.EXPECT().FindByID(gomock.Any(), existingUser.ID).Return(existingUser, nil)
		mockRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(fmt.Errorf("update user: %w", ErrEmailAlreadyExists))
		useCase := NewUpdateUserUseCase(mockRepository, bcrypt.MinCost)
		email := "taken@example.com"
		_, err := useCase.Execute(context.Background(), existingUser.ID, UpdateUserInput{Email: &email})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrEmailAlreadyExists) {
			t.Fatalf("errors.Is: got %v want %v", err, ErrEmailAlreadyExists)
		}
	})
}
