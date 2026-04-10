package usecase

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	domainuser "ilia-golang-challenge/services/ms-users/internal/domain/user"
)

type AccessTokenIssuer interface {
	IssueAccessToken(subjectUserID string) (string, error)
}

type AuthenticateUserInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	AccessToken string   `json:"access_token"`
	User        UserView `json:"user"`
}

type AuthenticateUserUseCase struct {
	repository  domainuser.Repository
	tokenIssuer AccessTokenIssuer
}

func NewAuthenticateUserUseCase(repository domainuser.Repository, tokenIssuer AccessTokenIssuer) *AuthenticateUserUseCase {
	return &AuthenticateUserUseCase{repository: repository, tokenIssuer: tokenIssuer}
}

func (useCase *AuthenticateUserUseCase) Execute(ctx context.Context, input AuthenticateUserInput) (*AuthResult, error) {
	email, err := domainuser.NormalizeEmail(input.Email)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", ErrInvalidCredentials)
	}

	foundUser, err := useCase.repository.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("auth: %w", ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("auth: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(input.Password)); err != nil {
		return nil, fmt.Errorf("auth: %w", ErrInvalidCredentials)
	}

	accessToken, err := useCase.tokenIssuer.IssueAccessToken(foundUser.ID)
	if err != nil {
		return nil, fmt.Errorf("auth: issue token: %w", err)
	}

	return &AuthResult{
		AccessToken: accessToken,
		User: UserView{
			ID:        foundUser.ID,
			FirstName: foundUser.FirstName,
			LastName:  foundUser.LastName,
			Email:     foundUser.Email,
		},
	}, nil
}
