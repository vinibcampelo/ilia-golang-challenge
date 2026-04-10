package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"ilia-golang-challenge/services/ms-users/internal/application/user/usecase"
	domainuser "ilia-golang-challenge/services/ms-users/internal/domain/user"
)

type UserHandler struct {
	createUserUseCase *usecase.CreateUserUseCase
	listUsersUseCase  *usecase.ListUsersUseCase
	getUserUseCase    *usecase.GetUserUseCase
	updateUserUseCase *usecase.UpdateUserUseCase
	deleteUserUseCase *usecase.DeleteUserUseCase
}

func NewUserHandler(
	createUserUseCase *usecase.CreateUserUseCase,
	listUsersUseCase *usecase.ListUsersUseCase,
	getUserUseCase *usecase.GetUserUseCase,
	updateUserUseCase *usecase.UpdateUserUseCase,
	deleteUserUseCase *usecase.DeleteUserUseCase,
) *UserHandler {
	return &UserHandler{
		createUserUseCase: createUserUseCase,
		listUsersUseCase:  listUsersUseCase,
		getUserUseCase:    getUserUseCase,
		updateUserUseCase: updateUserUseCase,
		deleteUserUseCase: deleteUserUseCase,
	}
}

func (h *UserHandler) PostUser(w http.ResponseWriter, r *http.Request) {
	var registrationRequest userRequest
	if err := json.NewDecoder(r.Body).Decode(&registrationRequest); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	userView, err := h.createUserUseCase.Execute(r.Context(), usecase.CreateUserInput{
		FirstName: registrationRequest.FirstName,
		LastName:  registrationRequest.LastName,
		Email:     registrationRequest.Email,
		Password:  registrationRequest.Password,
	})
	if err != nil {
		if isUserRegistrationValidationError(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, usecase.ErrEmailAlreadyExists) {
			http.Error(w, usecase.ErrEmailAlreadyExists.Error(), http.StatusConflict)
			return
		}
		respondInternalServerError(w)
		return
	}

	writeJSONResponse(w, http.StatusCreated, userView, fmt.Sprintf("/users/%s", userView.ID))
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.listUsersUseCase.Execute(r.Context())
	if err != nil {
		respondInternalServerError(w)
		return
	}
	if users == nil {
		users = []usecase.UserView{}
	}
	writeJSONResponse(w, http.StatusOK, users, "")
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	userView, err := h.getUserUseCase.Execute(r.Context(), userID)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidUserID) {
			http.Error(w, usecase.ErrInvalidUserID.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, usecase.ErrNotFound) {
			http.Error(w, usecase.ErrNotFound.Error(), http.StatusNotFound)
			return
		}
		respondInternalServerError(w)
		return
	}
	writeJSONResponse(w, http.StatusOK, userView, "")
}

func (h *UserHandler) PatchUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	var patchRequest userPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&patchRequest); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	userView, err := h.updateUserUseCase.Execute(r.Context(), userID, usecase.UpdateUserInput{
		FirstName: patchRequest.FirstName,
		LastName:  patchRequest.LastName,
		Email:     patchRequest.Email,
		Password:  patchRequest.Password,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidUserID) {
			http.Error(w, usecase.ErrInvalidUserID.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, usecase.ErrNoUpdateFields) {
			http.Error(w, usecase.ErrNoUpdateFields.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, usecase.ErrNotFound) {
			http.Error(w, usecase.ErrNotFound.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, usecase.ErrEmailAlreadyExists) {
			http.Error(w, usecase.ErrEmailAlreadyExists.Error(), http.StatusConflict)
			return
		}
		if isUserRegistrationValidationError(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		respondInternalServerError(w)
		return
	}

	writeJSONResponse(w, http.StatusOK, userView, "")
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if err := h.deleteUserUseCase.Execute(r.Context(), userID); err != nil {
		if errors.Is(err, usecase.ErrInvalidUserID) {
			http.Error(w, usecase.ErrInvalidUserID.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, usecase.ErrNotFound) {
			http.Error(w, usecase.ErrNotFound.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, usecase.ErrWalletHasNonZeroBalance) {
			http.Error(w, usecase.ErrWalletHasNonZeroBalance.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, usecase.ErrWalletServiceUnavailable) {
			http.Error(w, usecase.ErrWalletServiceUnavailable.Error(), http.StatusServiceUnavailable)
			return
		}
		respondInternalServerError(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, payload any, location string) {
	w.Header().Set("Content-Type", "application/json")
	if location != "" {
		w.Header().Set("Location", location)
	}
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondInternalServerError(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func isUserRegistrationValidationError(err error) bool {
	return errors.Is(err, domainuser.ErrInvalidFirstName) ||
		errors.Is(err, domainuser.ErrInvalidLastName) ||
		errors.Is(err, domainuser.ErrInvalidEmail) ||
		errors.Is(err, domainuser.ErrInvalidPassword) ||
		errors.Is(err, domainuser.ErrPasswordTooShort)
}
