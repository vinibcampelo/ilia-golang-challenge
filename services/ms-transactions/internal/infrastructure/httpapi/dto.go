package httpapi

type transactionRequest struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	Amount int64  `json:"amount"`
}
