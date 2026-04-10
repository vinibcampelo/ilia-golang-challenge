package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"ilia-golang-challenge/services/ms-transactions/internal/application/transaction/usecase"
)

type InternalWalletHandler struct {
	balanceUC *usecase.GetBalanceUseCase
}

func NewInternalWalletHandler(balanceUC *usecase.GetBalanceUseCase) *InternalWalletHandler {
	return &InternalWalletHandler{balanceUC: balanceUC}
}

type internalBalanceResponse struct {
	Balance int64 `json:"balance"`
}

func (h *InternalWalletHandler) GetBalanceInternal(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if _, err := uuid.Parse(userID); err != nil {
		http.Error(w, usecase.ErrInvalidUserID.Error(), http.StatusBadRequest)
		return
	}
	view, err := h.balanceUC.Execute(r.Context(), userID)
	if err != nil {
		respondInternalServerError(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(internalBalanceResponse{Balance: view.Amount})
}
