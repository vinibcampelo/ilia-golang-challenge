package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	domainuser "ilia-golang-challenge/ms-users/internal/domain/user"
)

type CreateUserInput struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
}

type CreateUserOutput struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
}

type CreateUserUseCase struct {
	repo       domainuser.Repository
	bcryptCost int
}

func NewCreateUserUseCase(repo domainuser.Repository, bcryptCost int) *CreateUserUseCase {
	if bcryptCost == 0 {
		bcryptCost = bcrypt.DefaultCost
	}
	return &CreateUserUseCase{repo: repo, bcryptCost: bcryptCost}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
	id := uuid.NewString()
	user, err := domainuser.NewUser(id, input.FirstName, input.LastName, input.Email, input.Password)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), uc.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user.Password = string(hash)
	if err := uc.repo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("persist user: %w", err)
	}
	return &CreateUserOutput{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}, nil
}
