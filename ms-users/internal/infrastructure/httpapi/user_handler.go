package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"ilia-golang-challenge/ms-users/internal/application/user/usecase"
	domainuser "ilia-golang-challenge/ms-users/internal/domain/user"
)

type UserHandler struct {
	createUser *usecase.CreateUserUseCase
}

func NewUserHandler(createUser *usecase.CreateUserUseCase) *UserHandler {
	return &UserHandler{createUser: createUser}
}

func (h *UserHandler) PostUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var requestBody userRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	out, err := h.createUser.Execute(r.Context(), usecase.CreateUserInput{
		FirstName: requestBody.FirstName,
		LastName:  requestBody.LastName,
		Email:     requestBody.Email,
		Password:  requestBody.Password,
	})

	if err != nil {
		if isUserRegistrationValidationError(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, domainuser.ErrEmailAlreadyExists) {
			http.Error(w, domainuser.ErrEmailAlreadyExists.Error(), http.StatusConflict)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("/users/%s", out.ID))
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(usersResponse{
		ID:        out.ID,
		FirstName: out.FirstName,
		LastName:  out.LastName,
		Email:     out.Email,
	})
}

func isUserRegistrationValidationError(err error) bool {
	return errors.Is(err, domainuser.ErrInvalidFirstName) ||
		errors.Is(err, domainuser.ErrInvalidLastName) ||
		errors.Is(err, domainuser.ErrInvalidEmail) ||
		errors.Is(err, domainuser.ErrInvalidPassword) ||
		errors.Is(err, domainuser.ErrPasswordTooShort)
}
