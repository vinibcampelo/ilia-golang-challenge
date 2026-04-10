package usecase

import "context"

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=user_active_gate.go -destination=mocks/user_active_gate_mock.go -package=mocks
type UserActiveGate interface {
	EnsureActive(ctx context.Context, userID string) error
}
