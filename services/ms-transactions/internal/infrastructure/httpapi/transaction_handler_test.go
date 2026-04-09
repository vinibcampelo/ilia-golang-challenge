package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"ilia-golang-challenge/services/ms-transactions/internal/application/transaction/usecase"
	domaintransaction "ilia-golang-challenge/services/ms-transactions/internal/domain/transaction"
	transactionmocks "ilia-golang-challenge/services/ms-transactions/internal/domain/transaction/mocks"
)

const (
	transactionHandlerTestUserID    = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	transactionHandlerTestOtherID   = "7c9e6679-7425-40de-944b-e07fc1f90ae7"
	transactionHandlerTestJWTSecret = "transaction-handler-test-jwt-secret-min-length!!"
)

func newTransactionTestRouter(tb testing.TB, repository domaintransaction.Repository) http.Handler {
	tb.Helper()
	handler := NewTransactionHandler(
		usecase.NewCreateTransactionUseCase(repository),
		usecase.NewListTransactionsUseCase(repository),
		usecase.NewGetBalanceUseCase(repository),
	)
	return NewRouter(handler, []byte(transactionHandlerTestJWTSecret), nil)
}

func transactionTestJWT(tb testing.TB, subject string) string {
	tb.Helper()
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(transactionHandlerTestJWTSecret))
	if err != nil {
		tb.Fatal(err)
	}
	return s
}

func mustMinorAmount(tb testing.TB, v int64) domaintransaction.MinorAmount {
	tb.Helper()
	m, err := domaintransaction.NewMinorAmount(v)
	if err != nil {
		tb.Fatal(err)
	}
	return m
}

func TestTransactionHandler_PostTransaction_success(tb *testing.T) {
	tb.Parallel()

	tb.Run("returns_201_with_json_and_location_after_credit", func(t *testing.T) {
		t.Parallel()
		subject := transactionHandlerTestUserID
		uid := uuid.MustParse(subject)
		pending := domaintransaction.NewTransaction(uid, domaintransaction.TypeCredit, mustMinorAmount(t, 25))

		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		mockRepository.EXPECT().Create(gomock.Any(), domaintransaction.RepositoryCreateInput{Transaction: pending}).
			Return(&domaintransaction.Transaction{
				ID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				UserID: uid,
				Type:   domaintransaction.TypeCredit,
				Amount: mustMinorAmount(t, 25),
			}, nil)

		srv := newTransactionTestRouter(t, mockRepository)
		body := `{"user_id":"` + subject + `","type":"CREDIT","amount":25}`
		request := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+transactionTestJWT(t, subject))
		responseRecorder := httptest.NewRecorder()
		srv.ServeHTTP(responseRecorder, request)

		if responseRecorder.Code != http.StatusCreated {
			t.Fatalf("status: got %d want %d, body: %s", responseRecorder.Code, http.StatusCreated, responseRecorder.Body.String())
		}
		if contentType := responseRecorder.Header().Get("Content-Type"); contentType != "application/json" {
			t.Fatalf("Content-Type: got %q want application/json", contentType)
		}
		location := responseRecorder.Header().Get("Location")
		if !strings.HasPrefix(location, "/transactions/") {
			t.Fatalf("Location: got %q want prefix /transactions/", location)
		}

		var parsed usecase.TransactionView
		if err := json.NewDecoder(responseRecorder.Body).Decode(&parsed); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if parsed.Amount != 25 || parsed.Type != "CREDIT" || parsed.UserID != subject {
			t.Fatalf("unexpected body: %+v", parsed)
		}
	})

	tb.Run("idempotency_replay_returns_same_201_location_and_body", func(t *testing.T) {
		t.Parallel()
		subject := transactionHandlerTestUserID
		uid := uuid.MustParse(subject)
		pending := domaintransaction.NewTransaction(uid, domaintransaction.TypeCredit, mustMinorAmount(t, 25))
		repoIn := domaintransaction.RepositoryCreateInput{
			Transaction:        pending,
			IdempotencyKey:     "replay-key",
			RequestFingerprint: usecase.TransactionRequestFingerprint(subject, domaintransaction.TypeCredit, 25),
		}
		created := &domaintransaction.Transaction{
			ID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			UserID: uid,
			Type:   domaintransaction.TypeCredit,
			Amount: mustMinorAmount(t, 25),
		}

		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		gomock.InOrder(
			mockRepository.EXPECT().Create(gomock.Any(), repoIn).Return(created, nil),
			mockRepository.EXPECT().Create(gomock.Any(), repoIn).Return(created, nil),
		)

		srv := newTransactionTestRouter(t, mockRepository)
		body := `{"user_id":"` + subject + `","type":"CREDIT","amount":25}`
		post := func() *httptest.ResponseRecorder {
			t.Helper()
			request := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Authorization", "Bearer "+transactionTestJWT(t, subject))
			request.Header.Set("Idempotency-Key", "replay-key")
			recorder := httptest.NewRecorder()
			srv.ServeHTTP(recorder, request)
			return recorder
		}

		first := post()
		second := post()
		if first.Code != http.StatusCreated || second.Code != http.StatusCreated {
			t.Fatalf("status: first %d second %d", first.Code, second.Code)
		}
		if first.Header().Get("Location") != second.Header().Get("Location") {
			t.Fatalf("Location: first %q second %q", first.Header().Get("Location"), second.Header().Get("Location"))
		}
		if !bytes.Equal(first.Body.Bytes(), second.Body.Bytes()) {
			t.Fatalf("body mismatch: first %s second %s", first.Body.String(), second.Body.String())
		}
	})
}

