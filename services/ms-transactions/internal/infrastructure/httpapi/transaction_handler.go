package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"ilia-golang-challenge/services/ms-transactions/internal/application/transaction/usecase"
)

type TransactionHandler struct {
	createUC  *usecase.CreateTransactionUseCase
	listUC    *usecase.ListTransactionsUseCase
	balanceUC *usecase.GetBalanceUseCase
}

func NewTransactionHandler(
	createUC *usecase.CreateTransactionUseCase,
	listUC *usecase.ListTransactionsUseCase,
	balanceUC *usecase.GetBalanceUseCase,
) *TransactionHandler {
	return &TransactionHandler{
		createUC:  createUC,
		listUC:    listUC,
		balanceUC: balanceUC,
	}
}

func (h *TransactionHandler) PostTransaction(w http.ResponseWriter, r *http.Request) {
	subject, ok := SubjectFromContext(r.Context())
	if !ok {
		writeUnauthorized(w)
		return
	}

	var body transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	view, err := h.createUC.Execute(r.Context(), usecase.CreateTransactionInput{
		SubjectUserID: subject,
		BodyUserID:    body.UserID,
		Type:          body.Type,
		Amount:        body.Amount,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidUserID):
			http.Error(w, usecase.ErrInvalidUserID.Error(), http.StatusBadRequest)
		case errors.Is(err, usecase.ErrForbiddenUser):
			http.Error(w, usecase.ErrForbiddenUser.Error(), http.StatusForbidden)
		case errors.Is(err, usecase.ErrInvalidTransactionType):
			http.Error(w, usecase.ErrInvalidTransactionType.Error(), http.StatusBadRequest)
		case errors.Is(err, usecase.ErrInvalidAmount):
			http.Error(w, usecase.ErrInvalidAmount.Error(), http.StatusBadRequest)
		case errors.Is(err, usecase.ErrInsufficientBalance):
			http.Error(w, usecase.ErrInsufficientBalance.Error(), http.StatusUnprocessableEntity)
		default:
			respondInternalServerError(w)
		}
		return
	}

	writeJSONResponse(w, http.StatusCreated, view, fmt.Sprintf("/transactions/%s", view.ID))
}

func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	subject, ok := SubjectFromContext(r.Context())
	if !ok {
		writeUnauthorized(w)
		return
	}

	typeQuery := r.URL.Query().Get("type")
	list, err := h.listUC.Execute(r.Context(), subject, typeQuery)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidTypeQuery) {
			http.Error(w, usecase.ErrInvalidTypeQuery.Error(), http.StatusBadRequest)
			return
		}
		respondInternalServerError(w)
		return
	}
	if list == nil {
		list = []usecase.TransactionView{}
	}
	writeJSONResponse(w, http.StatusOK, list, "")
}

func (h *TransactionHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	subject, ok := SubjectFromContext(r.Context())
	if !ok {
		writeUnauthorized(w)
		return
	}

	view, err := h.balanceUC.Execute(r.Context(), subject)
	if err != nil {
		respondInternalServerError(w)
		return
	}
	writeJSONResponse(w, http.StatusOK, view, "")
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
