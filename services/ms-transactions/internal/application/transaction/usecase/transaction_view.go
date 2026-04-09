package usecase

type TransactionView struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	Amount int64  `json:"amount"`
}

type BalanceView struct {
	Amount int64 `json:"amount"`
}
