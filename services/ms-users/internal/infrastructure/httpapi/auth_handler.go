package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"ilia-golang-challenge/services/ms-users/internal/application/user/usecase"
)

type AuthHandler struct {
	authenticateUserUseCase *usecase.AuthenticateUserUseCase
}

func NewAuthHandler(authenticateUserUseCase *usecase.AuthenticateUserUseCase) *AuthHandler {
	return &AuthHandler{authenticateUserUseCase: authenticateUserUseCase}
}

type authLoginRequest struct {
	User struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	} `json:"user"`
}

func (h *AuthHandler) PostAuth(w http.ResponseWriter, r *http.Request) {
	var loginBody authLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginBody); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	authResult, err := h.authenticateUserUseCase.Execute(r.Context(), usecase.AuthenticateUserInput{
		Email:    loginBody.User.Email,
		Password: loginBody.User.Password,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			http.Error(w, usecase.ErrInvalidCredentials.Error(), http.StatusUnauthorized)
			return
		}
		respondInternalServerError(w)
		return
	}

	writeJSONResponse(w, http.StatusOK, authResult, "")
}
