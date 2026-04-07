package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	domainuser "ilia-golang-challenge/ms-users/internal/domain/user"
)

type UpdateUserInput struct {
	FirstName *string
	LastName  *string
	Email     *string
	Password  *string
}

type UpdateUserUseCase struct {
	repository domainuser.Repository
	bcryptCost int
}

func NewUpdateUserUseCase(repository domainuser.Repository, bcryptCost int) *UpdateUserUseCase {
	if bcryptCost == 0 {
		bcryptCost = bcrypt.DefaultCost
	}
	return &UpdateUserUseCase{repository: repository, bcryptCost: bcryptCost}
}

func (useCase *UpdateUserUseCase) Execute(ctx context.Context, userID string, input UpdateUserInput) (*UserView, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, fmt.Errorf("update user: %w", ErrInvalidUserID)
	}

	userUpdate := domainuser.UserUpdate{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  input.Password,
	}
	if userUpdate.IsEmpty() {
		return nil, fmt.Errorf("update user: %w", ErrNoUpdateFields)
	}

	foundUser, err := useCase.repository.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	if err := foundUser.Update(userUpdate); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	if input.Password != nil {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(foundUser.Password), useCase.bcryptCost)
		if err != nil {
			return nil, fmt.Errorf("update user: hash password: %w", err)
		}
		foundUser.Password = string(passwordHash)
	}

	if err := useCase.repository.Update(ctx, foundUser); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &UserView{
		ID:        foundUser.ID,
		FirstName: foundUser.FirstName,
		LastName:  foundUser.LastName,
		Email:     foundUser.Email,
	}, nil
}
