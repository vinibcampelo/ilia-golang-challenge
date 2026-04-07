package usecase

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	domainuser "ilia-golang-challenge/ms-users/internal/domain/user"
	domainusermocks "ilia-golang-challenge/ms-users/internal/domain/user/mocks"
)

const (
	testUserID1    = "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	testUserID2    = "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12"
	testFirstName1 = "user1"
	testLastName1  = "lastname1"
	testEmail1     = "user1@test.com"
	testFirstName2 = "user2"
	testLastName2  = "lastname2"
	testEmail2     = "user2@test.com"
)

func TestListUsersUseCase_Execute_success(t *testing.T) {
	t.Parallel()

	t.Run("maps_users_to_views_and_never_leaks_password", func(t *testing.T) {
		t.Parallel()
		repositoryUsers := []domainuser.User{
			{ID: testUserID1, FirstName: testFirstName1, LastName: testLastName1, Email: testEmail1},
			{ID: testUserID2, FirstName: testFirstName2, LastName: testLastName2, Email: testEmail2, Password: "hash-should-not-leak"},
		}
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockRepository.EXPECT().List(gomock.Any()).Return(repositoryUsers, nil)
		useCase := NewListUsersUseCase(mockRepository)

		userViews, err := useCase.Execute(context.Background())
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if len(userViews) != 2 {
			t.Fatalf("len: got %d want 2", len(userViews))
		}
		if userViews[0].ID != repositoryUsers[0].ID || userViews[0].FirstName != testFirstName1 || userViews[0].Email != testEmail1 {
			t.Fatalf("first view: %+v", userViews[0])
		}
		if userViews[1].FirstName != testFirstName2 || userViews[1].Email != testEmail2 {
			t.Fatalf("second view: %+v", userViews[1])
		}
	})

	t.Run("empty_repository_returns_empty_non_nil_slice", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockRepository.EXPECT().List(gomock.Any()).Return(nil, nil)
		useCase := NewListUsersUseCase(mockRepository)

		userViews, err := useCase.Execute(context.Background())
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if userViews == nil || len(userViews) != 0 {
			t.Fatalf("want empty non-nil slice, got %#v", userViews)
		}
	})
}

func TestListUsersUseCase_Execute_errors(t *testing.T) {
	t.Parallel()

	t.Run("repository_error_is_propagated", func(t *testing.T) {
		t.Parallel()
		repositoryError := errors.New("db down")
		controller := gomock.NewController(t)
		mockRepository := domainusermocks.NewMockRepository(controller)
		mockRepository.EXPECT().List(gomock.Any()).Return(nil, repositoryError)
		useCase := NewListUsersUseCase(mockRepository)

		_, err := useCase.Execute(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, repositoryError) {
			t.Fatalf("errors.Is: got %v want %v", err, repositoryError)
		}
	})
}
