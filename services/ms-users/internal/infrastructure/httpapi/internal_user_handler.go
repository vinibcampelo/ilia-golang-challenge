package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"ilia-golang-challenge/services/ms-users/internal/application/user/usecase"
)

type InternalUserHandler struct {
	getUserUseCase *usecase.GetUserUseCase
}

func NewInternalUserHandler(getUserUseCase *usecase.GetUserUseCase) *InternalUserHandler {
	return &InternalUserHandler{getUserUseCase: getUserUseCase}
}

type internalUserActiveResponse struct {
	ID     string `json:"id"`
	Active bool   `json:"active"`
}

func (h *InternalUserHandler) GetUserInternal(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if _, err := uuid.Parse(userID); err != nil {
		http.Error(w, usecase.ErrInvalidUserID.Error(), http.StatusBadRequest)
		return
	}
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(internalUserActiveResponse{ID: userView.ID, Active: true})
}