func TestTransactionHandler_PostTransaction_errors(tb *testing.T) {
	tb.Parallel()
	subject := transactionHandlerTestUserID
	uid := uuid.MustParse(subject)

	tb.Run("missing_bearer_returns_401", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		srv := newTransactionTestRouter(t, mockRepository)
		request := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(`{"user_id":"`+subject+`","type":"CREDIT","amount":1}`))
		request.Header.Set("Content-Type", "application/json")
		responseRecorder := httptest.NewRecorder()
		srv.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusUnauthorized {
			t.Fatalf("status: got %d want %d", responseRecorder.Code, http.StatusUnauthorized)
		}
	})

	tb.Run("user_id_mismatch_returns_403", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		srv := newTransactionTestRouter(t, mockRepository)
		body := `{"user_id":"` + transactionHandlerTestOtherID + `","type":"CREDIT","amount":1}`
		request := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+transactionTestJWT(t, subject))
		responseRecorder := httptest.NewRecorder()
		srv.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusForbidden {
			t.Fatalf("status: got %d want %d, body: %s", responseRecorder.Code, http.StatusForbidden, responseRecorder.Body.String())
		}
	})

	tb.Run("insufficient_balance_returns_422", func(t *testing.T) {
		t.Parallel()
		debitPending := domaintransaction.NewTransaction(uid, domaintransaction.TypeDebit, mustMinorAmount(t, 999))
		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		mockRepository.EXPECT().Create(gomock.Any(), domaintransaction.RepositoryCreateInput{Transaction: debitPending}).
			Return(nil, domaintransaction.ErrInsufficientBalance)

		srv := newTransactionTestRouter(t, mockRepository)
		body := `{"user_id":"` + subject + `","type":"DEBIT","amount":999}`
		request := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+transactionTestJWT(t, subject))
		responseRecorder := httptest.NewRecorder()
		srv.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status: got %d want %d, body: %s", responseRecorder.Code, http.StatusUnprocessableEntity, responseRecorder.Body.String())
		}
	})

	tb.Run("invalid_json_returns_400", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		srv := newTransactionTestRouter(t, mockRepository)
		request := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(`{`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+transactionTestJWT(t, subject))
		responseRecorder := httptest.NewRecorder()
		srv.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d want %d", responseRecorder.Code, http.StatusBadRequest)
		}
	})

	tb.Run("idempotency_key_exceeds_max_length_returns_400", func(t *testing.T) {
		t.Parallel()
		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		srv := newTransactionTestRouter(t, mockRepository)
		longKey := strings.Repeat("a", maxIdempotencyKeyRunes+1)
		body := `{"user_id":"` + subject + `","type":"CREDIT","amount":1}`
		request := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+transactionTestJWT(t, subject))
		request.Header.Set("Idempotency-Key", longKey)
		responseRecorder := httptest.NewRecorder()
		srv.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d want %d, body: %s", responseRecorder.Code, http.StatusBadRequest, responseRecorder.Body.String())
		}
	})

	tb.Run("idempotency_conflict_returns_409", func(t *testing.T) {
		t.Parallel()
		pending := domaintransaction.NewTransaction(uid, domaintransaction.TypeCredit, mustMinorAmount(t, 1))
		repoIn := domaintransaction.RepositoryCreateInput{
			Transaction:        pending,
			IdempotencyKey:     "conflict-key",
			RequestFingerprint: usecase.TransactionRequestFingerprint(subject, domaintransaction.TypeCredit, 1),
		}
		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		mockRepository.EXPECT().Create(gomock.Any(), repoIn).
			Return(nil, domaintransaction.ErrIdempotencyConflict)

		srv := newTransactionTestRouter(t, mockRepository)
		body := `{"user_id":"` + subject + `","type":"CREDIT","amount":1}`
		request := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+transactionTestJWT(t, subject))
		request.Header.Set("Idempotency-Key", "conflict-key")
		responseRecorder := httptest.NewRecorder()
		srv.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusConflict {
			t.Fatalf("status: got %d want %d, body: %s", responseRecorder.Code, http.StatusConflict, responseRecorder.Body.String())
		}
		responseBody := strings.TrimSpace(responseRecorder.Body.String())
		wantBody := usecase.ErrIdempotencyConflict.Error()
		if responseBody != wantBody {
			t.Fatalf("body: got %q want %q", responseBody, wantBody)
		}
	})
}

