package usecase

import "context"

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=wallet_zero_balance_gate.go -destination=mocks/wallet_zero_balance_gate_mock.go -package=mocks
type WalletZeroBalanceChecker interface {
	AssertZeroBalance(ctx context.Context, userID string) error
}
