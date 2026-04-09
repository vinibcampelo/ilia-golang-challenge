package persistence

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	domaintransaction "ilia-golang-challenge/services/ms-transactions/internal/domain/transaction"
)

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewPostgresTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func advisoryLockKeys(uid uuid.UUID) (int32, int32) {
	b := uid
	k1 := int32(binary.BigEndian.Uint32(b[0:4]))
	k2 := int32(binary.BigEndian.Uint32(b[4:8]))
	return k1, k2
}

const balanceQuery = `
SELECT COALESCE(SUM(CASE type WHEN 'CREDIT' THEN amount WHEN 'DEBIT' THEN -amount END), 0)
FROM transactions WHERE user_id = $1::uuid`

func (r *PostgresTransactionRepository) Create(ctx context.Context, tr domaintransaction.Transaction) (*domaintransaction.Transaction, error) {
	uid := tr.UserID

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	k1, k2 := advisoryLockKeys(uid)
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1, $2)`, k1, k2); err != nil {
		return nil, fmt.Errorf("advisory lock: %w", err)
	}

	var balance int64
	if err := tx.QueryRowContext(ctx, balanceQuery, uid.String()).Scan(&balance); err != nil {
		return nil, fmt.Errorf("balance: %w", err)
	}

	if err := tr.RejectsDebitIfInsufficient(balance); err != nil {
		return nil, err
	}

	var id uuid.UUID
	var createdAt time.Time
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO transactions (user_id, type, amount) VALUES ($1::uuid, $2, $3)
		RETURNING id, created_at`,
		uid.String(), tr.Type.String(), tr.Amount.MinorUnits(),
	).Scan(&id, &createdAt); err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	out := tr
	out.ID = id
	out.CreatedAt = createdAt
	return &out, nil
}

func (r *PostgresTransactionRepository) ListByUser(ctx context.Context, userID string, filterType *domaintransaction.Type) ([]domaintransaction.Transaction, error) {
	query := `
		SELECT id, user_id, type, amount, created_at
		FROM transactions
		WHERE user_id = $1::uuid`
	args := []any{userID}
	if filterType != nil {
		query += ` AND type = $2`
		args = append(args, filterType.String())
	}
	query += ` ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var list []domaintransaction.Transaction
	for rows.Next() {
		var t domaintransaction.Transaction
		var typ string
		var amountRaw int64
		if err := rows.Scan(&t.ID, &t.UserID, &typ, &amountRaw, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		t.Type = domaintransaction.Type(typ)
		t.Amount = domaintransaction.RehydrateMinorAmount(amountRaw)
		list = append(list, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if list == nil {
		list = []domaintransaction.Transaction{}
	}
	return list, nil
}

func (r *PostgresTransactionRepository) GetBalance(ctx context.Context, userID string) (int64, error) {
	var balance int64
	err := r.db.QueryRowContext(ctx, balanceQuery, userID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("balance: %w", err)
	}
	return balance, nil
}
