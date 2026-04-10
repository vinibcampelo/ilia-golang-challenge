package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	domainuser "ilia-golang-challenge/services/ms-users/internal/domain/user"
)

type CreateUserInput struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
}

type CreateUserOutput = UserView

type CreateUserUseCase struct {
	repository domainuser.Repository
	bcryptCost int
}

func NewCreateUserUseCase(repository domainuser.Repository, bcryptCost int) *CreateUserUseCase {
	if bcryptCost == 0 {
		bcryptCost = bcrypt.DefaultCost
	}
	return &CreateUserUseCase{repository: repository, bcryptCost: bcryptCost}
}

func (useCase *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
	userID := uuid.NewString()
	password := input.Password

	createdUser, err := domainuser.New(
		userID,
		input.FirstName,
		input.LastName,
		input.Email,
		password,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), useCase.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	createdUser.Password = string(passwordHash)

	if err := useCase.repository.Save(ctx, createdUser); err != nil {
		return nil, fmt.Errorf("persist user: %w", err)
	}
	return &UserView{
		ID:        createdUser.ID,
		FirstName: createdUser.FirstName,
		LastName:  createdUser.LastName,
		Email:     createdUser.Email,
	}, nil
}