func TestTransactionHandler_GetTransactions_success(tb *testing.T) {
	tb.Parallel()

	tb.Run("returns_200_empty_json_array_when_no_rows", func(t *testing.T) {
		t.Parallel()
		subject := transactionHandlerTestUserID
		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		mockRepository.EXPECT().ListByUser(gomock.Any(), subject, gomock.Nil()).
			Return([]domaintransaction.Transaction{}, nil)

		srv := newTransactionTestRouter(t, mockRepository)
		request := httptest.NewRequest(http.MethodGet, "/transactions", nil)
		request.Header.Set("Authorization", "Bearer "+transactionTestJWT(t, subject))
		responseRecorder := httptest.NewRecorder()
		srv.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusOK {
			t.Fatalf("status: got %d want %d", responseRecorder.Code, http.StatusOK)
		}
		var parsed []usecase.TransactionView
		if err := json.NewDecoder(responseRecorder.Body).Decode(&parsed); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if parsed == nil || len(parsed) != 0 {
			t.Fatalf("want empty non-nil JSON array, got %#v", parsed)
		}
	})
}

func TestTransactionHandler_GetTransactions_errors(tb *testing.T) {
	tb.Parallel()

	tb.Run("invalid_type_query_returns_400", func(t *testing.T) {
		t.Parallel()
		subject := transactionHandlerTestUserID
		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		srv := newTransactionTestRouter(t, mockRepository)
		request := httptest.NewRequest(http.MethodGet, "/transactions?type=FOO", nil)
		request.Header.Set("Authorization", "Bearer "+transactionTestJWT(t, subject))
		responseRecorder := httptest.NewRecorder()
		srv.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d want %d", responseRecorder.Code, http.StatusBadRequest)
		}
	})
}

func TestTransactionHandler_GetBalance_success(tb *testing.T) {
	tb.Parallel()

	tb.Run("returns_200_json_balance", func(t *testing.T) {
		t.Parallel()
		subject := transactionHandlerTestUserID
		controller := gomock.NewController(t)
		mockRepository := transactionmocks.NewMockRepository(controller)
		mockRepository.EXPECT().GetBalance(gomock.Any(), subject).Return(int64(42), nil)

		srv := newTransactionTestRouter(t, mockRepository)
		request := httptest.NewRequest(http.MethodGet, "/balance", nil)
		request.Header.Set("Authorization", "Bearer "+transactionTestJWT(t, subject))
		responseRecorder := httptest.NewRecorder()
		srv.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusOK {
			t.Fatalf("status: got %d want %d", responseRecorder.Code, http.StatusOK)
		}
		var parsed usecase.BalanceView
		if err := json.NewDecoder(responseRecorder.Body).Decode(&parsed); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if parsed.Amount != 42 {
			t.Fatalf("Amount: got %d want 42", parsed.Amount)
		}
	})
}
